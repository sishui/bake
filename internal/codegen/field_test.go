package generate

import (
	"reflect"
	"testing"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
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
			t.Parallel()
			got := fieldImports(tt.desc, tt.custom)
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

