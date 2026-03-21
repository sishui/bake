// Package types provides PostgreSQL database type to Go type mapping.
package types

import (
	"errors"
	"strings"

	"github.com/sishui/bake/internal/schema"
)

func init() {
	DescFuncs["postgres"] = PostgresDescFunc
}

func PostgresDescFunc(c *schema.Column) (Desc, error) {
	switch c.DataType {
	case "int2", "smallint":
		return newNumericDesc(c, "int16", "int16"), nil
	case "int4", "integer":
		return newNumericDesc(c, "int32", "int32"), nil
	case "bigint", "int8":
		return newNumericDesc(c, "int64", "int64"), nil
	case "float4", "real":
		return newNumericDesc(c, "float32", "float32"), nil
	case "float8", "double precision":
		return newNumericDesc(c, "float64", "float64"), nil
	case "numeric", "decimal":
		return newDecimalDesc(c), nil
	case "bool":
		return newBoolDesc(c), nil
	case "text", "character varying", "character", "varchar", "char", "bpchar":
		return newStringDesc(c), nil
	case "bytea", "bit":
		return newBytesDesc(), nil
	case "date", "timestamp", "timestamp without time zone", "timestamp with time zone", "time", "time without time zone":
		return newTimeDesc(c), nil
	case "json", "jsonb":
		return newJSONDesc(), nil
	case "uuid":
		return newUUIDDesc(c), nil
	case "inet", "cidr":
		return newIPDesc(c), nil
	case "interval":
		return newIntervalDesc(c), nil
	case "ARRAY":
		return newArrayDesc(c)
	default:
		if strings.HasPrefix(c.DataType, "enum_") {
			return newEnumDesc(c), nil
		}
		return Desc{}, errors.Join(ErrUnsupportedType, errors.New(c.DataType))
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

var ErrUnsupportedArrayType = errors.New("unsupported array type")

func newArrayDesc(c *schema.Column) (Desc, error) {
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
		return Desc{}, errors.Join(ErrUnsupportedArrayType, errors.New(c.ColumnType))
	}
	return Desc{
		Type: typ,
		Kind: KindStruct,
		Imports: []string{
			"github.com/google/uuid",
		},
	}, nil
}
