// Package types provides database column type to Go type mapping.
package types

import (
	"fmt"

	"github.com/sishui/bake/internal/schema"
)

// Mapper is the interface that wraps the Desc method for type mapping.
type Mapper interface {
	Desc(c *schema.Column) (Desc, error)
}

var mappers = map[string]Mapper{}

// Register registers a mapper for type mapping.
func Register(name string, m Mapper) {
	mappers[name] = m
}

// NewDesc returns a Desc for the given mapper name and column.
func NewDesc(name string, c *schema.Column) (Desc, error) {
	m, ok := mappers[name]
	if !ok {
		return Desc{}, fmt.Errorf("unsupported mapper: %s", name)
	}
	return m.Desc(c)
}

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

// newBytesDesc returns []byte type. Nullable is not applied because
// []byte's zero value (nil) already represents SQL NULL.
func newBytesDesc(_ *schema.Column) Desc {
	return Desc{
		Type: "[]byte",
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

// newGeometryDesc returns []byte for spatial types (WKB format).
// Drivers return spatial data as WKB bytes; users can parse with their preferred library.
func newGeometryDesc(_ *schema.Column) Desc {
	return Desc{
		Type: "[]byte",
		Kind: KindBytes,
	}
}
