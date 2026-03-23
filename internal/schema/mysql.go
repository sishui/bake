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

type mysqlScheme struct {
	db       *sql.DB
	cfg      *config.DB
	database string
}

func NewMySQL(cfg *config.DB) (Scheme, error) {
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

func (s *mysqlScheme) Load(ctx context.Context) ([]*Table, error) {
	slog.InfoContext(ctx, "load schema", "driver", s.cfg.Driver, "dsn", s.cfg.DSN, "database", s.database)
	tables, err := s.loadTables(ctx)
	if err != nil {
		return nil, err
	}
	columns, err := s.loadColumns(ctx, tables)
	if err != nil {
		return nil, err
	}
	foreignKeys, err := s.loadForeignKeys(ctx)
	if err != nil {
		return nil, err
	}
	assignColumns(tables, columns)
	assignForeignKeys(tables, columns, foreignKeys)
	return tables, nil
}

func (s *mysqlScheme) Close() error {
	return s.db.Close()
}

func (s *mysqlScheme) loadTables(ctx context.Context) ([]*Table, error) {
	rows, err := s.db.QueryContext(ctx, mysqlTablesCommentQuery, s.database)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "close rows", "error", err)
		}
	}()
	result := make([]*Table, 0, 128)
	for rows.Next() {
		var t Table
		err = rows.Scan(&t.Name, &t.Comment)
		if err != nil {
			return nil, err
		}
		if !shouldIncludeTable(ctx, t.Name, s.cfg) {
			continue
		}
		result = append(result, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *mysqlScheme) loadColumns(ctx context.Context, tables []*Table) (map[string][]*Column, error) {
	rows, err := s.db.QueryContext(ctx, mysqlColumnMetadataQuery, s.database)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "close rows", "error", err)
		}
	}()
	result := make(map[string][]*Column, len(tables))
	for _, t := range tables {
		result[t.Name] = make([]*Column, 0, 32)
	}
	for rows.Next() {
		var c Column
		var columnDefault sql.NullString
		err = rows.Scan(&c.Table, &c.Name, &c.OrdinalPosition, &columnDefault, &c.Nullable, &c.DataType, &c.ColumnType, &c.Key, &c.Extra, &c.Comment)
		if err != nil {
			return nil, err
		}
		c.Default = columnDefault.String
		result[c.Table] = append(result[c.Table], &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *mysqlScheme) loadForeignKeys(ctx context.Context) ([]ForeignKey, error) {
	rows, err := s.db.QueryContext(ctx, mysqlForeignKeyQuery, s.database)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "close rows", "error", err)
		}
	}()
	var result []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		var tableName string
		err = rows.Scan(&fk.ConstraintName, &tableName, &fk.ColumnName, &fk.RefTable, &fk.RefColumn)
		if err != nil {
			return nil, err
		}
		result = append(result, fk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
