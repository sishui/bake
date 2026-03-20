// Package schema contains the schema for the database.
package schema

import "strings"

type Table struct {
	Name    string
	Comment string
	Columns []*Column
}

type Column struct {
	Table           string
	Name            string // ColumnName
	OrdinalPosition int
	Default         string // ColumnDefault
	Nullable        string
	DataType        string
	ColumnType      string
	Key             string // ColumnKey
	Extra           string
	Comment         string // ColumnComment
}

func (c *Column) IsNullable() bool {
	return c.Nullable == "YES" || c.Nullable == "true"
}

func (c *Column) IsPrimaryKey() bool {
	return c.Key == "PRI"
}

func (c *Column) IsAutoIncrement() bool {
	return c.Extra == "auto_increment"
}

func (c *Column) IsUnsigned() bool {
	return strings.Contains(c.ColumnType, "unsigned")
}
