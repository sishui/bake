package schema

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/sishui/bake/internal/config"
)

const pgTableCommentsQuery = `
SELECT
  c.relname AS TABLE_NAME,
  obj_description (c.OID) AS table_comment
FROM
  pg_class c
  JOIN pg_namespace n ON n.OID = c.relnamespace
WHERE
  c.relkind = 'r'
  AND n.nspname = $1
ORDER BY
  c.relname;
`

const pgColumnMetadataQuery = `
SELECT
  c.relname AS TABLE_NAME,
  a.attname AS COLUMN_NAME,
  a.attnum AS ordinal_position,
  pg_get_expr (ad.adbin, ad.adrelid) AS column_default,
  NOT a.attnotnull AS is_nullable,
  t.typname AS data_type,
  format_type (a.atttypid, a.atttypmod) AS column_type,
  CASE
    WHEN pk.attname IS NOT NULL THEN
      'PRI'
    ELSE
      ''
  END AS column_key,
  '' AS extra,
  col_description (a.attrelid, a.attnum) AS column_comment
FROM
  pg_attribute a
  JOIN pg_class c ON a.attrelid = c.OID
  JOIN pg_namespace n ON n.OID = c.relnamespace
  JOIN pg_type t ON a.atttypid = t.OID
  LEFT JOIN pg_attrdef ad ON a.attrelid = ad.adrelid AND a.attnum = ad.adnum
  LEFT JOIN (
    SELECT
      a.attname,
      i.indrelid
    FROM
      pg_index i
      JOIN pg_attribute a ON a.attrelid = i.indrelid
      AND a.attnum = ANY (i.indkey)
    WHERE
      i.indisprimary
  ) pk ON pk.indrelid = c.OID
  AND pk.attname = a.attname
WHERE
  n.nspname = $1
  AND a.attnum > 0
  AND NOT a.attisdropped
ORDER BY
  c.relname,
  a.attnum;
`

const pgForeignKeyQuery = `
SELECT
  con.conname AS constraint_name,
  rel.relname AS table_name,
  att.attname AS column_name,
  frel.relname AS referenced_table_name,
  fatt.attname AS referenced_column_name
FROM pg_constraint con
JOIN pg_class rel ON rel.oid = con.conrelid
JOIN pg_class frel ON frel.oid = con.confrelid
JOIN unnest(con.conkey) WITH ORDINALITY AS cols(attnum, ord) ON TRUE
JOIN pg_attribute att ON att.attrelid = rel.oid AND att.attnum = cols.attnum
JOIN unnest(con.confkey) WITH ORDINALITY AS fcols(attnum, ord) ON cols.ord = fcols.ord
JOIN pg_attribute fatt ON fatt.attrelid = frel.oid AND fatt.attnum = fcols.attnum
WHERE con.contype = 'f'
  AND rel.relnamespace = (
    SELECT oid FROM pg_namespace WHERE nspname = $1
  );
`

type postgres struct {
	db  *sql.DB
	cfg *config.DB
}

func NewPostgres(cfg *config.DB) (Scheme, error) {
	db, err := openDB("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &postgres{
		db:  db,
		cfg: cfg,
	}, nil
}

func (s *postgres) Load(ctx context.Context) ([]*Table, error) {
	slog.InfoContext(ctx, "load schema", "driver", s.cfg.Driver, "dsn", s.cfg.DSN, "schema", s.cfg.Schema)
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
	assignForeignKeys(tables, foreignKeys)
	return tables, nil
}

func (s *postgres) Close() error {
	return s.db.Close()
}

func (s *postgres) loadTables(ctx context.Context) ([]*Table, error) {
	rows, err := s.db.QueryContext(ctx, pgTableCommentsQuery, s.cfg.Schema)
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
		var comment sql.NullString
		err = rows.Scan(&t.Name, &comment)
		if err != nil {
			return nil, err
		}
		if !shouldIncludeTable(ctx, t.Name, s.cfg) {
			continue
		}
		t.Comment = comment.String
		result = append(result, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *postgres) loadColumns(ctx context.Context, tables []*Table) (map[string][]*Column, error) {
	rows, err := s.db.QueryContext(ctx, pgColumnMetadataQuery, s.cfg.Schema)
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
		var comment sql.NullString
		err = rows.Scan(&c.Table, &c.Name, &c.OrdinalPosition, &columnDefault, &c.Nullable, &c.DataType, &c.ColumnType, &c.Key, &c.Extra, &comment)
		if err != nil {
			return nil, err
		}
		c.Default = columnDefault.String
		c.Comment = comment.String
		result[c.Table] = append(result[c.Table], &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *postgres) loadForeignKeys(ctx context.Context) ([]ForeignKey, error) {
	rows, err := s.db.QueryContext(ctx, pgForeignKeyQuery, s.cfg.Schema)
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
		err = rows.Scan(&fk.ConstraintName, &fk.Table, &fk.ColumnName, &fk.RefTable, &fk.RefColumn)
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
