package codegen

import (
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/sishui/bake/internal/config"
)

func TestNewStructField(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.CustomField
		want *Field
	}{
		{
			name: "basic field",
			cfg:  &config.CustomField{Name: "Theme", Type: "string"},
			want: &Field{
				Name:     "Theme",
				Type:     "string",
				Tag:      "`json:\"theme,omitempty\"`",
				Comments: nil,
				Imports:  nil,
			},
		},
		{
			name: "field with comment",
			cfg:  &config.CustomField{Name: "Count", Type: "int", Comment: "Number of items"},
			want: &Field{
				Name:     "Count",
				Type:     "int",
				Tag:      "`json:\"count,omitempty\"`",
				Comments: []string{"Number of items"},
				Imports:  nil,
			},
		},
		{
			name: "field with custom import",
			cfg:  &config.CustomField{Name: "Data", Type: "json.RawMessage", Import: "encoding/json"},
			want: &Field{
				Name:     "Data",
				Type:     "json.RawMessage",
				Tag:      "`json:\"data,omitempty\"`",
				Comments: nil,
				Imports:  []string{"encoding/json"},
			},
		},
		{
			name: "field with custom json tag name",
			cfg: &config.CustomField{
				Name: "RefreshInterval",
				Type: "int",
				Tags: []*config.Tag{
					{Key: "json", Name: "refresh_interval", Options: []string{"omitempty"}},
				},
			},
			want: &Field{
				Name: "RefreshInterval",
				Type: "int",
				Tag:  "`json:\"refresh_interval,omitempty\"`",
			},
		},
		{
			name: "field with custom json tag omitting omitempty",
			cfg: &config.CustomField{
				Name: "Required",
				Type: "string",
				Tags: []*config.Tag{
					{Key: "json", Name: "required"},
				},
			},
			want: &Field{
				Name: "Required",
				Type: "string",
				Tag:  "`json:\"required\"`",
			},
		},
		{
			name: "field with extra tags",
			cfg: &config.CustomField{
				Name: "Email",
				Type: "string",
				Tags: []*config.Tag{
					{Key: "yaml", Name: "email"},
					{Key: "form", Name: "Email"},
				},
			},
			want: &Field{
				Name: "Email",
				Type: "string",
				Tag:  "`form:\"Email\" json:\"email,omitempty\" yaml:\"email\"`",
			},
		},
		{
			name: "field with slice type",
			cfg:  &config.CustomField{Name: "Tags", Type: "[]string"},
			want: &Field{
				Name: "Tags",
				Type: "[]string",
				Tag:  "`json:\"tags,omitempty\"`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := newStructField(tt.cfg)
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Type != tt.want.Type {
				t.Errorf("GoType = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Tag != tt.want.Tag {
				t.Errorf("Tag = %q, want %q", got.Tag, tt.want.Tag)
			}
			if !reflect.DeepEqual(got.Comments, tt.want.Comments) {
				t.Errorf("Comment = %v, want %v", got.Comments, tt.want.Comments)
			}
			if !reflect.DeepEqual(got.Imports, tt.want.Imports) {
				t.Errorf("Imports = %v, want %v", got.Imports, tt.want.Imports)
			}
		})
	}
}

func TestNewCustomStruct(t *testing.T) {
	cfg := &config.Config{
		Version: "v0.2.0",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name:    "Config",
		Comment: "Application config",
		Fields: []*config.CustomField{
			{Name: "Theme", Type: "string"},
		},
	}

	data := NewCustomStruct(cfg, cs)

	if data.Version != "v0.2.0" {
		t.Errorf("Version = %q, want %q", data.Version, "v0.2.0")
	}
	if data.Package != "model" {
		t.Errorf("Package = %q, want %q", data.Package, "model")
	}
	if data.Module != "github.com/sishui/bake/examples/custom" {
		t.Errorf("Module = %q, want %q", data.Module, "github.com/sishui/bake/examples/custom")
	}
	if data.Model != "Config" {
		t.Errorf("Model = %q, want %q", data.Model, "Config")
	}
	if len(data.Fields) != 1 {
		t.Errorf("len(Fields) = %d, want 1", len(data.Fields))
	}

	// Check default imports are present
	if len(data.Imports) == 0 {
		t.Fatal("Imports is empty, expected at least default imports")
	}
}

func TestNewCustomStruct_WithCustomImport(t *testing.T) {
	cfg := &config.Config{
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name: "Event",
		Fields: []*config.CustomField{
			{
				Name:   "CreatedAt",
				Type:   "time.Time",
				Import: "time",
			},
		},
	}

	data := NewCustomStruct(cfg, cs)

	// Verify custom import appears in the grouped imports (non-default)
	found := false
	for _, group := range data.Imports {
		for _, imp := range group {
			if imp == "time" {
				found = true
			}
		}
	}
	if !found {
		t.Error("time import not found in grouped imports")
	}
}

