// Package schema contains the schema for the database.
package schema

import "strings"

type Table struct {
	Name               string
	Comment            string
	Columns            []*Column
	ForeignKeys        []ForeignKey
	ReverseForeignKeys []ForeignKey // Foreign keys from other tables referencing this table
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
	ForeignKey      *ForeignKey
	IndexName       string
	NonUnique       int
	IsMultiKey      bool
}

type ForeignKey struct {
	ConstraintName string // FK constraint name
	Table          string // Source table (e.g., "posts")
	ColumnName     string // Column in source table (e.g., "user_id")
	RefTable       string // Referenced table (e.g., "users")
	RefColumn      string // Referenced column (e.g., "id")
}

type Index struct {
	Table      string
	NonUnique  int
	ColumnName string
	IndexName  string
}

func (c *Column) IsNullable() bool {
	switch c.Nullable {
	case "YES", "y", "yes", "true", "t":
		return true
	default:
		return false
	}
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

func (c *Column) IsForeignKey() bool {
	return c.ForeignKey != nil
}

func (c *Column) IsUnique() bool {
	return c.IndexName != "" && c.NonUnique == 0
}
