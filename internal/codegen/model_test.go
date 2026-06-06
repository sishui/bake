package generate

import (
	"strings"
	"testing"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
)

func TestNewField(t *testing.T) {
	n := naming.New("ID", "URL")
	tests := []struct {
		name       string
		column     schema.Column
		custom     *config.CustomTable
		driver     string
		wantName   string
		wantType   string
		wantKind   string
		wantPri    bool
		wantNull   bool
		wantErr    bool
	}{
		{
			name:     "int primary key",
			column:   schema.Column{Name: "id", DataType: "int", ColumnType: "int(11)", Key: schema.ColumnKeyPrimary, Nullable: "NO", Extra: schema.ColumnExtraAutoIncrement},
			driver:   "mysql",
			wantName: "ID",
			wantType: "int32",
			wantKind: "NUMERIC",
			wantPri:  true,
		},
		{
			name:     "nullable varchar",
			column:   schema.Column{Name: "name", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "YES"},
			driver:   "mysql",
			wantName: "Name",
			wantType: "*string",
			wantKind: "STRING",
			wantNull: true,
		},
		{
			name:     "datetime",
			column:   schema.Column{Name: "created_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO"},
			driver:   "mysql",
			wantName: "CreatedAt",
			wantType: "time.Time",
			wantKind: "TIME",
		},
		{
			name:     "postgres uuid",
			column:   schema.Column{Name: "uuid_col", DataType: "uuid", ColumnType: "uuid", Nullable: "YES"},
			driver:   "postgres",
			wantName: "UuidCol",
			wantType: "*uuid.UUID",
			wantKind: "STRUCT",
			wantNull: true,
		},
		{
			name:     "column with comment",
			column:   schema.Column{Name: "title", DataType: "varchar", ColumnType: "varchar(100)", Nullable: "NO", Comment: "post title"},
			driver:   "mysql",
			wantName: "Title",
			wantType: "string",
			wantKind: "STRING",
		},
		{
			name:    "unsupported type returns error",
			column:  schema.Column{Name: "col", DataType: "unknown_type", ColumnType: "unknown_type"},
			driver:  "mysql",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewField(&tt.column, tt.custom, tt.driver, n)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewField() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Kind != tt.wantKind {
				t.Errorf("Kind = %q, want %q", got.Kind, tt.wantKind)
			}
			if got.IsPrimary != tt.wantPri {
				t.Errorf("IsPrimary = %v, want %v", got.IsPrimary, tt.wantPri)
			}
			if got.IsNullable != tt.wantNull {
				t.Errorf("IsNullable = %v, want %v", got.IsNullable, tt.wantNull)
			}
			if got.ColumnName != tt.column.Name {
				t.Errorf("ColumnName = %q, want %q", got.ColumnName, tt.column.Name)
			}
		})
	}
}

func TestNewField_WithCustomField(t *testing.T) {
	n := naming.New("ID")
	column := &schema.Column{Name: "status", DataType: "varchar", ColumnType: "varchar(20)", Nullable: "NO"}
	customTable := &config.CustomTable{
		Fields: map[string]*config.CustomField{
			"status": {
				Name: "State",
				Type: "int32",
			},
		},
	}

	got, err := NewField(column, customTable, "mysql", n)
	if err != nil {
		t.Fatalf("NewField() error = %v", err)
	}
	if got.Name != "State" {
		t.Errorf("Name = %q, want %q", got.Name, "State")
	}
	if got.Type != "int32" {
		t.Errorf("Type = %q, want %q", got.Type, "int32")
	}
}

func TestNewField_WithCustomTags(t *testing.T) {
	n := naming.New()
	column := &schema.Column{Name: "email", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "NO"}
	customTable := &config.CustomTable{
		Tags: []*config.Tag{
			{Key: "yaml", Name: "$SnakeCase"},
		},
		Fields: map[string]*config.CustomField{
			"email": {
				Tags: []*config.Tag{
					{Key: "validate", Name: "required,email"},
				},
			},
		},
	}

	got, err := NewField(column, customTable, "mysql", n)
	if err != nil {
		t.Fatalf("NewField() error = %v", err)
	}
	if !strings.Contains(got.Tag, `validate:"required,email"`) {
		t.Errorf("Tag = %q, want to contain validate tag", got.Tag)
	}
	if !strings.Contains(got.Tag, `yaml:"email"`) {
		t.Errorf("Tag = %q, want to contain yaml tag", got.Tag)
	}
}

