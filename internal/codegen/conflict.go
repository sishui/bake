package codegen

import (
	"slices"
	"strings"

	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
)

// Conflict is an upsert conflict target of the generated model.
type Conflict struct {
	Key   string // constant suffix, e.g. "TenantIDEmail"
	Value string // CONFLICT (tenant_id, email) DO UPDATE
}

// Excluded is one assignment for the SET clause of an upsert.
type Excluded struct {
	Key   string // constant suffix, e.g. "Subject"
	Value string // subject = EXCLUDED.subject
}

func newConflicts(keys []schema.UniqueKey, columns []*schema.Column, n *naming.Naming) []Conflict {
	if len(keys) == 0 {
		return nil
	}
	conflicts := make([]Conflict, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		var suffix strings.Builder
		for _, column := range key.Columns {
			suffix.WriteString(n.ColumnToField(column))
		}
		name := suffix.String()
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		value := "CONFLICT (" + strings.Join(key.Columns, ", ") + ")"
		if hasUpdatableColumn(columns, key.Columns) {
			value += " DO UPDATE"
		} else {
			value += " DO NOTHING"
		}
		conflicts = append(conflicts, Conflict{Key: name, Value: value})
	}
	return conflicts
}

// newExcluded returns the SET assignments a caller can pick from: one per
// column, leaving out primary keys and created_at. PostgreSQL reads the new row
// from EXCLUDED; MySQL uses VALUES(col), because its `new` row alias requires an
// INSERT ... AS new that bun's MySQL dialect does not emit.
func newExcluded(columns []*schema.Column, n *naming.Naming, driver string) []Excluded {
	var excluded []Excluded
	for _, c := range columns {
		if c.IsPrimaryKey() || c.Name == "created_at" {
			continue
		}
		value := c.Name + " = EXCLUDED." + c.Name
		if driver == "mysql" {
			value = c.Name + " = VALUES(" + c.Name + ")"
		}
		excluded = append(excluded, Excluded{
			Key:   n.ColumnToField(c.Name),
			Value: value,
		})
	}
	return excluded
}

func hasUpdatableColumn(columns []*schema.Column, key []string) bool {
	for _, c := range columns {
		if c.IsPrimaryKey() || c.Name == "created_at" || slices.Contains(key, c.Name) {
			continue
		}
		return true
	}
	return false
}
