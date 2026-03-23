package schema

import (
	"context"
	"testing"

	"github.com/sishui/bake/internal/config"
)

func TestShouldIncludeTable(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		excluded []string
		included []string
		want     bool
	}{
		{
			name:  "no filters - include all",
			table: "users",
			want:  true,
		},
		{
			name:     "excluded table",
			table:    "users",
			excluded: []string{"users", "posts"},
			want:     false,
		},
		{
			name:     "not in excluded",
			table:    "comments",
			excluded: []string{"users", "posts"},
			want:     true,
		},
		{
			name:     "in included list",
			table:    "users",
			included: []string{"users", "posts"},
			want:     true,
		},
		{
			name:     "not in included list",
			table:    "comments",
			included: []string{"users", "posts"},
			want:     false,
		},
		{
			name:     "excluded takes priority over included",
			table:    "users",
			excluded: []string{"users"},
			included: []string{"users", "posts"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.DB{
				Excluded: tt.excluded,
				Included: tt.included,
			}
			got := shouldIncludeTable(context.Background(), tt.table, cfg)
			if got != tt.want {
				t.Errorf("shouldIncludeTable(%q) = %v, want %v", tt.table, got, tt.want)
			}
		})
	}
}

func TestAssignColumns(t *testing.T) {
	tables := []*Table{
		{Name: "users"},
		{Name: "posts"},
	}
	columns := map[string][]*Column{
		"users": {
			{Name: "id", OrdinalPosition: 2},
			{Name: "name", OrdinalPosition: 1},
			{Name: "email", OrdinalPosition: 3},
		},
		"posts": {
			{Name: "title", OrdinalPosition: 1},
			{Name: "id", OrdinalPosition: 0},
		},
	}

	assignColumns(tables, columns)

	// Check users columns are sorted
	if len(tables[0].Columns) != 3 {
		t.Fatalf("users: got %d columns, want 3", len(tables[0].Columns))
	}
	if tables[0].Columns[0].Name != "name" {
		t.Errorf("users[0]: got %q, want %q", tables[0].Columns[0].Name, "name")
	}
	if tables[0].Columns[1].Name != "id" {
		t.Errorf("users[1]: got %q, want %q", tables[0].Columns[1].Name, "id")
	}

	// Check posts columns are sorted
	if len(tables[1].Columns) != 2 {
		t.Fatalf("posts: got %d columns, want 2", len(tables[1].Columns))
	}
	if tables[1].Columns[0].Name != "id" {
		t.Errorf("posts[0]: got %q, want %q", tables[1].Columns[0].Name, "id")
	}
	if tables[1].Columns[1].Name != "title" {
		t.Errorf("posts[1]: got %q, want %q", tables[1].Columns[1].Name, "title")
	}
}

func TestColumnIsNullable(t *testing.T) {
	tests := []struct {
		nullable string
		want     bool
	}{
		{"YES", true},
		{"yes", false},
		{"true", true},
		{"false", false},
		{"NO", false},
		{"", false},
	}

	for _, tt := range tests {
		c := &Column{Nullable: tt.nullable}
		got := c.IsNullable()
		if got != tt.want {
			t.Errorf("IsNullable(%q) = %v, want %v", tt.nullable, got, tt.want)
		}
	}
}

func TestColumnIsPrimaryKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"PRI", true},
		{"MUL", false},
		{"UNI", false},
		{"", false},
	}

	for _, tt := range tests {
		c := &Column{Key: tt.key}
		got := c.IsPrimaryKey()
		if got != tt.want {
			t.Errorf("IsPrimaryKey(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestColumnIsAutoIncrement(t *testing.T) {
	tests := []struct {
		extra string
		want  bool
	}{
		{"auto_increment", true},
		{"", false},
		{"DEFAULT_GENERATED", false},
	}

	for _, tt := range tests {
		c := &Column{Extra: tt.extra}
		got := c.IsAutoIncrement()
		if got != tt.want {
			t.Errorf("IsAutoIncrement(%q) = %v, want %v", tt.extra, got, tt.want)
		}
	}
}

func TestColumnIsUnsigned(t *testing.T) {
	tests := []struct {
		columnType string
		want       bool
	}{
		{"int(11) unsigned", true},
		{"bigint(20) unsigned", true},
		{"int(11)", false},
		{"varchar(255)", false},
		{"", false},
	}

	for _, tt := range tests {
		c := &Column{ColumnType: tt.columnType}
		got := c.IsUnsigned()
		if got != tt.want {
			t.Errorf("IsUnsigned(%q) = %v, want %v", tt.columnType, got, tt.want)
		}
	}
}
