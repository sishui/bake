// Package types db type to go type mapping
package types

import (
	"strings"

	"github.com/sishui/bake/internal/schema"
)

func MySQLDescFunc(c *schema.Column) Desc {
	switch c.DataType {
	case "tinyint":
		if strings.Contains(c.ColumnType, "(1)") {
			return newBoolDesc(c)
		}
		return newNumericDesc(c, "int8", "uint8")
	case "smallint":
		return newNumericDesc(c, "int16", "uint16")
	case "numeric", "integer", "int", "mediumint":
		return newNumericDesc(c, "int32", "uint32")
	case "bigint":
		return newNumericDesc(c, "int64", "uint64")
	case "float":
		return newNumericDesc(c, "float32", "float32")
	case "real", "double":
		return newNumericDesc(c, "float64", "float64")
	case "decimal":
		return newDecimalDesc(c)
	case "char", "varchar", "text", "longtext", "mediumtext", "tinytext":
		return newStringDesc(c)
	case "binary", "varbinary", "blob", "longblob", "mediumblob", "tinyblob", "bit":
		return newBytesDesc()
	case "date", "time", "datetime", "timestamp":
		return newTimeDesc(c)
	case "year":
		return newYearDesc(c)
	case "json":
		return newJSONDesc()
	case "enum", "set":
		return newEnumDesc(c)
	default:
		panic("unsupported type: " + c.DataType)
	}
}
