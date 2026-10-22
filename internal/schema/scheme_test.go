package schema

import (
	"context"
	"reflect"
	"slices"
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
			t.Parallel()
			cfg := &config.DB{
				Exclude: tt.excluded,
				Include: tt.included,
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

	assignColumns(tables, columns, nil)

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
		// MySQL values
		{"YES", true},
		{"NO", false},
		{"yes", true},
		{"no", false},
		// PostgreSQL values (t/f)
		{"t", true},
		{"f", false},
		{"true", true},
		{"false", false},
		{"", false},
		{"anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.nullable, func(t *testing.T) {
			c := &Column{Nullable: tt.nullable}
			got := c.IsNullable()
			if got != tt.want {
				t.Errorf("IsNullable(%q) = %v, want %v", tt.nullable, got, tt.want)
			}
		})
	}
}

func TestColumnIsPrimaryKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{ColumnKeyPrimary, true},
		{ColumnKeyMulti, false},
		{ColumnKeyUnique, false},
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
		{ColumnExtraAutoIncrement, true},
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

func TestAssignColumns_CompositeUniqueIndex(t *testing.T) {
	tables := []*Table{
		{Name: "bookings"},
	}
	columns := map[string][]*Column{
		"bookings": {
			{Name: "id", OrdinalPosition: 1, Key: ColumnKeyPrimary},
			{Name: "start_at", OrdinalPosition: 2, Key: ColumnKeyMulti},
			{Name: "end_at", OrdinalPosition: 3},
		},
	}
	indexes := []Index{
		{Table: "bookings", NonUnique: 0, IndexName: "idx_unique_range", ColumnName: "start_at"},
		{Table: "bookings", NonUnique: 0, IndexName: "idx_unique_range", ColumnName: "end_at"},
	}

	assignColumns(tables, columns, indexes)

	startAt := tables[0].Columns[1]
	if !startAt.IsUnique() {
		t.Error("start_at IsUnique() = false, want true")
	}
	if !startAt.IsMultiKey {
		t.Error("start_at IsMultiKey = false, want true")
	}

	endAt := tables[0].Columns[2]
	if !endAt.IsUnique() {
		t.Error("end_at IsUnique() = false, want true")
	}
	if !endAt.IsMultiKey {
		t.Error("end_at IsMultiKey = false, want true")
	}
}

func TestAssignColumns_SingleUniqueIndex(t *testing.T) {
	tables := []*Table{
		{Name: "users"},
	}
	columns := map[string][]*Column{
		"users": {
			{Name: "id", OrdinalPosition: 1, Key: ColumnKeyPrimary},
			{Name: "email", OrdinalPosition: 2, Key: ColumnKeyUnique},
		},
	}
	indexes := []Index{
		{Table: "users", NonUnique: 0, IndexName: "idx_email", ColumnName: "email"},
	}

	assignColumns(tables, columns, indexes)

	email := tables[0].Columns[1]
	if !email.IsUnique() {
		t.Error("email IsUnique() = false, want true")
	}
	if email.IsMultiKey {
		t.Error("email IsMultiKey = true, want false")
	}
}

func TestAssignColumns_PreferUniqueIndex(t *testing.T) {
	tables := []*Table{
		{Name: "events"},
	}
	columns := map[string][]*Column{
		"events": {
			{Name: "id", OrdinalPosition: 1, Key: ColumnKeyPrimary},
			{Name: "code", OrdinalPosition: 2, Key: ColumnKeyMulti},
			{Name: "name", OrdinalPosition: 3},
		},
	}
	indexes := []Index{
		{Table: "events", NonUnique: 1, IndexName: "idx_code", ColumnName: "code"},
		{Table: "events", NonUnique: 0, IndexName: "idx_unique_code_name", ColumnName: "code"},
		{Table: "events", NonUnique: 0, IndexName: "idx_unique_code_name", ColumnName: "name"},
	}

	assignColumns(tables, columns, indexes)

	code := tables[0].Columns[1]
	if !code.IsUnique() {
		t.Error("code IsUnique() = false, want true")
	}
	if !code.IsMultiKey {
		t.Error("code IsMultiKey = false, want true")
	}

	name := tables[0].Columns[2]
	if !name.IsUnique() {
		t.Error("name IsUnique() = false, want true")
	}
	if !name.IsMultiKey {
		t.Error("name IsMultiKey = false, want true")
	}
}

