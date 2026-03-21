// Package generate model
package generate

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/logger"
	"github.com/sishui/bake/internal/schema"
)

func Run(c *config.Config) error {
	logger.Init(c.Log)
	err := os.MkdirAll(c.Output.Dir, 0o755)
	if err != nil {
		return err
	}
	ctx := context.Background()
	tmpl, err := parseTemplates(c.Template)
	if err != nil {
		return err
	}
	err = cleanDir(c.Output.Dir)
	if err != nil {
		return err
	}
	return run(ctx, c, tmpl)
}

func run(ctx context.Context, c *config.Config, tmpl *templates) error {
	initialisms := c.Naming()

	for _, db := range c.DB {
		if err := generate(ctx, db, c, tmpl, initialisms); err != nil {
			return err
		}
	}
	return nil
}

func generate(ctx context.Context, db *config.DB, c *config.Config, tmpl *templates, initialisms map[string]string) error {
	s, err := schema.New(db)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			slog.ErrorContext(ctx, "close db", "error", err)
		}
	}()
	tables, err := s.Load(ctx)
	if err != nil {
		return err
	}
	for _, table := range tables {
		m := NewModel(table, db, c, initialisms)
		filename, err := tmpl.writeTo(ctx, c.Template.Model, c.Output.Dir, m.Table, m)
		if err != nil {
			return err
		}
		if c.Log.Level == "debug" {
			slog.DebugContext(ctx, "generate", "table", m.Table, "file", filename)
		}
	}
	return nil
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