func TestNewField_Imports(t *testing.T) {
	n := naming.New()
	tests := []struct {
		name        string
		column      schema.Column
		driver      string
		wantContain string // substring that should appear in imports
	}{
		{
			name:        "time column imports time",
			column:      schema.Column{Name: "created_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO"},
			driver:      "mysql",
			wantContain: "time",
		},
		{
			name:        "decimal column imports shopspring",
			column:      schema.Column{Name: "price", DataType: "decimal", ColumnType: "decimal(10,2)", Nullable: "NO"},
			driver:      "mysql",
			wantContain: "github.com/shopspring/decimal",
		},
		{
			name:        "json column imports encoding/json",
			column:      schema.Column{Name: "meta", DataType: "json", ColumnType: "json", Nullable: "NO"},
			driver:      "mysql",
			wantContain: "encoding/json",
		},
		{
			name:        "postgres uuid imports google/uuid",
			column:      schema.Column{Name: "id", DataType: "uuid", ColumnType: "uuid", Nullable: "NO"},
			driver:      "postgres",
			wantContain: "github.com/google/uuid",
		},
		{
			name:        "postgres inet imports net",
			column:      schema.Column{Name: "ip", DataType: "inet", ColumnType: "inet", Nullable: "NO"},
			driver:      "postgres",
			wantContain: "net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewField(&tt.column, nil, tt.driver, n)
			if err != nil {
				t.Fatalf("NewField() error = %v", err)
			}
			found := false
			for _, imp := range got.Imports {
				if imp == tt.wantContain {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Imports = %v, want to contain %q", got.Imports, tt.wantContain)
			}
		})
	}
}

func TestNewField_Comments(t *testing.T) {
	n := naming.New()
	tests := []struct {
		name         string
		column       schema.Column
		wantComments []string
	}{
		{
			name:         "single line comment",
			column:       schema.Column{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Comment: "primary key"},
			wantComments: []string{"primary key"},
		},
		{
			name:         "multi line comment",
			column:       schema.Column{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Comment: "line1\nline2"},
			wantComments: []string{"line1", "line2"},
		},
		{
			name:         "empty comment",
			column:       schema.Column{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO"},
			wantComments: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewField(&tt.column, nil, "mysql", n)
			if err != nil {
				t.Fatalf("NewField() error = %v", err)
			}
			if len(got.Comments) != len(tt.wantComments) {
				t.Fatalf("Comments = %v, want %v", got.Comments, tt.wantComments)
			}
			for i, c := range got.Comments {
				if c != tt.wantComments[i] {
					t.Errorf("Comments[%d] = %q, want %q", i, c, tt.wantComments[i])
				}
			}
		})
	}
}

func TestNewModel(t *testing.T) {
	n := naming.New("ID")
	cfg := &config.Config{
		Version:  "v0.2.1",
		Timezone: "",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, Extra: schema.ColumnExtraAutoIncrement, OrdinalPosition: 1},
			{Name: "name", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "NO", OrdinalPosition: 2},
			{Name: "email", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "YES", OrdinalPosition: 3},
		},
	}

	m, err := NewModel(table, db, cfg, n)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Check basic model attributes
	if m.Table != "users" {
		t.Errorf("Table = %q, want %q", m.Table, "users")
	}
	if m.Model != "User" {
		t.Errorf("Model = %q, want %q", m.Model, "User")
	}
	if m.Alias != "users_alias" {
		t.Errorf("Alias = %q, want %q", m.Alias, "users_alias")
	}
	if m.Version != "v0.2.1" {
		t.Errorf("Version = %q, want %q", m.Version, "v0.2.1")
	}
	if m.Package != "model" {
		t.Errorf("Package = %q, want %q", m.Package, "model")
	}

	// Check fields count (3 columns, no foreign keys)
	if len(m.Fields) != 3 {
		t.Fatalf("len(Fields) = %d, want 3", len(m.Fields))
	}

	// Check first field (id)
	idField := m.Fields[0]
	if idField.Name != "ID" {
		t.Errorf("Fields[0].Name = %q, want %q", idField.Name, "ID")
	}
	if idField.Type != "int32" {
		t.Errorf("Fields[0].Type = %q, want %q", idField.Type, "int32")
	}
	if !idField.IsPrimary {
		t.Error("Fields[0].IsPrimary = false, want true")
	}
	if idField.Kind != "NUMERIC" {
		t.Errorf("Fields[0].Kind = %q, want %q", idField.Kind, "NUMERIC")
	}

	// Check nullable field (email)
	emailField := m.Fields[2]
	if !emailField.IsNullable {
		t.Error("Fields[2].IsNullable = false, want true")
	}

	// Check imports contain expected packages
	importStr := strings.Join(flattenImports(m.Imports), ",")
	if !strings.Contains(importStr, "context") {
		t.Error("Imports missing context")
	}
	if !strings.Contains(importStr, "github.com/uptrace/bun") {
		t.Error("Imports missing bun")
	}
}

