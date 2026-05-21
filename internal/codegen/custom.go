package generate

import (
	"sort"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
)

// CustomStruct holds data for rendering a single custom struct template.
type CustomStruct struct {
	Version string
	Package string
	Module  string
	Imports [][]string
	Name    string
	Comment []string
	Fields  []*StructField
}

// StructField represents a field in a custom struct.
type StructField struct {
	Name        string   // Go field name (PascalCase)
	AlignedName string   // Name padded to max field name width
	GoType      string   // Go type
	AlignedType string   // GoType padded to max type width
	Tag         string   // Struct tag string (including backticks)
	AlignedTag  string   // Tag padded to max tag width
	Comment     []string // Field comment lines
	Imports     []string // Additional imports needed by this field's type
}

func (f *StructField) FieldName() string       { return f.Name }
func (f *StructField) FieldType() string       { return f.GoType }
func (f *StructField) GetTag() string          { return f.Tag }
func (f *StructField) GetComments() []string   { return f.Comment }
func (f *StructField) SetAlignedName(s string) { f.AlignedName = s }
func (f *StructField) SetAlignedType(s string) { f.AlignedType = s }
func (f *StructField) SetAlignedTag(s string)  { f.AlignedTag = s }

// NewCustomStruct creates CustomStruct from a single custom struct configuration.
func NewCustomStruct(cfg *config.Config, st *config.CustomStruct) *CustomStruct {
	fields := make([]*StructField, 0, len(st.Fields))
	for _, f := range st.Fields {
		fields = append(fields, newStructField(f))
	}
	sortFields(fields)

	groups := groupFields(fields)
	alignFields(groups)

	imports := make([]string, 0, 4+len(fields))
	imports = append(imports, "database/sql", "database/sql/driver", "encoding/json", "errors")
	for _, f := range fields {
		imports = append(imports, f.Imports...)
	}

	return &CustomStruct{
		Version: cfg.Version,
		Package: cfg.Output.Package,
		Module:  cfg.Output.Module,
		Imports: groupImports(cfg.Output.Module, imports...),
		Name:    st.Name,
		Comment: naming.SplitCommentLines(st.Comment),
		Fields:  fields,
	}
}

func newStructField(cfg *config.CustomField) *StructField {
	name := cfg.Name
	jsonName := naming.ToSnakeCase(name)

	// Collect all tags, handling json tag separately from others
	var jsonTag *Tag
	var extraTags []*Tag

	for _, v := range cfg.Tags {
		if v.Key == "json" {
			jsonTag = newJSONTag(naming.ToSnakeCase(name), v.Options...)
			continue
		}
		extraTags = append(extraTags, newCustomTags(name, v)...)
	}

	// Default json tag with omitempty if no custom json tag defined
	if jsonTag == nil {
		jsonTag = newJSONTag(jsonName, "omitempty")
	}

	tags := NewTags(jsonTag)
	tags.Add(extraTags...)

	var imports []string
	if cfg.Import != "" {
		imports = append(imports, cfg.Import)
	}

	return &StructField{
		Name:    name,
		GoType:  cfg.Type,
		Tag:     tags.String(),
		Comment: naming.SplitCommentLines(cfg.Comment),
		Imports: imports,
	}
}

func sortFields(fields []*StructField) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
}