func TestCustomStructTemplateRender_WithCustomImport(t *testing.T) {
	tmpl, err := parseTemplates("")
	if err != nil {
		t.Fatalf("parseTemplates() error = %v", err)
	}

	cfg := &config.Config{
		Version: "v0.2.0",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name: "Event",
		Fields: []*config.CustomField{
			{
				Name:    "CreatedAt",
				Type:    "time.Time",
				Import:  "time",
				Comment: "Event timestamp",
			},
			{Name: "Name", Type: "string"},
		},
	}

	data := NewCustomStruct(cfg, cs)
	buffer, err := tmpl.render("custom", data)
	if err != nil {
		t.Fatalf("tmpl.render() error = %v", err)
	}

	output := buffer.String()

	// Verify custom import appears in the generated imports block
	if !strings.Contains(output, "\"time\"") {
		t.Error("output missing time import")
	}

	// Verify the field uses the imported type
	if !strings.Contains(output, "CreatedAt time.Time `json:\"created_at,omitempty\"` // Event timestamp") {
		t.Error("output missing CreatedAt field with time.Time type")
	}

	// Verify generated code is valid Go
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", output, parser.ParseComments)
	if err != nil {
		t.Errorf("generated code is not valid Go: %v", err)
	}
}

func TestSortFields(t *testing.T) {
	fields := []*Field{
		{Name: "Zulu", Type: "string", Tag: "`json:\"zulu\"`"},
		{Name: "Alpha", Type: "int", Tag: "`json:\"alpha\"`"},
		{Name: "Beta", Type: "bool", Tag: "`json:\"beta\"`"},
	}

	sortFields(fields)

	if fields[0].Name != "Alpha" {
		t.Errorf("fields[0].Name = %q, want %q", fields[0].Name, "Alpha")
	}
	if fields[1].Name != "Beta" {
		t.Errorf("fields[1].Name = %q, want %q", fields[1].Name, "Beta")
	}
	if fields[2].Name != "Zulu" {
		t.Errorf("fields[2].Name = %q, want %q", fields[2].Name, "Zulu")
	}
}

func TestCustomStructTemplateRender(t *testing.T) {
	tmpl, err := parseTemplates("")
	if err != nil {
		t.Fatalf("parseTemplates() error = %v", err)
	}

	cfg := &config.Config{
		Version: "v0.2.0",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name:    "Config",
		Comment: "Application configuration",
		Fields: []*config.CustomField{
			{Name: "Theme", Type: "string", Comment: "UI theme"},
			{Name: "Enabled", Type: "bool"},
		},
	}

	data := NewCustomStruct(cfg, cs)
	buffer, err := tmpl.render("custom", data)
	if err != nil {
		t.Fatalf("tmpl.render() error = %v", err)
	}

	output := buffer.String()

	// Check generated file header
	if !strings.Contains(output, "Code generated by bake. DO NOT EDIT.") {
		t.Error("output missing 'Code generated by bake' header")
	}
	if !strings.Contains(output, "version: v0.2.0") {
		t.Error("output missing version")
	}

	// Check package declaration
	if !strings.Contains(output, "package model") {
		t.Error("output missing package declaration")
	}

	// Check imports
	if !strings.Contains(output, "\"database/sql/driver\"") {
		t.Error("output missing driver import")
	}
	if !strings.Contains(output, "\"encoding/json\"") {
		t.Error("output missing json import")
	}
	if !strings.Contains(output, "\"fmt\"") {
		t.Error("output missing fmt import")
	}

	// Check Config struct
	if !strings.Contains(output, "// Application configuration") {
		t.Error("output missing Config comment")
	}
	if !strings.Contains(output, "type Config struct {") {
		t.Error("output missing Config struct definition")
	}
	// With field alignment, Theme (5) padded to maxNameLen=7, string is maxTypeLen=6.
	themeLine := "Theme   string `json:\"theme,omitempty\"`   // UI theme"
	if !strings.Contains(output, themeLine) {
		t.Errorf("output missing Theme field with alignment\n  want: %q\n  got:  %q\n", themeLine, output)
	}
	// Enabled (7) is max name width; bool (4) padded to maxTypeLen=6.
	enabledLine := "Enabled bool   `json:\"enabled,omitempty\"`"
	if !strings.Contains(output, enabledLine) {
		t.Errorf("output missing Enabled field with alignment\n  want: %q\n", enabledLine)
	}

	// Check Scan/Value methods
	if !strings.Contains(output, "func (o *Config) Scan(src any) error {") {
		t.Error("output missing Config Scan method")
	}
	if !strings.Contains(output, "func (o Config) Value() (driver.Value, error) {") {
		t.Error("output missing Config Value method")
	}

	// Verify fields are sorted alphabetically (Enabled < Theme)
	themeIdx := strings.Index(output, "Theme   string ")
	enabledIdx := strings.Index(output, "Enabled bool   ")
	if enabledIdx > themeIdx {
		t.Error("fields not sorted: Enabled should appear before Theme")
	}

	// Verify generated code is valid Go
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", output, parser.ParseComments)
	if err != nil {
		t.Errorf("generated code is not valid Go: %v", err)
	}
}