func TestNewModel_WithForeignKey(t *testing.T) {
	n := naming.New()
	cfg := &config.Config{
		Version: "v0.2.1",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}

	postsTable := &schema.Table{
		Name: "posts",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, OrdinalPosition: 1},
			{Name: "user_id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", OrdinalPosition: 2},
			{Name: "title", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "NO", OrdinalPosition: 3},
		},
		ForeignKeys: []schema.ForeignKey{
			{ConstraintName: "fk_posts_user", Table: "posts", ColumnName: "user_id", RefTable: "users", RefColumn: "id"},
		},
	}

	m, err := NewModel(postsTable, db, cfg, n)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// 3 columns + 1 belongs-to relation
	if len(m.Fields) != 4 {
		t.Fatalf("len(Fields) = %d, want 4", len(m.Fields))
	}

	// Last field should be the belongs-to relation
	relation := m.Fields[3]
	if !relation.IsRelation {
		t.Error("Fields[3].IsRelation = false, want true")
	}
	if !relation.IsCustom {
		t.Error("Fields[3].IsCustom = false, want true")
	}
	if relation.Kind != "STRUCT" {
		t.Errorf("Fields[3].Kind = %q, want %q", relation.Kind, "STRUCT")
	}
}

func TestNewModel_WithReverseForeignKey(t *testing.T) {
	n := naming.New()
	cfg := &config.Config{
		Version: "v0.2.1",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}

	usersTable := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, OrdinalPosition: 1},
			{Name: "name", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "NO", OrdinalPosition: 2},
		},
		ReverseForeignKeys: []schema.ForeignKey{
			{ConstraintName: "fk_posts_user", Table: "posts", ColumnName: "user_id", RefTable: "users", RefColumn: "id"},
		},
	}

	m, err := NewModel(usersTable, db, cfg, n)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// 2 columns + 1 has-many relation
	if len(m.Fields) != 3 {
		t.Fatalf("len(Fields) = %d, want 3", len(m.Fields))
	}

	relation := m.Fields[2]
	if !relation.IsRelation {
		t.Error("Fields[2].IsRelation = false, want true")
	}
	if !strings.HasPrefix(relation.Type, "[]*") {
		t.Errorf("Fields[2].Type = %q, want slice-of-pointers type", relation.Type)
	}
}

func TestNewModel_WithCustomTable(t *testing.T) {
	n := naming.New()
	cfg := &config.Config{
		Version: "v0.2.1",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
		Custom: map[string]*config.CustomTable{
			"users": {
				Name: "app_users",
				Tags: []*config.Tag{
					{Key: "yaml", Name: "$SnakeCase"},
				},
				Fields: map[string]*config.CustomField{
					"status": {
						Name: "State",
						Type: "int32",
					},
				},
			},
		},
	}
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, OrdinalPosition: 1},
			{Name: "status", DataType: "varchar", ColumnType: "varchar(20)", Nullable: "NO", OrdinalPosition: 2},
		},
	}

	m, err := NewModel(table, db, cfg, n)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Custom name should override auto-generated
	if m.Model != "AppUser" {
		t.Errorf("Model = %q, want %q", m.Model, "AppUser")
	}

	// Custom field should override column type
	statusField := m.Fields[1]
	if statusField.Name != "State" {
		t.Errorf("Fields[1].Name = %q, want %q", statusField.Name, "State")
	}
	if statusField.Type != "int32" {
		t.Errorf("Fields[1].Type = %q, want %q", statusField.Type, "int32")
	}
}

