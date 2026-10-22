package codegen

import (
	"reflect"
	"testing"

	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
)

func TestNewConflicts(t *testing.T) {
	columns := []*schema.Column{
		{Name: "id", Key: schema.ColumnKeyPrimary},
		{Name: "tenant_id"},
		{Name: "email"},
		{Name: "name"},
		{Name: "created_at"},
	}

	tests := []struct {
		name string
		keys []schema.UniqueKey
		cols []*schema.Column
		want []Conflict
	}{
		{
			name: "empty",
			cols: columns,
		},
		{
			name: "primary key",
			keys: []schema.UniqueKey{{Columns: []string{"id"}}},
			cols: columns,
			want: []Conflict{{Key: "ID", Value: "CONFLICT (id) DO UPDATE"}},
		},
		{
			name: "unique column",
			keys: []schema.UniqueKey{{Columns: []string{"email"}}},
			cols: columns,
			want: []Conflict{{Key: "Email", Value: "CONFLICT (email) DO UPDATE"}},
		},
		{
			name: "composite key keeps column order",
			keys: []schema.UniqueKey{{Columns: []string{"tenant_id", "email"}}},
			cols: columns,
			want: []Conflict{{
				Key:   "TenantIDEmail",
				Value: "CONFLICT (tenant_id, email) DO UPDATE",
			}},
		},
		{
			name: "nothing left to update",
			keys: []schema.UniqueKey{{Columns: []string{"id"}}},
			cols: []*schema.Column{{Name: "id", Key: schema.ColumnKeyPrimary}},
			want: []Conflict{{Key: "ID", Value: "CONFLICT (id) DO NOTHING"}},
		},
		{
			name: "same columns under two constraints emit one constant",
			keys: []schema.UniqueKey{
				{Columns: []string{"id"}},
				{Columns: []string{"id"}},
			},
			cols: columns,
			want: []Conflict{{Key: "ID", Value: "CONFLICT (id) DO UPDATE"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := newConflicts(tt.keys, tt.cols, naming.New("ID"))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newConflicts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewExcluded(t *testing.T) {
	columns := []*schema.Column{
		{Name: "id", Key: schema.ColumnKeyPrimary},
		{Name: "tenant_id"},
		{Name: "email"},
		{Name: "created_at"},
	}

	tests := []struct {
		name   string
		driver string
		want   []Excluded
	}{
		{
			name:   "postgres reads the excluded row",
			driver: "postgres",
			want: []Excluded{
				{Key: "TenantID", Value: "tenant_id = EXCLUDED.tenant_id"},
				{Key: "Email", Value: "email = EXCLUDED.email"},
			},
		},
		{
			name:   "mysql uses VALUES()",
			driver: "mysql",
			want: []Excluded{
				{Key: "TenantID", Value: "tenant_id = VALUES(tenant_id)"},
				{Key: "Email", Value: "email = VALUES(email)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := newExcluded(columns, naming.New("ID"), tt.driver)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}
