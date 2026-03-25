// Package generate provides model generation from database schema.
package generate

import (
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
	"github.com/sishui/bake/internal/schema"
)

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
	tmpl, err := parseTemplates(c.Template)
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
	initialisms := c.Naming()
	totalTables := 0

	for _, db := range c.DB {
		slog.DebugContext(ctx, "processing database", "driver", db.Driver, "schema", db.Schema)
		count, err := generate(ctx, db, c, tmpl, initialisms)
		if err != nil {
			return 0, err
		}
		totalTables += count
	}
	return totalTables, nil
}

func generate(ctx context.Context, db *config.DB, c *config.Config, tmpl *templates, initialisms map[string]string) (int, error) {
	tables, err := loadSchema(ctx, db)
	if err != nil {
		return 0, err
	}

	results := processTablesConcurrently(ctx, tables, db, c, tmpl, initialisms)

	errs := collectErrors(results)
	if len(errs) > 0 {
		return 0, errors.Join(errs...)
	}
	return len(tables), nil
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
	model    *Model
	filename string
	err      error
}

func processTablesConcurrently(ctx context.Context, tables []*schema.Table, db *config.DB, c *config.Config, tmpl *templates, initialisms map[string]string) <-chan result {
	results := make(chan result, len(tables))
	var wg sync.WaitGroup
	limiter := semaphore.NewWeighted(int64(runtime.NumCPU() * 4))

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
			m, err := NewModel(table, db, c, initialisms)
			if err != nil {
				select {
				case <-ctx.Done():
				case results <- result{err: err}:
				}
				return
			}
			filename, err := tmpl.writeTo(ctx, c.Template.Model, c.Output.Dir, m.Table, m)
			if err != nil {
				select {
				case <-ctx.Done():
				case results <- result{err: err}:
				}
				return
			}
			slog.DebugContext(ctx, "generated model", "table", m.Table, "model", m.Model, "file", filename)
			select {
			case <-ctx.Done():
			case results <- result{model: m, filename: filename}:
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
