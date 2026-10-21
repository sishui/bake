package codegen

import (
	"sort"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

type Field struct {
	Imports    []Import // imports required by the field type
	Name       string   // field name, e.g: ID
	Type       string   // type, e.g: string | int32 | int8
	Tag        string   // tag, e.g: `bun:"id,pk,autoincrement,notnull" json:"id,omitempty"`
	Comments   []string // comments e.g: user id
	ColumnName string   // db column name e.g: id
	Kind       string   // logic type: NUMERIC |STRING | ...
	IsPrimary  bool     // is primary key
	IsNullable bool     // is nullable
	IsCustom   bool     // is custom type
	IsRelation bool     // is relation
}

func NewField(c *schema.Column, customTable *config.CustomTable, driver string, n *naming.Naming, imports []*config.Import) (*Field, error) {
	desc, err := types.NewDesc(driver, c)
	if err != nil {
		return nil, err
	}
	var (
		customField *config.CustomField
		customTags  []*config.Tag
	)
	if customTable != nil {
		customTags = customTable.Tags
		customField = customTable.Fields[c.Name]
	}
	return &Field{
		Imports:    fieldImports(desc, customField, imports),
		Name:       fieldName(c, customField, n),
		Type:       fieldType(desc, customField),
		Tag:        fieldTags(c, customField, customTags).String(),
		Comments:   naming.SplitCommentLines(c.Comment),
		ColumnName: c.Name,
		Kind:       desc.Kind,
		IsPrimary:  c.IsPrimaryKey(),
		IsNullable: c.IsNullable(),
		IsCustom:   false,
		IsRelation: false,
	}, nil
}

func NewCustomField(fieldName string, customTable *config.CustomTable, n *naming.Naming, imports []*config.Import) *Field {
	customField, ok := customTable.Fields[fieldName]
	if !ok {
		return nil
	}
	name := fieldName
	if customField.Name != "" {
		name = customField.Name
	}
	tags := NewTags(newJSONTag(name, "omitempty"))
	tags.Add(newCustomTags(name, customTable.Tags...)...)
	tags.Add(newCustomTags(name, customField.Tags...)...)
	field := &Field{
		Imports:    importsForType(customField.Type, imports),
		Name:       n.ColumnToField(name),
		Type:       customField.Type,
		Tag:        tags.String(),
		Comments:   naming.SplitCommentLines(customField.Comment),
		ColumnName: name,
		Kind:       types.TypeToKind(customField.Type),
		IsPrimary:  false,
		IsNullable: false,
		IsCustom:   true,
		IsRelation: hasRelationship(customField.Tags...),
	}
	return field
}

func fieldName(c *schema.Column, customField *config.CustomField, naming *naming.Naming) string {
	if customField == nil || customField.Name == "" {
		return naming.ColumnToField(c.Name)
	}
	return naming.ColumnToField(customField.Name)
}

func fieldType(desc types.Desc, customField *config.CustomField) string {
	if customField == nil || customField.Type == "" {
		return desc.Type
	}
	return customField.Type
}

func fieldImports(desc types.Desc, customField *config.CustomField, imports []*config.Import) []Import {
	var extra []Import
	if customField != nil {
		extra = importsForType(customField.Type, imports)
	}
	if len(desc.Imports) == 0 {
		return extra
	}
	result := make([]Import, 0, len(desc.Imports)+len(extra))
	for _, path := range desc.Imports {
		result = append(result, Import{Path: path})
	}
	return append(result, extra...)
}

func fieldTags(c *schema.Column, customField *config.CustomField, objectTags []*config.Tag) *Tags {
	tags := NewTags(newBunTag(c), newJSONTag(c.Name, "omitempty"))
	tags.Add(newCustomTags(c.Name, objectTags...)...)
	if customField == nil {
		return tags
	}
	for _, v := range customField.Tags {
		name := v.Name
		if name == "" {
			if customField.Name == "" {
				name = c.Name
			} else {
				name = customField.Name
			}
		}
		if v.Key == "json" {
			options := make([]string, 0, len(v.Options))
			options = append(options, "omitempty")
			tags.Add(newJSONTag(name, options...))
			continue
		}
		tags.Add(newCustomTags(name, v)...)
	}
	return tags
}

func sortFields(fields []*Field) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
}
