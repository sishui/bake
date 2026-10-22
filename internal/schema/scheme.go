package schema

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sort"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sishui/bake/internal/config"
)

// Driver is the interface for database driver registration.
// Each driver must register itself via init() to be discoverable.
type Driver interface {
	// Open creates a new Scheme for the given configuration.
	Open(cfg *config.DB) (Scheme, error)
}

var drivers = make(map[string]Driver)

// Register registers a database driver by name.
// Called from init() in driver implementations.
func Register(name string, driver Driver) {
	if driver == nil {
		panic("schema: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("schema: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

type Scheme interface {
	io.Closer
	Load(ctx context.Context) ([]*Table, error)
}

func New(cfg *config.DB) (Scheme, error) {
	driver, ok := drivers[cfg.Driver]
	if !ok {
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
	return driver.Open(cfg)
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
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
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
	result := make([]Index, 0)
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

func assignIndexes(tables []*Table, indexes []Index) {
	tableMap := make(map[string]*Table, len(tables))
	pkColumns := make(map[string]map[string]struct{}, len(tables))
	for _, t := range tables {
		tableMap[t.Name] = t
		pkColumns[t.Name] = primaryColumns(t)
	}

	columns, unique, order := groupIndexColumns(indexes)
	for _, key := range order {
		if t, ok := tableMap[key.table]; ok {
			t.addIndex(key, columns[key], unique[key], pkColumns[key.table])
		}
	}

	for _, t := range tables {
		sort.Strings(t.Indexes)
	}
}

// groupIndexColumns groups index rows by (table, index) keeping the column
// order of each index, and reports whether it is unique.
func groupIndexColumns(indexes []Index) (map[tableIndex][]string, map[tableIndex]bool, []tableIndex) {
	columns := make(map[tableIndex][]string, len(indexes))
	unique := make(map[tableIndex]bool, len(indexes))
	order := make([]tableIndex, 0, len(indexes))
	for _, idx := range indexes {
		if idx.IndexName == "" {
			continue
		}
		key := tableIndex{table: idx.Table, indexName: idx.IndexName}
		if _, ok := columns[key]; !ok {
			order = append(order, key)
			unique[key] = idx.NonUnique == 0
		}
		columns[key] = append(columns[key], idx.ColumnName)
	}
	return columns, unique, order
}

func (t *Table) addIndex(key tableIndex, columns []string, unique bool, pkColumns map[string]struct{}) {
	if key.indexName != IndexNamePrimary && !isPrimaryIndex(pkColumns, columns) {
		t.Indexes = append(t.Indexes, key.indexName)
	}
	if unique && !t.hasUniqueKey(columns) {
		t.Keys = append(t.Keys, UniqueKey{Columns: columns})
	}
}

func primaryColumns(t *Table) map[string]struct{} {
	columns := make(map[string]struct{})
	for _, c := range t.Columns {
		if c.IsPrimaryKey() {
			columns[c.Name] = struct{}{}
		}
	}
	return columns
}

func (t *Table) hasUniqueKey(columns []string) bool {
	for _, key := range t.Keys {
		if slices.Equal(key.Columns, columns) {
			return true
		}
	}
	return false
}

// isPrimaryIndex reports whether every column of the index is a primary key
// column, i.e. the index is the one backing the primary key.
func isPrimaryIndex(pkColumns map[string]struct{}, columns []string) bool {
	if len(pkColumns) == 0 || len(columns) == 0 {
		return false
	}
	for _, c := range columns {
		if _, ok := pkColumns[c]; !ok {
			return false
		}
	}
	return true
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
	for _, m := range matches {
		if m.IndexName == IndexNamePrimary {
			continue
		}
		if m.NonUnique == 0 {
			return m, true
		}
	}
	return Index{}, false
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
