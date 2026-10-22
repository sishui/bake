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
	Imports             [][]Import // imports
	Driver              string     // driver name
	BunModel            string     // bun.BaseModel
	Table               string     // table name
	Model               string     // model name
	Alias               string     // model alias
	Comments            []string   // model comments
	Fields              []*Field   // fields
	Indexes             []Index    // indexes
	Conflicts           []Conflict // upsert conflict targets
	Excluded            []Excluded // upsert SET assignments
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
		Indexes:  newIndexes(t.Indexes...),
		Timezone: cfg.Timezone,
	}
	m.Excluded = newExcluded(t.Columns, n, db.Driver)
	if db.Driver == "postgres" {
		m.Conflicts = newConflicts(t.Keys, t.Columns, n)
	}

	customTable := db.Custom[t.Name]
	applyCustom(m, customTable)
	fields := make([]*Field, 0, len(t.Columns)*2)
	for _, c := range t.Columns {
		f, err := NewField(c, customTable, db.Driver, n, cfg.Imports)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	columns := make(map[string]struct{}, len(t.Columns))
	for _, f := range fields {
		columns[f.ColumnName] = struct{}{}
	}
	fields = append(fields, newCustomFields(customTable, columns, n, cfg.Imports)...)
	m.Fields = fields
	initModel(m)
	return m, nil
}

func initModel(m *Model) {
	imports := make([]Import, 0, len(m.Fields)*2+2)
	imports = append(imports, Import{Path: "context"}, Import{Path: "github.com/uptrace/bun"})

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
		imports = append(imports, Import{Path: "time"})
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

func newCustomFields(customTable *config.CustomTable, columns map[string]struct{}, n *naming.Naming, imports []*config.Import) []*Field {
	if customTable == nil {
		return nil
	}
	results := make([]*Field, 0, len(customTable.Fields))
	for k, field := range customTable.Fields {
		if _, ok := columns[k]; ok {
			continue
		}
		if field.Name == "" && field.Type == "" && !hasRelationship(field.Tags...) {
			continue
		}
		f := NewCustomField(k, customTable, n, imports)
		if f != nil { // defensive: avoid nil element
			results = append(results, f)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}
