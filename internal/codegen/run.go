// Package generate provides model generation from database schema.
package generate

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/logger"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
)

const (
	customStructTemplate = "custom"
)

var modelTemplates = map[string]string{
	"model":       "",
	"model.alias": "alias",
}

func Run(c *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Init(c.Log)

	start := time.Now()
	slog.Info("bake started")

	slog.DebugContext(ctx, "creating output directory", "dir", c.Output.Dir)
	err := os.MkdirAll(c.Output.Dir, 0o755)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "parsing templates", "dir", c.Template.Dir)
	tmpl, err := parseTemplates(c.Template.Dir)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "cleaning old generated files", "dir", c.Output.Dir)
	err = cleanDir(c.Output.Dir)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "loading schema and generating models")
	tableCount, err := run(ctx, c, tmpl)
	if err != nil {
		return err
	}

	slog.Info("bake completed", "duration", time.Since(start).String(), "tables", tableCount)
	return nil
}

func run(ctx context.Context, c *config.Config, tmpl *templates) (int, error) {
	nm := naming.New(c.Initialisms...)
	totalTables := 0

	for _, db := range c.DB {
		slog.DebugContext(ctx, "processing database", "driver", db.Driver, "schema", db.Schema)
		count, err := generate(ctx, db, c, tmpl, nm)
		if err != nil {
			return 0, err
		}
		totalTables += count
	}

	if len(c.Custom) > 0 {
		slog.DebugContext(ctx, "generating custom struct", "count", len(c.Custom))
		if err := generateCustomStruct(ctx, c, tmpl); err != nil {
			return 0, err
		}
	}

	return totalTables, nil
}

func generate(ctx context.Context, db *config.DB, c *config.Config, tmpl *templates, nm *naming.Naming) (int, error) {
	tables, err := loadSchema(ctx, db)
	if err != nil {
		return 0, err
	}

	results := processTablesConcurrently(ctx, tables, db, c, tmpl, nm)

	errs := collectErrors(results)
	if len(errs) > 0 {
		return 0, errors.Join(errs...)
	}
	return len(tables), nil
}

func generateCustomStruct(ctx context.Context, c *config.Config, tmpl *templates) error {
	for _, cs := range c.Custom {
		data := NewCustomStruct(c, cs)
		buffer, err := tmpl.render(customStructTemplate, data)
		if err != nil {
			return err
		}
		filename := naming.ToSnakeCase(cs.Name)
		_, err = writeFile(ctx, c.Output.Dir, filename, buffer)
		if err != nil {
			return err
		}
		slog.DebugContext(ctx, "generated custom struct", "struct", cs.Name, "file", filename+".gen.go")
	}
	return nil
}

func loadSchema(ctx context.Context, db *config.DB) ([]*schema.Table, error) {
	s, err := schema.New(db)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := s.Close(); err != nil {
			slog.ErrorContext(ctx, "close db", "error", err)
		}
	}()

	slog.DebugContext(ctx, "loading schema", "driver", db.Driver, "schema", db.Schema)
	tables, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	slog.DebugContext(ctx, "loaded tables", "count", len(tables))

	return tables, nil
}

type result struct {
	model     *Model
	filenames []string
	err       error
}

func processTable(ctx context.Context, table *schema.Table, db *config.DB, c *config.Config, tmpl *templates, nm *naming.Naming) result {
	m, err := NewModel(table, db, c, nm)
	if err != nil {
		return result{err: err}
	}
	filenames := make([]string, 0, len(modelTemplates))
	for template, suffix := range modelTemplates {
		buf, err := tmpl.render(template, m)
		if err != nil {
			return result{err: err}
		}
		filename := m.Table
		if len(suffix) > 0 {
			filename = m.Table + "." + suffix
		}
		filename, err = writeFile(ctx, c.Output.Dir, filename, buf)
		if err != nil {
			return result{err: err}
		}
		filenames = append(filenames, filename)
	}
	slog.DebugContext(ctx, "generated model", "table", m.Table, "model", m.Model, "files", filenames)
	return result{model: m, filenames: filenames}
}

func sendResult(ctx context.Context, ch chan<- result, r result) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- r:
		return true
	}
}

func processTablesConcurrently(ctx context.Context, tables []*schema.Table, db *config.DB, c *config.Config, tmpl *templates, nm *naming.Naming) <-chan result {
	results := make(chan result, len(tables))
	var wg sync.WaitGroup
	limiter := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, table := range tables {
		if err := limiter.Acquire(ctx, 1); err != nil {
			if !errors.Is(err, context.Canceled) {
				results <- result{err: err}
			}
			break
		}
		wg.Add(1)
		go func(table *schema.Table) {
			defer wg.Done()
			defer limiter.Release(1)

			select {
			case <-ctx.Done():
				return
			default:
			}

			slog.DebugContext(ctx, "processing table", "name", table.Name, "columns", len(table.Columns))
			r := processTable(ctx, table, db, c, tmpl, nm)
			if !sendResult(ctx, results, r) {
				return
			}
		}(table)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func collectErrors(results <-chan result) []error {
	var errs []error
	for r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}
	return errs
}

func cleanDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "gen.go") || strings.HasSuffix(path, "gen_test.go") {
			return os.Remove(path)
		}
		return nil
	})
}

func writeFile(ctx context.Context, dir string, filename string, data *bytes.Buffer) (string, error) {
	fullPath := filepath.Join(dir, filename+".gen.go")
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	_, err = data.WriteTo(file)
	closeErr := file.Close()
	if err != nil {
		if removeErr := os.Remove(fullPath); removeErr != nil {
			slog.ErrorContext(ctx, "remove file", "file", fullPath, "error", removeErr)
		}
		return "", err
	}
	if closeErr != nil {
		slog.ErrorContext(ctx, "close file", "file", fullPath, "error", closeErr)
	}
	return fullPath, nil
}
