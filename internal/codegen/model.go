package codegen

import (
	"sort"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
	"github.com/sishui/bake/internal/types"
)

type Model struct {
	Version             string     // bake version
	Module              string     // module path
	Package             string     // package name
	Imports             [][]string // imports
	Driver              string     // driver name
	BunModel            string     // bun.BaseModel
	Table               string     // table name
	Model               string     // model name
	Alias               string     // model alias
	Comments            []string   // model comments
	Fields              []*Field   // fields
	Timezone            string     // timezone
	CreatedAtType       string     // created_at type
	UpdatedAtType       string     // updated_at type
	DeletedAtType       string     // deleted_at type
	HasField            bool       // has at least one field
	HasNullable         bool       // has a nullable field
	HasString           bool       // has a STRING field
	HasOrdered          bool       // has an ordered field (string/numeric/time)
	HasOrderedNonString bool       // has a numeric or time field
	HasEquatable        bool       // has an equatable field (numeric/string/bool/time/enum)
	HasRelation         bool       // has a relation field
	HasArithmetic       bool       // has a non-primary numeric field
	HasTime             bool       // has a TIME field
}

func NewModel(t *schema.Table, db *config.DB, cfg *config.Config, n *naming.Naming) (*Model, error) {
	m := &Model{
		Version:  cfg.Version,
		Module:   cfg.Output.Module,
		Package:  cfg.Output.Package,
		Imports:  nil,
		Driver:   db.Driver,
		BunModel: "bun.BaseModel",
		Table:    t.Name,
		Model:    naming.TableToStruct(t.Name),
		Alias:    naming.TableToAlias(t.Name),
		Comments: naming.SplitCommentLines(t.Comment),
		Timezone: cfg.Timezone,
	}

	customTable := db.Custom[t.Name]
	applyCustom(m, customTable)
	fields := make([]*Field, 0, len(t.Columns)*2)
	for _, c := range t.Columns {
		f, err := NewField(c, customTable, db.Driver, n)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	columns := make(map[string]struct{}, len(t.Columns))
	fields = append(fields, newBelongsToFields(t.ForeignKeys, customTable)...)
	fields = append(fields, newHasManyFields(t.ReverseForeignKeys, customTable)...)

	for _, f := range fields {
		columns[f.ColumnName] = struct{}{}
	}
	fields = append(fields, newCustomFields(customTable, columns, n)...)
	m.Fields = fields
	initModel(m)
	return m, nil
}

func initModel(m *Model) {
	imports := make([]string, 0, len(m.Fields)*2+1)
	imports = append(imports, "context", "github.com/uptrace/bun")

	for _, f := range m.Fields {
		imports = append(imports, f.Imports...)

		m.HasField = true

		if f.IsNullable {
			m.HasNullable = true
		}

		if f.IsRelation {
			m.HasRelation = true
		}

		switch f.Kind {
		case types.KindString:
			m.HasString = true
			m.HasOrdered = true
			m.HasEquatable = true

		case types.KindNumeric:
			m.HasOrdered = true
			m.HasOrderedNonString = true
			m.HasEquatable = true
			if !f.IsPrimary {
				m.HasArithmetic = true
			}

		case types.KindTime:
			m.HasOrdered = true
			m.HasOrderedNonString = true
			m.HasEquatable = true
			m.HasTime = true

		case types.KindBoolean, types.KindEnum:
			m.HasEquatable = true
		}

		switch f.ColumnName {
		case "created_at":
			m.CreatedAtType = f.Type
		case "updated_at":
			m.UpdatedAtType = f.Type
		case "deleted_at":
			m.DeletedAtType = f.Type
		default:
			continue
		}
		imports = append(imports, "time")
	}

	m.Imports = groupImports(m.Module, imports...)
}

func applyCustom(m *Model, customTable *config.CustomTable) {
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

func newCustomFields(customTable *config.CustomTable, columns map[string]struct{}, n *naming.Naming) []*Field {
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
		f := NewCustomField(k, customTable, n)
		if f != nil { // defensive: avoid nil element
			results = append(results, f)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
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

func newHasManyFields(reverseForeignKeys []schema.ForeignKey, customTable *config.CustomTable) []*Field {
	if len(reverseForeignKeys) == 0 {
		return nil
	}
	results := make([]*Field, 0, len(reverseForeignKeys))
	for _, fk := range reverseForeignKeys {
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
