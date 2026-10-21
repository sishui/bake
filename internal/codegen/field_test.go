package codegen

import (
	"reflect"
	"testing"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

func TestFieldImports(t *testing.T) {
	uuid := &config.Import{Package: "github.com/google/uuid", Types: []string{"uuid.UUID"}}
	dec := &config.Import{Package: "github.com/shopspring/decimal", Alias: "dec", Types: []string{"dec.Decimal"}}

	tests := []struct {
		name    string
		desc    types.Desc
		custom  *config.CustomField
		imports []*config.Import
		want    []Import
	}{
		{
			name:   "no custom field",
			desc:   types.Desc{Imports: []string{"time"}},
			custom: nil,
			want:   []Import{{Path: "time"}},
		},
		{
			name:   "no type override",
			desc:   types.Desc{Imports: []string{"time"}},
			custom: &config.CustomField{},
			want:   []Import{{Path: "time"}},
		},
		{
			name:   "type override without a configured import",
			desc:   types.Desc{Imports: []string{"time"}},
			custom: &config.CustomField{Type: "string"},
			want:   []Import{{Path: "time"}},
		},
		{
			name:    "configured import matched by type",
			desc:    types.Desc{Imports: []string{"time"}},
			custom:  &config.CustomField{Type: "uuid.UUID"},
			imports: []*config.Import{uuid},
			want:    []Import{{Path: "time"}, {Path: "github.com/google/uuid"}},
		},
		{
			name:    "pointer type still matches",
			desc:    types.Desc{},
			custom:  &config.CustomField{Type: "*uuid.UUID"},
			imports: []*config.Import{uuid},
			want:    []Import{{Path: "github.com/google/uuid"}},
		},
		{
			name:    "alias is rendered in the import spec",
			desc:    types.Desc{},
			custom:  &config.CustomField{Type: "[]dec.Decimal"},
			imports: []*config.Import{dec},
			want:    []Import{{Path: "github.com/shopspring/decimal", Alias: "dec"}},
		},
		{
			name:    "slice of pointers matches",
			desc:    types.Desc{},
			custom:  &config.CustomField{Type: "[]*uuid.UUID"},
			imports: []*config.Import{uuid},
			want:    []Import{{Path: "github.com/google/uuid"}},
		},
		{
			name:    "map value and key match",
			desc:    types.Desc{},
			custom:  &config.CustomField{Type: "map[uuid.UUID]dec.Decimal"},
			imports: []*config.Import{uuid, dec},
			want:    []Import{{Path: "github.com/google/uuid"}, {Path: "github.com/shopspring/decimal", Alias: "dec"}},
		},
		{
			name:    "partial identifier does not match",
			desc:    types.Desc{},
			custom:  &config.CustomField{Type: "myuuid.UUID"},
			imports: []*config.Import{uuid},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fieldImports(tt.desc, tt.custom, tt.imports)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fieldImports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldName(t *testing.T) {
	n := naming.New()
	tests := []struct {
		name   string
		col    *schema.Column
		custom *config.CustomField
		want   string
	}{
		{
			name:   "no custom field",
			col:    &schema.Column{Name: "user_name"},
			custom: nil,
			want:   "UserName",
		},
		{
			name:   "with custom name",
			col:    &schema.Column{Name: "user_name"},
			custom: &config.CustomField{Name: "Name"},
			want:   "Name",
		},
		{
			name:   "custom field with empty name",
			col:    &schema.Column{Name: "user_name"},
			custom: &config.CustomField{},
			want:   "UserName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fieldName(tt.col, tt.custom, n)
			if got != tt.want {
				t.Errorf("fieldName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType(t *testing.T) {
	tests := []struct {
		name   string
		desc   types.Desc
		custom *config.CustomField
		want   string
	}{
		{
			name:   "no custom field",
			desc:   types.Desc{Type: "string"},
			custom: nil,
			want:   "string",
		},
		{
			name:   "with custom type",
			desc:   types.Desc{Type: "string"},
			custom: &config.CustomField{Type: "*User"},
			want:   "*User",
		},
		{
			name:   "custom field with empty type",
			desc:   types.Desc{Type: "int32"},
			custom: &config.CustomField{},
			want:   "int32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fieldType(tt.desc, tt.custom)
			if got != tt.want {
				t.Errorf("fieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}
