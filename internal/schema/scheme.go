package schema

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sishui/bake/internal/config"
)

// var dsnRegex = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*)://(?:[^@]+@)?([^/]+)/([^?]+)`)

type Scheme interface {
	io.Closer
	Load(ctx context.Context) ([]*Table, error)
}

func New(cfg *config.DB) (Scheme, error) {
	switch cfg.Driver {
	case "mysql":
		return NewMySQL(cfg)
	case "postgres":
		return NewPostgres(cfg)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}

func openDB(driver string, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxIdleTime(time.Minute)
	db.SetConnMaxLifetime(time.Minute * 3)
	return db, nil
}

// shouldIncludeTable returns true if the table should be included based on config.
func shouldIncludeTable(ctx context.Context, name string, cfg *config.DB) bool {
	if slices.Index(cfg.Excluded, name) >= 0 {
		slog.DebugContext(ctx, "table", name, "excluded table")
		return false
	}
	if len(cfg.Included) > 0 && slices.Index(cfg.Included, name) == -1 {
		slog.DebugContext(ctx, "table", name, "skipping table")
		return false
	}
	return true
}

// assignColumns assigns columns to tables and sorts them by ordinal position.
func assignColumns(tables []*Table, columns map[string][]*Column) {
	for _, t := range tables {
		t.Columns = columns[t.Name]
		sort.Slice(t.Columns, func(i, j int) bool {
			return t.Columns[i].OrdinalPosition < t.Columns[j].OrdinalPosition
		})
	}
}
