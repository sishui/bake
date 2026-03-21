package generate

import (
	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

type Field struct {
	Imports     []string // field imports libs
	Name        string   // field name, e.g: ID
	AlignedName string   // aligned field name e.g: ID
	Type        string   // type, e.g: string | int32 | int8
	AlignedType string   // aligned type e.g: string | int32 | int8
	Tag         string   // tag, e.g: `bun:"id,pk,autoincrement,notnull" json:"id,omitempty"`
	AlignedTag  string   // aligned tag e.g: `bun:"id,pk,autoincrement,notnull" json:"id,omitempty"`
	Comments    []string // comments e.g: user id
	ColumnName  string   // db column name e.g: id
	Kind        string   // logic type: NUMERIC |STRING | ...
	IsPrimary   bool     // is primary key
	IsNullable  bool     // is nullable
	IsCustom    bool     // is custom type
	IsRelation  bool     // is relation
}

func NewField(c *schema.Column, customTable *config.CustomTable, driver string, initialisms map[string]string) *Field {
	var desc types.Desc
	switch driver {
	case "mysql":
		desc = types.MySQLDescFunc(c)
	case "postgres":
		desc = types.PostgresDescFunc(c)
	default:
		panic("unsupported driver")
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
		Imports:    fieldImports(desc, customField),
		Name:       fieldName(c, customField, initialisms),
		Type:       fieldType(desc, customField),
		Tag:        fieldTags(c, customField, customTags).String(),
		Comments:   naming.SplitCommentLines(c.Comment),
		ColumnName: c.Name,
		Kind:       desc.Kind,
		IsPrimary:  c.IsPrimaryKey(),
		IsNullable: c.IsNullable(),
		IsCustom:   false,
		IsRelation: false,
	}
}

func NewCustomField(fieldName string, customTable *config.CustomTable) *Field {
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
		Imports:    nil,
		Name:       naming.ColumnToField(name, nil),
		Type:       customField.Type,
		Tag:        tags.String(),
		Comments:   naming.SplitCommentLines(customField.Comment),
		ColumnName: name,
		Kind:       types.TypeToKind(customField.Type),
		IsPrimary:  false,
		IsNullable: false,
		IsCustom:   true,
		IsRelation: customField.Relation,
	}
	if customField.Import != "" {
		field.Imports = append(field.Imports, customField.Import)
	}
	return field
}

func fieldName(c *schema.Column, customField *config.CustomField, initialisms map[string]string) string {
	if customField == nil || customField.Name == "" {
		return naming.ColumnToField(c.Name, initialisms)
	}
	return naming.ColumnToField(customField.Name, initialisms)
}

func fieldType(desc types.Desc, customField *config.CustomField) string {
	if customField == nil || customField.Type == "" {
		return desc.Type
	}
	return customField.Type
}

func fieldImports(desc types.Desc, customField *config.CustomField) []string {
	if customField == nil || customField.Import == "" {
		return desc.Imports
	}
	desc.Imports = append(desc.Imports, customField.Import)
	return desc.Imports
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