func TestNewModel_MaxLengths(t *testing.T) {
	n := naming.New()
	cfg := &config.Config{
		Version: "v0.2.1",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}
	table := &schema.Table{
		Name: "orders",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, OrdinalPosition: 1},
			{Name: "user_id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", OrdinalPosition: 2},
			{Name: "total_price", DataType: "decimal", ColumnType: "decimal(10,2)", Nullable: "NO", OrdinalPosition: 3},
			{Name: "notes", DataType: "text", ColumnType: "text", Nullable: "YES", OrdinalPosition: 4},
			{Name: "created_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO", OrdinalPosition: 5},
			{Name: "deleted_at", DataType: "datetime", ColumnType: "datetime", Nullable: "YES", OrdinalPosition: 6},
		},
	}

	m, err := NewModel(table, db, cfg, n)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// MaxFieldLength should be the longest field name
	if m.MaxFieldLength < 10 { // "TotalPrice" = 10
		t.Errorf("MaxFieldLength = %d, want >= 10", m.MaxFieldLength)
	}

	// MaxNullableLength should track nullable fields (notes=5, deleted_at=10)
	if m.MaxNullableLength < 5 { // "Notes" = 5
		t.Errorf("MaxNullableLength = %d, want >= 5", m.MaxNullableLength)
	}

	// MaxStringLength should track string fields (notes=5)
	if m.MaxStringLength < 5 {
		t.Errorf("MaxStringLength = %d, want >= 5", m.MaxStringLength)
	}

	// MaxNumericLength should track numeric fields (id=2, user_id=6, total_price=10)
	if m.MaxNumericLength < 10 {
		t.Errorf("MaxNumericLength = %d, want >= 10", m.MaxNumericLength)
	}

	// MaxTimeLength should track time fields
	if m.MaxTimeLength < 9 {
		t.Errorf("MaxTimeLength = %d, want >= 9", m.MaxTimeLength)
	}

	// MaxArithmeticLength should track non-pk numeric (user_id=6, total_price=10)
	if m.MaxArithmeticLength < 10 {
		t.Errorf("MaxArithmeticLength = %d, want >= 10", m.MaxArithmeticLength)
	}
}

func TestNewModel_TimeHooks(t *testing.T) {
	n := naming.New()
	cfg := &config.Config{
		Version:  "v0.2.1",
		Timezone: "UTC",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}
	table := &schema.Table{
		Name: "logs",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, OrdinalPosition: 1},
			{Name: "created_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO", OrdinalPosition: 2},
			{Name: "updated_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO", OrdinalPosition: 3},
			{Name: "deleted_at", DataType: "datetime", ColumnType: "datetime", Nullable: "YES", OrdinalPosition: 4},
		},
	}

	m, err := NewModel(table, db, cfg, n)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	if m.CreatedAtType != "time.Time" {
		t.Errorf("CreatedAtType = %q, want %q", m.CreatedAtType, "time.Time")
	}
	if m.UpdatedAtType != "time.Time" {
		t.Errorf("UpdatedAtType = %q, want %q", m.UpdatedAtType, "time.Time")
	}
	// deleted_at is nullable so type is *time.Time
	if m.DeletedAtType != "*time.Time" {
		t.Errorf("DeletedAtType = %q, want %q", m.DeletedAtType, "*time.Time")
	}
}

func TestNewModel_UnsupportedType(t *testing.T) {
	n := naming.New()
	cfg := &config.Config{
		Version: "v0.2.1",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}
	table := &schema.Table{
		Name: "bad_table",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, OrdinalPosition: 1},
			{Name: "geo", DataType: "unknown_type", ColumnType: "unknown_type", Nullable: "NO", OrdinalPosition: 2},
		},
	}

	_, err := NewModel(table, db, cfg, n)
	if err == nil {
		t.Fatal("NewModel() should return error for unsupported type")
	}
}

func BenchmarkNewModel(b *testing.B) {
	n := naming.New("ID", "URL")
	cfg := &config.Config{
		Version: "v0.2.1",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake",
		},
	}
	db := &config.DB{
		Driver: "mysql",
		DSN:    "root:@tcp(127.0.0.1:3306)/test",
	}
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int", ColumnType: "int(11)", Nullable: "NO", Key: schema.ColumnKeyPrimary, Extra: schema.ColumnExtraAutoIncrement, OrdinalPosition: 1},
			{Name: "name", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "NO", OrdinalPosition: 2},
			{Name: "email", DataType: "varchar", ColumnType: "varchar(255)", Nullable: "YES", OrdinalPosition: 3},
			{Name: "age", DataType: "int", ColumnType: "int(11)", Nullable: "NO", OrdinalPosition: 4},
			{Name: "created_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO", OrdinalPosition: 5},
			{Name: "updated_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO", OrdinalPosition: 6},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		m, err := NewModel(table, db, cfg, n)
		if err != nil {
			b.Fatal(err)
		}
		if len(m.Fields) == 0 {
			b.Fatal("Fields is empty")
		}
	}
}

// flattenImports flattens grouped imports into a single slice.
func flattenImports(groups [][]string) []string {
	var result []string
	for _, g := range groups {
		result = append(result, g...)
	}
	return result
}
