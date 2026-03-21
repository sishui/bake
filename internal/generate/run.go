// Package generate provides model generation from database schema.
package generate

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/logger"
	"github.com/sishui/bake/internal/schema"
)

func Run(c *config.Config) error {
	logger.Init(c.Log)

	start := time.Now()
	slog.Info("bake started")

	slog.Debug("step: creating output directory", "dir", c.Output.Dir)
	err := os.MkdirAll(c.Output.Dir, 0o755)
	if err != nil {
		return err
	}

	ctx := context.Background()

	slog.Debug("step: parsing templates", "dir", c.Template.Dir)
	tmpl, err := parseTemplates(c.Template)
	if err != nil {
		return err
	}

	slog.Debug("step: cleaning old generated files", "dir", c.Output.Dir)
	err = cleanDir(c.Output.Dir)
	if err != nil {
		return err
	}

	slog.Debug("step: loading schema and generating models")
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
		slog.Debug("processing database", "driver", db.Driver, "schema", db.Schema)
		count, err := generate(ctx, db, c, tmpl, initialisms)
		if err != nil {
			return 0, err
		}
		totalTables += count
	}
	return totalTables, nil
}

func generate(ctx context.Context, db *config.DB, c *config.Config, tmpl *templates, initialisms map[string]string) (int, error) {
	s, err := schema.New(db)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := s.Close(); err != nil {
			slog.ErrorContext(ctx, "close db", "error", err)
		}
	}()
	slog.Debug("loading schema", "driver", db.Driver, "schema", db.Schema)
	tables, err := s.Load(ctx)
	if err != nil {
		return 0, err
	}
	slog.Debug("loaded tables", "count", len(tables))

	type result struct {
		model    *Model
		filename string
		err      error
	}

	results := make(chan result, len(tables))
	var wg sync.WaitGroup

	for _, table := range tables {
		wg.Add(1)
		go func(table *schema.Table) {
			defer wg.Done()
			slog.Debug("processing table", "name", table.Name, "columns", len(table.Columns))
			m, err := NewModel(table, db, c, initialisms)
			if err != nil {
				results <- result{err: err}
				return
			}
			filename, err := tmpl.writeTo(ctx, c.Template.Model, c.Output.Dir, m.Table, m)
			if err != nil {
				results <- result{err: err}
				return
			}
			slog.Debug("generated model", "table", m.Table, "model", m.Model, "file", filename)
			results <- result{model: m, filename: filename}
		}(table)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			return 0, r.err
		}
	}

	return len(tables), nil
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
			return os.RemoveAll(path)
		}
		return nil
	})
}
