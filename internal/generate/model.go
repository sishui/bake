// Package generate provides for bun model
package generate

import (
	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

type Model struct {
	Version            string     // bake version
	Package            string     // package name
	Imports            [][]string // imports
	BunModel           string     // bun.BaseModel
	Table              string     // table name
	Model              string     // model name
	Alias              string     // model alias
	Comments           []string   // model comments
	Fields             []*Field   // fields
	Timezone           string     // timezone
	CreatedAtType      string     // created_at type
	UpdatedAtType      string     // updated_at type
	DeletedAtType      string     // deleted_at type
	MaxFieldLength     int        // max field length
	MaxNullableLength  int        // max nullable length
	MaxStringLength    int        // max string length
	MaxNumericLength   int        // max numeric length
	MaxOrderedLength   int        // max ordered length
	MaxEquatableLength int        // max equatable length
	MaxRelationLength  int        // max relation length
}

func NewModel(t *schema.Table, db *config.DB, cfg *config.Config, initialisms map[string]string) (*Model, error) {
	m := &Model{
		Version:  cfg.Version,
		Package:  cfg.Output.Package,
		Imports:  nil,
		BunModel: "bun.BaseModel",
		Table:    t.Name,
		Model:    naming.TableToStruct(t.Name),
		Alias:    naming.TableToAlias(t.Name),
		Comments: naming.SplitCommentLines(t.Comment),
		Timezone: cfg.Timezone,
	}
	fields := make([]*Field, 0, len(t.Columns)+1)
	customTable := db.Customs[t.Name]
	m.applyCustom(customTable)
	fields = append(fields, &Field{Name: m.BunModel})
	columns := make(map[string]struct{}, len(t.Columns))
	for _, c := range t.Columns {
		f, err := NewField(c, customTable, db.Driver, initialisms)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
		columns[c.Name] = struct{}{}
	}

	fields = append(fields, newCustomFields(customTable, columns)...)
	fields = append(fields, newBelongsToRelationFields(t.ForeignKeys, customTable)...)
	fields = append(fields, newReverseRelationFields(t.ReverseForeignKeys, columns, customTable)...)
	alignFields(groupFields(fields))
	m.Fields = fields[1:]
	m.init()
	return m, nil
}

func (m *Model) init() {
	imports := make([]string, 0, len(m.Fields)*2+1)
	imports = append(imports, "context", "github.com/uptrace/bun")

	for _, f := range m.Fields {
		imports = append(imports, f.Imports...)
		nameLen := len(f.Name)

		m.MaxFieldLength = max(m.MaxFieldLength, nameLen)

		if f.IsNullable {
			m.MaxNullableLength = max(m.MaxNullableLength, nameLen)
		}

		if f.IsRelation {
			m.MaxRelationLength = max(m.MaxRelationLength, nameLen)
		}

		switch f.Kind {
		case types.KindString:
			m.MaxStringLength = max(m.MaxStringLength, nameLen)
			m.MaxOrderedLength = max(m.MaxOrderedLength, nameLen)
			m.MaxEquatableLength = max(m.MaxEquatableLength, nameLen)

		case types.KindNumeric:
			m.MaxNumericLength = max(m.MaxNumericLength, nameLen)
			m.MaxOrderedLength = max(m.MaxOrderedLength, nameLen)
			m.MaxEquatableLength = max(m.MaxEquatableLength, nameLen)

		case types.KindTime:
			m.MaxOrderedLength = max(m.MaxOrderedLength, nameLen)
			m.MaxEquatableLength = max(m.MaxEquatableLength, nameLen)

		case types.KindBoolean, types.KindEnum:
			m.MaxEquatableLength = max(m.MaxEquatableLength, nameLen)
		}
		switch f.ColumnName {
		case "created_at":
			m.CreatedAtType = f.Type
			imports = append(imports, "time")
		case "updated_at":
			m.UpdatedAtType = f.Type
			imports = append(imports, "time")
		case "deleted_at":
			m.DeletedAtType = f.Type
			imports = append(imports, "time")
		default:
		}
	}

	m.Imports = groupImports(m.Package, imports...)
	m.BunModel = naming.Align(m.BunModel, m.MaxFieldLength)
}

func (m *Model) applyCustom(customTable *config.CustomTable) {
	if customTable == nil {
		return
	}
	if customTable.Name != "" {
		m.Model = naming.TableToStruct(customTable.Name)
	}
	if customTable.Alias != "" {
		m.Alias = naming.TableToAlias(customTable.Alias)
	}
	if customTable.Comment != "" {
		m.Comments = naming.SplitCommentLines(customTable.Comment)
	}
}

func groupFields(fields []*Field) [][]*Field {
	groups := make([][]*Field, 0, len(fields))
	var current []*Field

	for _, field := range fields {
		if len(field.Comments) > 1 && len(current) > 0 {
			groups = append(groups, current)
			current = make([]*Field, 0, len(fields))
		}
		current = append(current, field)
	}

	if len(current) > 0 {
		groups = append(groups, current)
	}

	return groups
}

func alignFields(groups [][]*Field) {
	for _, group := range groups {
		maxName, maxType, maxTag := maxFieldAttr(group)
		for j, field := range group {
			group[j].AlignedName = naming.Align(field.Name, maxName)
			group[j].AlignedType = naming.Align(field.Type, maxType)
			if len(field.Comments) > 1 {
				continue
			}
			group[j].AlignedTag = naming.Align(field.Tag, maxTag)
		}
	}
}

func maxFieldAttr(fields []*Field) (int, int, int) {
	var maxName, maxType, maxTag int
	for _, field := range fields {
		maxName = max(len(field.Name), maxName)
		maxType = max(len(field.Type), maxType)
		if len(field.Comments) > 1 {
			continue
		}
		maxTag = max(len(field.Tag), maxTag)
	}
	return maxName, maxType, maxTag
}

func newCustomFields(customTable *config.CustomTable, columns map[string]struct{}) []*Field {
	if customTable == nil {
		return nil
	}
	results := make([]*Field, 0, len(customTable.Fields))
	for k, field := range customTable.Fields {
		if _, ok := columns[k]; ok {
			continue
		}
		if field.Name == "" && field.Type == "" && !field.Relation {
			continue
		}
		results = append(results, NewCustomField(k, customTable))
	}
	return results
}

func newBelongsToRelationFields(foreignKeys []schema.ForeignKey, customTable *config.CustomTable) []*Field {
	if len(foreignKeys) == 0 {
		return nil
	}
	results := make([]*Field, 0, len(foreignKeys))
	for _, fk := range foreignKeys {
		// Generate the belongs-to relation field
		// e.g., for posts.user_id -> users.id, generate User *User
		fieldName := naming.TableToStruct(fk.RefTable) // "users" -> "User"

		// Check if custom field exists for merging
		var customField *config.CustomField
		if customTable != nil {
			name := naming.Singular(fk.RefTable) // "user"
			customField = customTable.Fields[name]
		}

		name := naming.ToSnakeCase(naming.Singular(fk.RefTable))
		fieldType := "*" + fieldName // "*User"
		tags := NewTags(NewTag("bun", name, "rel:belongs-to", "join:"+fk.ColumnName+"="+fk.RefColumn), NewTag("json", name, "omitempty"))

		// Merge with custom config if exists
		if customField != nil {
			if customField.Name != "" {
				fieldName = customField.Name
			}
			if customField.Type != "" {
				fieldType = customField.Type
			}
			if len(customField.Tags) > 0 {
				tags = NewTags(newCustomTags(name, customField.Tags...)...)
			}
		}

		field := &Field{
			Name:       fieldName,
			Type:       fieldType,
			Tag:        tags.String(),
			ColumnName: fk.ColumnName,
			Kind:       types.KindStruct,
			IsCustom:   true,
			IsRelation: true,
		}
		results = append(results, field)
	}
	return results
}

func newReverseRelationFields(reverseForeignKeys []schema.ForeignKey, columns map[string]struct{}, customTable *config.CustomTable) []*Field {
	if len(reverseForeignKeys) == 0 {
		return nil
	}
	results := make([]*Field, 0, len(reverseForeignKeys))
	for _, fk := range reverseForeignKeys {
		// Skip if the column already exists
		if _, ok := columns[fk.ColumnName]; ok {
			continue
		}

		fieldName := naming.ToCamelCase(fk.Table)           // "Posts"
		fieldType := "[]*" + naming.TableToStruct(fk.Table) // "[]*Post"
		tags := NewTags(NewTag("bun", fk.Table, "rel:has-many", "join:"+fk.RefColumn+"="+fk.ColumnName), NewTag("json", fk.Table, "omitempty"))

		// Check if custom field exists for merging
		var customField *config.CustomField
		if customTable != nil {
			customField = customTable.Fields[fk.Table]
		}

		// Merge with custom config if exists
		if customField != nil {
			if customField.Name != "" {
				fieldName = customField.Name
			}
			if customField.Type != "" {
				fieldType = customField.Type
			}
			if len(customField.Tags) > 0 {
				tags = NewTags(newCustomTags(fk.Table, customField.Tags...)...)
			}
		}

		field := &Field{
			Name:       fieldName,
			Type:       fieldType,
			Tag:        tags.String(),
			ColumnName: fk.ColumnName,
			Kind:       types.KindStruct,
			IsCustom:   true,
			IsRelation: true,
		}
		results = append(results, field)
	}
	return results
}
