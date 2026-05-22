package schema

import (
	"context"
	"database/sql"
	"fmt"
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
    WHEN pk.contype = 'p' THEN
      'PRI'
    WHEN EXISTS (
        SELECT
          1
        FROM
          pg_index i
          JOIN pg_constraint uc ON uc.conindid = i.indexrelid
        WHERE
          i.indrelid = c.OID
          AND a.attnum = ANY (i.indkey)
          AND uc.contype = 'u'
      ) THEN
      'UNI'
    WHEN EXISTS (
        SELECT
          1
        FROM
          pg_index i
        WHERE
          i.indrelid = c.OID
          AND a.attnum = ANY (i.indkey)
          AND i.indisunique = FALSE
          AND i.indisprimary = FALSE
      ) THEN
      'MUL'
    ELSE
      ''
  END AS column_key,
  -- extra（PG没有auto_increment概念，用 identity / sequence 替代）
  CASE
    WHEN pg_get_serial_sequence (c.OID :: REGCLASS :: TEXT, a.attname) IS NOT NULL THEN
      'auto_increment'
    ELSE
      ''
  END AS extra,
  col_description (c.OID, a.attnum) AS column_comment
FROM
  pg_class c
  JOIN pg_namespace n ON n.OID = c.relnamespace
  JOIN pg_attribute a ON a.attrelid = c.OID
  LEFT JOIN pg_attrdef ad ON ad.adrelid = c.OID
  AND ad.adnum = a.attnum
  LEFT JOIN pg_type t ON t.OID = a.atttypid
  -- primary key constraint
  LEFT JOIN pg_constraint pk ON pk.conrelid = c.OID
  AND pk.contype = 'p'
  AND a.attnum = ANY (pk.conkey)
WHERE
  n.nspname = $1
  AND c.relkind = 'r'
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

const pgIndexQuery = `
SELECT
  t.relname AS table_name,
  CASE
    WHEN ix.indisunique THEN 0
    ELSE 1
  END AS non_unique,
  i.relname AS index_name,
  a.attname AS column_name
FROM pg_index ix
JOIN pg_class t
  ON t.oid = ix.indrelid
JOIN pg_class i
  ON i.oid = ix.indexrelid
JOIN pg_attribute a
    ON a.attrelid = t.oid
   AND a.attnum = ANY(ix.indkey)

WHERE t.relnamespace = (
    SELECT oid
    FROM pg_namespace
    WHERE nspname = $1
)
ORDER BY
    table_name,
    index_name;
`

func init() {
	Register("postgres", &postgresDriver{})
}

type postgresDriver struct{}

func (d *postgresDriver) Open(cfg *config.DB) (Scheme, error) {
	if cfg.Schema == "" {
		return nil, fmt.Errorf("db.schema is required for postgres")
	}
	db, err := openDB("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &postgresScheme{
		db:  db,
		cfg: cfg,
	}, nil
}

type postgresScheme struct {
	db  *sql.DB
	cfg *config.DB
}

func (s *postgresScheme) Load(ctx context.Context) ([]*Table, error) {
	slog.InfoContext(ctx, "load schema", "driver", s.cfg.Driver, "dsn", s.cfg.DSN, "schema", s.cfg.Schema)
	tables, err := loadTables(ctx, s.db, s.cfg.Schema, pgTableCommentsQuery, s.cfg)
	if err != nil {
		return nil, err
	}
	indexes, err := loadIndexes(ctx, s.db, s.cfg.Schema, pgIndexQuery)
	if err != nil {
		return nil, err
	}
	columns, err := loadColumns(ctx, s.db, s.cfg.Schema, pgColumnMetadataQuery, tables)
	if err != nil {
		return nil, err
	}
	foreignKeys, err := loadForeignKeys(ctx, s.db, s.cfg.Schema, pgForeignKeyQuery)
	if err != nil {
		return nil, err
	}
	assignColumns(tables, columns, indexes)
	assignForeignKeys(tables, foreignKeys)
	return tables, nil
}

func (s *postgresScheme) Close() error {
	return s.db.Close()
}