func TestColumnIndexes(t *testing.T) {
	t.Run("EmptyInput_ReturnsEmptyMap", func(t *testing.T) {
		result := columnIndexes(nil)
		if len(result) != 0 {
			t.Errorf("got %d entries, want 0", len(result))
		}
	})

	t.Run("MultipleIndexesForSameColumn_Grouped", func(t *testing.T) {
		indexes := []Index{
			{Table: "users", ColumnName: "email", IndexName: "idx_email", NonUnique: 0},
			{Table: "users", ColumnName: "email", IndexName: "idx_email_2", NonUnique: 1},
		}
		result := columnIndexes(indexes)
		key := tableColumn{table: "users", col: "email"}
		if len(result[key]) != 2 {
			t.Errorf("got %d indexes for users.email, want 2", len(result[key]))
		}
	})

	t.Run("DifferentTablesAndColumns_SeparateKeys", func(t *testing.T) {
		indexes := []Index{
			{Table: "users", ColumnName: "email", IndexName: "idx_email", NonUnique: 0},
			{Table: "posts", ColumnName: "title", IndexName: "idx_title", NonUnique: 1},
		}
		result := columnIndexes(indexes)
		if len(result) != 2 {
			t.Errorf("got %d entries, want 2", len(result))
		}
		usersKey := tableColumn{table: "users", col: "email"}
		postsKey := tableColumn{table: "posts", col: "title"}
		if len(result[usersKey]) != 1 {
			t.Errorf("users.email: got %d indexes, want 1", len(result[usersKey]))
		}
		if len(result[postsKey]) != 1 {
			t.Errorf("posts.title: got %d indexes, want 1", len(result[postsKey]))
		}
	})
}

func TestBestIndexForColumn(t *testing.T) {
	t.Run("NoMatches_ReturnsFalse", func(t *testing.T) {
		colIndexes := map[tableColumn][]Index{}
		_, ok := bestIndexForColumn(colIndexes, "users", "email")
		if ok {
			t.Errorf("got ok=true, want false")
		}
	})

	t.Run("PrimaryOnly_ReturnsFalse", func(t *testing.T) {
		colIndexes := map[tableColumn][]Index{
			{table: "users", col: "id"}: {
				{Table: "users", ColumnName: "id", IndexName: IndexNamePrimary, NonUnique: 0},
			},
		}
		_, ok := bestIndexForColumn(colIndexes, "users", "id")
		if ok {
			t.Errorf("got ok=true for PRIMARY-only, want false")
		}
	})

	t.Run("UniqueIndex_Returned", func(t *testing.T) {
		want := Index{Table: "users", ColumnName: "email", IndexName: "idx_email", NonUnique: 0}
		colIndexes := map[tableColumn][]Index{
			{table: "users", col: "email"}: {want},
		}
		got, ok := bestIndexForColumn(colIndexes, "users", "email")
		if !ok {
			t.Fatalf("got ok=false, want true")
		}
		if got.IndexName != want.IndexName {
			t.Errorf("IndexName = %q, want %q", got.IndexName, want.IndexName)
		}
	})

	t.Run("MultipleIndexes_FirstUniqueNonPrimaryReturned", func(t *testing.T) {
		colIndexes := map[tableColumn][]Index{
			{table: "users", col: "id"}: {
				{Table: "users", ColumnName: "id", IndexName: IndexNamePrimary, NonUnique: 0},
				{Table: "users", ColumnName: "id", IndexName: "idx_id_unique", NonUnique: 0},
				{Table: "users", ColumnName: "id", IndexName: "idx_id_dup", NonUnique: 1},
			},
		}
		got, ok := bestIndexForColumn(colIndexes, "users", "id")
		if !ok {
			t.Fatalf("got ok=false, want true")
		}
		if got.IndexName != "idx_id_unique" {
			t.Errorf("IndexName = %q, want %q", got.IndexName, "idx_id_unique")
		}
	})

	t.Run("OnlyNonUnique_ReturnsFalse", func(t *testing.T) {
		colIndexes := map[tableColumn][]Index{
			{table: "posts", col: "title"}: {
				{Table: "posts", ColumnName: "title", IndexName: "idx_title", NonUnique: 1},
			},
		}
		_, ok := bestIndexForColumn(colIndexes, "posts", "title")
		if ok {
			t.Errorf("got ok=true for non-unique only, want false")
		}
	})
}

