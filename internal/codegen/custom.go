package generate

import (
	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
)

// NewCustomStruct creates Model from a single custom struct configuration.
func NewCustomStruct(cfg *config.Config, st *config.CustomStruct) *Model {
	fields := make([]*Field, 0, len(st.Fields))
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

	return &Model{
		Version:  cfg.Version,
		Package:  cfg.Output.Package,
		Module:   cfg.Output.Module,
		Imports:  groupImports(cfg.Output.Module, imports...),
		Model:    st.Name,
		Comments: naming.SplitCommentLines(st.Comment),
		Fields:   fields,
	}
}

func newStructField(cfg *config.CustomField) *Field {
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

	return &Field{
		Name:     name,
		Type:     cfg.Type,
		Tag:      tags.String(),
		Comments: naming.SplitCommentLines(cfg.Comment),
		Imports:  imports,
	}
}
