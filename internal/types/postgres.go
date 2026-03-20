package types

import (
	"strings"

	"github.com/sishui/bake/internal/schema"
)

func PostgresDescFunc(c *schema.Column) Desc {
	switch c.DataType {
	case "int2", "smallint":
		return newNumericDesc(c, "int16", "int16")
	case "int4", "integer":
		return newNumericDesc(c, "int32", "int32")
	case "bigint", "int8":
		return newNumericDesc(c, "int64", "int64")
	case "float4", "real":
		return newNumericDesc(c, "float32", "float32")
	case "float8", "double precision":
		return newNumericDesc(c, "float64", "float64")
	case "numeric", "decimal":
		return newDecimalDesc(c)
	case "bool":
		return newBoolDesc(c)
	case "text", "character varying", "character", "varchar", "char", "bpchar":
		return newStringDesc(c)
	case "bytea", "bit":
		return newBytesDesc()
	case "date", "timestamp", "timestamp without time zone", "timestamp with time zone", "time", "time without time zone":
		return newTimeDesc(c)
	case "json", "jsonb":
		return newJSONDesc()
	case "uuid":
		return newUUIDDesc(c)
	case "inet", "cidr":
		return newIPDesc(c)
	case "interval":
		return newIntervalDesc(c)
	case "ARRAY":
		return newArrayDesc(c)
	default:
		if strings.HasPrefix(c.DataType, "enum_") {
			return newEnumDesc(c)
		}
		panic("unsupported type: " + c.DataType)
	}
}

func newIPDesc(c *schema.Column) Desc {
	typ := "net.IP"
	if c.IsNullable() {
		typ = "*" + typ
	}

	return Desc{
		Type: typ,
		Kind: KindStruct,
		Imports: []string{
			"net",
		},
	}
}

func newIntervalDesc(c *schema.Column) Desc {
	typ := "time.Duration"
	if c.IsNullable() {
		typ = "*" + typ
	}

	return Desc{
		Type: typ,
		Kind: KindTime,
		Imports: []string{
			"time",
		},
	}
}

func newArrayDesc(c *schema.Column) Desc {
	typ := ""
	switch c.ColumnType {
	case "_text", "_varchar":
		typ = "[]string"
	case "_int4":
		typ = "[]int32"
	case "_int8":
		typ = "[]int64"
	case "_uuid":
		typ = "[]uuid.UUID"
	default:
		panic("unsupported array type: " + c.ColumnType)
	}
	return Desc{
		Type: typ,
		Kind: KindStruct,
		Imports: []string{
			"github.com/google/uuid",
		},
	}
}