func TestCustomStructTemplateRender_MultiLineComment(t *testing.T) {
	tmpl, err := parseTemplates("")
	if err != nil {
		t.Fatalf("parseTemplates() error = %v", err)
	}

	cfg := &config.Config{
		Version: "v0.2.0",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name: "Config",
		Fields: []*config.CustomField{
			{Name: "Simple", Type: "string", Comment: "Single line"},
			{Name: "Complex", Type: "int", Comment: "First line\nSecond line\nThird line"},
			{Name: "Normal", Type: "bool"},
		},
	}

	data := NewCustomStruct(cfg, cs)
	buffer, err := tmpl.render("custom", data)
	if err != nil {
		t.Fatalf("tmpl.render() error = %v", err)
	}

	output := strings.ReplaceAll(buffer.String(), "\r\n", "\n")

	// Multi-line comment should render each comment line above the field
	if !strings.Contains(output, "// First line") {
		t.Error("output missing '// First line' comment")
	}
	if !strings.Contains(output, "// Second line") {
		t.Error("output missing '// Second line' comment")
	}
	if !strings.Contains(output, "// Third line") {
		t.Error("output missing '// Third line' comment")
	}

	// Simple (single-line comment): group 1 maxName=6, maxType=6, maxTag=27
	if !strings.Contains(output, "Simple  string `json:\"simple,omitempty\"` // Single line") {
		t.Error("output missing Simple field with single-line comment")
	}

	// Complex (multi-line comment): comments above field, uses raw tag
	if !strings.Contains(output, "// First line\n\t// Second line\n\t// Third line\n\tComplex int    `json:\"complex,omitempty\"`") {
		t.Error("output missing Complex field with multi-line comment")
	}

	// Normal: in same group as Complex, maxName=7, maxType=4 → padded
	if !strings.Contains(output, "Normal  bool   `json:\"normal,omitempty\"` //") {
		t.Error("output missing Normal field")
	}

	// Verify generated code is valid Go
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", output, parser.ParseComments)
	if err != nil {
		t.Errorf("generated code is not valid Go: %v", err)
	}
}

func BenchmarkNewCustomStruct(b *testing.B) {
	cfg := &config.Config{
		Version: "v0.2.0",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name: "BenchStruct",
		Fields: []*config.CustomField{
			{Name: "ID", Type: "int", Comment: "Unique identifier"},
			{Name: "Name", Type: "string", Comment: "Display name"},
			{Name: "Active", Type: "bool"},
			{Name: "Tags", Type: "[]string", Comment: "Associated tags"},
			{Name: "Score", Type: "float64"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		data := NewCustomStruct(cfg, cs)
		if len(data.Fields) == 0 {
			b.Fatal("Fields is empty")
		}
	}
}

func BenchmarkCustomStructTemplateRender(b *testing.B) {
	tmpl, err := parseTemplates("")
	if err != nil {
		b.Fatalf("parseTemplates() error = %v", err)
	}

	cfg := &config.Config{
		Version: "v0.2.0",
		Output: &config.Output{
			Package: "model",
			Module:  "github.com/sishui/bake/examples/custom",
		},
	}
	cs := &config.CustomStruct{
		Name: "BenchStruct",
		Fields: []*config.CustomField{
			{Name: "ID", Type: "int", Comment: "Unique identifier"},
			{Name: "Name", Type: "string", Comment: "Display name"},
			{Name: "Active", Type: "bool"},
			{Name: "Tags", Type: "[]string", Comment: "Associated tags"},
			{Name: "Score", Type: "float64"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		data := NewCustomStruct(cfg, cs)
		buf, err := tmpl.render("custom", data)
		if err != nil {
			b.Fatalf("render() error = %v", err)
		}
		if buf.Len() == 0 {
			b.Fatal("empty output")
		}
	}
}

func TestStructField_EmptyTagsField(t *testing.T) {
	// Verify that field-level tags being empty doesn't cause issues
	cfg := &config.CustomField{
		Name: "Email",
		Type: "string",
		Tags: []*config.Tag{
			{Key: "", Name: "", Options: nil}, // empty/invalid tag
		},
	}

	field := newStructField(cfg)
	if field.Name != "Email" {
		t.Errorf("Name = %q, want %q", field.Name, "Email")
	}
	if field.Tag != "`json:\"email,omitempty\"`" {
		t.Errorf("Tag = %q, want %q", field.Tag, "`json:\"email,omitempty\"`")
	}
}
