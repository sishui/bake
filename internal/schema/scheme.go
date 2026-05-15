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

func loadTables(ctx context.Context, db *sql.DB, schema string, querySQL string, cfg *config.DB) ([]*Table, error) {
	rows, err := db.QueryContext(ctx, querySQL, schema)
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
		if !shouldIncludeTable(ctx, t.Name, cfg) {
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

func loadIndexes(ctx context.Context, db *sql.DB, schema string, querySQL string) ([]Index, error) {
	rows, err := db.QueryContext(ctx, querySQL, schema)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "close rows", "error", err)
		}
	}()
	var result []Index
	for rows.Next() {
		var idx Index
		err = rows.Scan(&idx.Table, &idx.NonUnique, &idx.IndexName, &idx.ColumnName)
		if err != nil {
			return nil, err
		}
		result = append(result, idx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func loadForeignKeys(ctx context.Context, db *sql.DB, scheme string, querySQL string) ([]ForeignKey, error) {
	rows, err := db.QueryContext(ctx, querySQL, scheme)
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

func loadColumns(ctx context.Context, db *sql.DB, schema string, querySQL string, tables []*Table) (map[string][]*Column, error) {
	rows, err := db.QueryContext(ctx, querySQL, schema)
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

// shouldIncludeTable returns true if the table should be included based on config.
func shouldIncludeTable(ctx context.Context, name string, cfg *config.DB) bool {
	if slices.Index(cfg.Exclude, name) >= 0 {
		slog.DebugContext(ctx, "table", name, "excluded table")
		return false
	}
	if len(cfg.Include) > 0 && slices.Index(cfg.Include, name) == -1 {
		slog.DebugContext(ctx, "table", name, "skipping table")
		return false
	}
	return true
}

type tableColumn struct {
	table string
	col   string
}

type tableIndex struct {
	table     string
	indexName string
}

func columnIndexes(indexes []Index) map[tableColumn][]Index {
	m := make(map[tableColumn][]Index, len(indexes))
	for _, idx := range indexes {
		key := tableColumn{table: idx.Table, col: idx.ColumnName}
		m[key] = append(m[key], idx)
	}
	return m
}

func bestIndexForColumn(colIndexes map[tableColumn][]Index, table, col string) (Index, bool) {
	matches := colIndexes[tableColumn{table: table, col: col}]
	if len(matches) == 0 {
		return Index{}, false
	}
	best := matches[0]
	for _, m := range matches[1:] {
		if m.NonUnique < best.NonUnique {
			best = m
		}
	}
	return best, true
}

func assignColumns(tables []*Table, columns map[string][]*Column, indexes []Index) {
	colIndexes := columnIndexes(indexes)

	// Group indexes by (table, index_name) to detect composite indexes.
	indexMembers := make(map[tableIndex]int)
	for _, idx := range indexes {
		key := tableIndex{table: idx.Table, indexName: idx.IndexName}
		indexMembers[key]++
	}

	for _, t := range tables {
		t.Columns = columns[t.Name]
		sort.Slice(t.Columns, func(i, j int) bool {
			return t.Columns[i].OrdinalPosition < t.Columns[j].OrdinalPosition
		})
		for _, column := range t.Columns {
			if column.Key == "PRI" {
				continue
			}
			best, ok := bestIndexForColumn(colIndexes, t.Name, column.Name)
			if !ok {
				continue
			}
			column.IndexName = best.IndexName
			column.NonUnique = best.NonUnique
			column.IsMultiKey = best.NonUnique == 0 && indexMembers[tableIndex{table: t.Name, indexName: best.IndexName}] >= 2
		}
	}
}

// assignForeignKeys assigns foreign key information to tables and columns.
func assignForeignKeys(tables []*Table, foreignKeys []ForeignKey) {
	// Create a map for quick table lookup
	tableMap := make(map[string]*Table, len(tables))
	for _, t := range tables {
		tableMap[t.Name] = t
	}

	// Assign foreign keys to tables and columns
	for _, fk := range foreignKeys {
		// Find the source table
		sourceTable, ok := tableMap[fk.Table]
		if !ok {
			continue
		}

		// Find the column in the source table
		for _, c := range sourceTable.Columns {
			if c.Name == fk.ColumnName {
				c.ForeignKey = &fk
				sourceTable.ForeignKeys = append(sourceTable.ForeignKeys, fk)

				// Add reverse foreign key to the referenced table
				if refTable, ok := tableMap[fk.RefTable]; ok {
					refTable.ReverseForeignKeys = append(refTable.ReverseForeignKeys, fk)
				}
				break
			}
		}
	}
}