func TestColumnIsUnique(t *testing.T) {
	tests := []struct {
		name     string
		column   Column
		expected bool
	}{
		{
			name:     "EmptyIndexName_ReturnsFalse",
			column:   Column{IndexName: "", NonUnique: 0},
			expected: false,
		},
		{
			name:     "NonUniqueZeroWithIndexName_ReturnsTrue",
			column:   Column{IndexName: "idx_email", NonUnique: 0},
			expected: true,
		},
		{
			name:     "NonUniqueOne_ReturnsFalse",
			column:   Column{IndexName: "idx_title", NonUnique: 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.IsUnique()
			if got != tt.expected {
				t.Errorf("IsUnique() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAssignIndexes(t *testing.T) {
	tables := []*Table{
		{Name: "users", Columns: []*Column{
			{Name: "id", Key: ColumnKeyPrimary},
			{Name: "email"},
			{Name: "a"},
			{Name: "b"},
		}},
		{Name: "posts", Columns: []*Column{
			{Name: "id", Key: ColumnKeyPrimary},
			{Name: "title"},
		}},
	}
	indexes := []Index{
		{Table: "users", ColumnName: "id", IndexName: IndexNamePrimary, NonUnique: 0},
		{Table: "users", ColumnName: "id", IndexName: "users_pkey", NonUnique: 0},
		{Table: "users", ColumnName: "email", IndexName: "idx_users_email", NonUnique: 0},
		{Table: "users", ColumnName: "a", IndexName: "idx_users_a_b", NonUnique: 1},
		{Table: "users", ColumnName: "b", IndexName: "idx_users_a_b", NonUnique: 1},
		{Table: "posts", ColumnName: "title", IndexName: "idx_posts_title", NonUnique: 1},
	}

	assignIndexes(tables, indexes)

	if want := []string{"idx_users_a_b", "idx_users_email"}; !slices.Equal(tables[0].Indexes, want) {
		t.Errorf("users.Indexes = %v, want %v", tables[0].Indexes, want)
	}
	if want := []string{"idx_posts_title"}; !slices.Equal(tables[1].Indexes, want) {
		t.Errorf("posts.Indexes = %v, want %v", tables[1].Indexes, want)
	}
}

func TestAssignUniqueKeys(t *testing.T) {
	tables := []*Table{{Name: "users"}}
	indexes := []Index{
		{Table: "users", ColumnName: "id", IndexName: "users_pkey", NonUnique: 0},
		{Table: "users", ColumnName: "tenant_id", IndexName: "uk_tenant_email", NonUnique: 0},
		{Table: "users", ColumnName: "email", IndexName: "uk_tenant_email", NonUnique: 0},
		{Table: "users", ColumnName: "name", IndexName: "idx_name", NonUnique: 1},
	}

	assignIndexes(tables, indexes)

	want := []UniqueKey{
		{Columns: []string{"id"}},
		{Columns: []string{"tenant_id", "email"}},
	}
	if !reflect.DeepEqual(tables[0].Keys, want) {
		t.Errorf("users.Keys = %v, want %v", tables[0].Keys, want)
	}
}

func TestAssignUniqueKeys_DeduplicatesSameColumns(t *testing.T) {
	tables := []*Table{{Name: "users"}}
	indexes := []Index{
		{Table: "users", ColumnName: "id", IndexName: "users_pkey", NonUnique: 0},
		{Table: "users", ColumnName: "id", IndexName: "users_id_key", NonUnique: 0},
	}

	assignIndexes(tables, indexes)

	if len(tables[0].Keys) != 1 {
		t.Fatalf("users.Keys = %v, want a single key", tables[0].Keys)
	}
}
