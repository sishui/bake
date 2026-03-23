package generate

import (
	"reflect"
	"testing"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

func TestFieldImports(t *testing.T) {
	tests := []struct {
		name   string
		desc   types.Desc
		custom *config.CustomField
		want   []string
	}{
		{
			name:   "no custom field",
			desc:   types.Desc{Imports: []string{"time"}},
			custom: nil,
			want:   []string{"time"},
		},
		{
			name:   "no custom import",
			desc:   types.Desc{Imports: []string{"time"}},
			custom: &config.CustomField{},
			want:   []string{"time"},
		},
		{
			name:   "with custom import",
			desc:   types.Desc{Imports: []string{"time"}},
			custom: &config.CustomField{Import: "github.com/google/uuid"},
			want:   []string{"time", "github.com/google/uuid"},
		},
		{
			name:   "nil desc imports",
			desc:   types.Desc{},
			custom: &config.CustomField{Import: "github.com/google/uuid"},
			want:   []string{"github.com/google/uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldImports(tt.desc, tt.custom)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fieldImports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldName(t *testing.T) {
	tests := []struct {
		name        string
		col         *schema.Column
		custom      *config.CustomField
		initialisms map[string]string
		want        string
	}{
		{
			name:        "no custom field",
			col:         &schema.Column{Name: "user_name"},
			custom:      nil,
			initialisms: nil,
			want:        "UserName",
		},
		{
			name:        "with custom name",
			col:         &schema.Column{Name: "user_name"},
			custom:      &config.CustomField{Name: "Name"},
			initialisms: nil,
			want:        "Name",
		},
		{
			name:        "custom field with empty name",
			col:         &schema.Column{Name: "user_name"},
			custom:      &config.CustomField{},
			initialisms: nil,
			want:        "UserName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldName(tt.col, tt.custom, tt.initialisms)
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
			got := fieldType(tt.desc, tt.custom)
			if got != tt.want {
				t.Errorf("fieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []*Field
		want   int // number of groups
	}{
		{
			name:   "empty",
			fields: []*Field{},
			want:   0,
		},
		{
			name: "single field no comment",
			fields: []*Field{
				{Name: "ID"},
			},
			want: 1,
		},
		{
			name: "two fields no multi-line comment",
			fields: []*Field{
				{Name: "ID"},
				{Name: "Name"},
			},
			want: 1,
		},
		{
			name: "field with multi-line comment starts new group",
			fields: []*Field{
				{Name: "ID"},
				{Name: "Name", Comments: []string{"line1", "line2"}},
				{Name: "Email"},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupFields(tt.fields)
			if len(got) != tt.want {
				t.Errorf("groupFields() returned %d groups, want %d", len(got), tt.want)
			}
		})
	}
}

func TestMaxFieldAttr(t *testing.T) {
	fields := []*Field{
		{Name: "ID", Type: "int32", Tag: `bun:"id,pk"`},
		{Name: "UserName", Type: "string", Tag: `bun:"user_name"`},
		{Name: "Email", Type: "*string", Tag: `bun:"email,nullzero"`},
	}

	maxName, maxType, maxTag := maxFieldAttr(fields)
	if maxName != 8 { // "UserName"
		t.Errorf("maxName = %d, want 8", maxName)
	}
	if maxType != 7 { // "*string"
		t.Errorf("maxType = %d, want 7", maxType)
	}
	if maxTag != 20 { // `bun:"email,nullzero"`
		t.Errorf("maxTag = %d, want 20", maxTag)
	}
}
