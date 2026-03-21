// Package types db type to go type mapping
package types

import (
	"errors"
	"strings"

	"github.com/sishui/bake/internal/schema"
)

var ErrUnsupportedType = errors.New("unsupported type")

func init() {
	DescFuncs["mysql"] = MySQLDescFunc
}

func MySQLDescFunc(c *schema.Column) (Desc, error) {
	switch c.DataType {
	case "tinyint":
		if strings.Contains(c.ColumnType, "(1)") {
			return newBoolDesc(c), nil
		}
		return newNumericDesc(c, "int8", "uint8"), nil
	case "smallint":
		return newNumericDesc(c, "int16", "uint16"), nil
	case "numeric", "integer", "int", "mediumint":
		return newNumericDesc(c, "int32", "uint32"), nil
	case "bigint":
		return newNumericDesc(c, "int64", "uint64"), nil
	case "float":
		return newNumericDesc(c, "float32", "float32"), nil
	case "real", "double":
		return newNumericDesc(c, "float64", "float64"), nil
	case "decimal":
		return newDecimalDesc(c), nil
	case "char", "varchar", "text", "longtext", "mediumtext", "tinytext":
		return newStringDesc(c), nil
	case "binary", "varbinary", "blob", "longblob", "mediumblob", "tinyblob", "bit":
		return newBytesDesc(), nil
	case "date", "time", "datetime", "timestamp":
		return newTimeDesc(c), nil
	case "year":
		return newYearDesc(c), nil
	case "json":
		return newJSONDesc(), nil
	case "enum", "set":
		return newEnumDesc(c), nil
	default:
		return Desc{}, errors.Join(ErrUnsupportedType, errors.New(c.DataType))
	}
}
