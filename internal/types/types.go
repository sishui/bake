// Package types db type to go type mapping
package types

import (
	"github.com/sishui/bake/internal/schema"
)

type DescFunc func(c *schema.Column) (Desc, error)

var DescFuncs = map[string]DescFunc{}

type Desc struct {
	Type    string   // go Type
	Kind    string   // logic Type
	Imports []string // imports packages
}

func newNumericDesc(c *schema.Column, signed string, unsigned string) Desc {
	typ := signed
	if c.IsUnsigned() {
		typ = unsigned
	}
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type: typ,
		Kind: KindNumeric,
	}
}

func newStringDesc(c *schema.Column) Desc {
	typ := "string"
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type: typ,
		Kind: KindString,
	}
}

func newBytesDesc() Desc {
	typ := "[]byte"
	return Desc{
		Type: typ,
		Kind: KindBytes,
	}
}

func newTimeDesc(c *schema.Column) Desc {
	typ := "time.Time"
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type:    typ,
		Kind:    KindTime,
		Imports: []string{"time"},
	}
}

func newYearDesc(c *schema.Column) Desc {
	typ := "int"
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type: typ,
		Kind: KindNumeric,
	}
}

func newDecimalDesc(c *schema.Column) Desc {
	typ := "decimal.Decimal"
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type: typ,
		Kind: KindNumeric,
		Imports: []string{
			"github.com/shopspring/decimal",
		},
	}
}

func newJSONDesc() Desc {
	typ := "json.RawMessage"
	return Desc{
		Type: typ,
		Kind: KindBytes,
		Imports: []string{
			"encoding/json",
		},
	}
}

func newBoolDesc(c *schema.Column) Desc {
	typ := "bool"
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type: typ,
		Kind: KindBoolean,
	}
}

func newEnumDesc(c *schema.Column) Desc {
	typ := "string"
	if c.IsNullable() {
		typ = "*" + typ
	}
	return Desc{
		Type: typ,
		Kind: KindEnum,
	}
}

func newUUIDDesc(c *schema.Column) Desc {
	typ := "uuid.UUID"
	if c.IsNullable() {
		typ = "*" + typ
	}

	return Desc{
		Type: typ,
		Kind: KindStruct,
		Imports: []string{
			"github.com/google/uuid",
		},
	}
}
