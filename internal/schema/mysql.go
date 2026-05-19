package schema

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/go-sql-driver/mysql"

	"github.com/sishui/bake/internal/config"
)

const mysqlTablesCommentQuery = `
SELECT
  table_name,
  table_comment
FROM
  information_schema.TABLES
WHERE
  table_schema = ?;
`

const mysqlColumnMetadataQuery = `
SELECT
  table_name,
  column_name,
  ordinal_position,
  column_default,
  is_nullable,
  data_type,
  column_type,
  column_key,
  extra,
  column_comment
FROM
  information_schema.COLUMNS
WHERE
  table_schema = ?
ORDER BY
  ordinal_position;
`

const mysqlForeignKeyQuery = `
SELECT
  constraint_name,
  table_name,
  column_name,
  referenced_table_name,
  referenced_column_name
FROM
  information_schema.KEY_COLUMN_USAGE
WHERE
  table_schema = ?
  AND referenced_table_name IS NOT NULL
ORDER BY
  constraint_name,
  ordinal_position;
`

const mysqlIndexQuery = `
SELECT
  table_name,
  non_unique,
  index_name,
  column_name
FROM
  information_schema.STATISTICS
WHERE
  table_schema = ?
ORDER BY
  table_name,
  index_name;
`

func init() {
	Register("mysql", &mysqlDriver{})
}

type mysqlDriver struct{}

func (d *mysqlDriver) Open(cfg *config.DB) (Scheme, error) {
	c, err := mysql.ParseDSN(cfg.DSN)
	if err != nil {
		return nil, err
	}
	db, err := openDB("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &mysqlScheme{
		db:       db,
		cfg:      cfg,
		database: c.DBName,
	}, nil
}

type mysqlScheme struct {
	db       *sql.DB
	cfg      *config.DB
	database string
}

func (s *mysqlScheme) Load(ctx context.Context) ([]*Table, error) {
	slog.InfoContext(ctx, "load schema", "driver", s.cfg.Driver, "dsn", s.cfg.DSN, "database", s.database)
	tables, err := loadTables(ctx, s.db, s.database, mysqlTablesCommentQuery, s.cfg)
	if err != nil {
		return nil, err
	}
	indexes, err := loadIndexes(ctx, s.db, s.database, mysqlIndexQuery)
	if err != nil {
		return nil, err
	}
	columns, err := loadColumns(ctx, s.db, s.database, mysqlColumnMetadataQuery, tables)
	if err != nil {
		return nil, err
	}
	foreignKeys, err := loadForeignKeys(ctx, s.db, s.database, mysqlForeignKeyQuery)
	if err != nil {
		return nil, err
	}
	assignColumns(tables, columns, indexes)
	assignForeignKeys(tables, foreignKeys)
	return tables, nil
}

func (s *mysqlScheme) Close() error {
	return s.db.Close()
}
