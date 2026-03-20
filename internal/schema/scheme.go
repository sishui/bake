package schema

import (
	"context"
	"database/sql"
	"fmt"
	"io"
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
