package codegen

import (
	"strings"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

func hasRelationship(tags ...*config.Tag) bool {
	for _, tag := range tags {
		if tag.Key != "bun" {
			continue
		}
		for _, opt := range tag.Options {
			name, value, ok := strings.Cut(opt, ":")
			if !ok {
				continue
			}
			switch strings.ToLower(name) {
			case "m2m":
				return true
			case "rel":
				switch strings.ToLower(value) {
				case "belongs-to", "has-one", "has-many":
					return true
				}
			default:
				continue
			}
		}
	}
	return false
}

func newBelongsToFields(foreignKeys []schema.ForeignKey, customTable *config.CustomTable) []*Field {
	if len(foreignKeys) == 0 {
		return nil
	}
	results := make([]*Field, 0, len(foreignKeys))
	for _, fk := range foreignKeys {
		fieldName := naming.TableToStruct(fk.RefTable)
		singular := naming.Singular(fk.RefTable)
		name := naming.ToSnakeCase(singular)
		tags := NewTags(NewTag("bun", name, "rel:belongs-to", "join:"+fk.ColumnName+"="+fk.RefColumn), NewTag("json", name, "omitempty"))

		field := newRelationField(customTable, fieldName, "*"+fieldName, tags, singular)
		field.ColumnName = name

		results = append(results, field)
	}
	return results
}

func newHasManyFields(foreignKeys []schema.ForeignKey, customTable *config.CustomTable) []*Field {
	if len(foreignKeys) == 0 {
		return nil
	}
	results := make([]*Field, 0, len(foreignKeys))
	for _, fk := range foreignKeys {
		fieldName := naming.ToCamelCase(fk.Table)
		fieldType := "[]*" + naming.TableToStruct(fk.Table)
		tags := NewTags(NewTag("bun", fk.Table, "rel:has-many", "join:"+fk.RefColumn+"="+fk.ColumnName), NewTag("json", fk.Table, "omitempty"))
		field := newRelationField(customTable, fieldName, fieldType, tags, fk.Table)
		field.ColumnName = fk.Table

		results = append(results, field)
	}
	return results
}

func newRelationField(customTable *config.CustomTable, fieldName, fieldType string, tags *Tags, name string) *Field {
	if customTable == nil {
		return &Field{
			Name:       fieldName,
			Type:       fieldType,
			Tag:        tags.String(),
			Kind:       types.KindStruct,
			IsCustom:   true,
			IsRelation: true,
		}
	}
	// Apply table-level tags
	if len(customTable.Tags) > 0 {
		tags.Add(newCustomTags(name, customTable.Tags...)...)
	}
	// Apply field-level config
	customField := customTable.Fields[name]
	if customField == nil {
		return &Field{
			Name:       fieldName,
			Type:       fieldType,
			Tag:        tags.String(),
			Kind:       types.KindStruct,
			IsCustom:   true,
			IsRelation: true,
		}
	}
	if customField.Name != "" {
		fieldName = customField.Name
	}
	if customField.Type != "" {
		fieldType = customField.Type
	}
	if len(customField.Tags) > 0 {
		tags.Add(newCustomTags(name, customField.Tags...)...)
	}
	return &Field{
		Name:       fieldName,
		Type:       fieldType,
		Tag:        tags.String(),
		Kind:       types.KindStruct,
		IsCustom:   true,
		IsRelation: true,
	}
}
