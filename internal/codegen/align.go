package generate

import (
	"github.com/sishui/bake/internal/naming"
)

// groupFields splits fields into groups, starting a new group before a
// multi-line comment field. This ensures multi-line comment field tags don't
// influence the tag alignment of preceding fields.
func groupFields(fields []*Field) [][]*Field {
	groups := make([][]*Field, 0, len(fields))
	current := make([]*Field, 0, len(fields))

	for _, f := range fields {
		if len(f.Comments) > 1 && len(current) > 0 {
			groups = append(groups, current)
			current = make([]*Field, 0, len(fields))
		}
		current = append(current, f)
	}

	if len(current) > 0 {
		groups = append(groups, current)
	}

	return groups
}

// alignFields aligns field name, type, and tag within each group using
// naming.Align. Multi-line comment fields are skipped for tag alignment.
func alignFields(groups [][]*Field) {
	for _, group := range groups {
		maxName, maxType, maxTag := maxFieldAttr(group)
		for _, f := range group {
			f.AlignedName = naming.Align(f.Name, maxName)
			f.AlignedType = naming.Align(f.Type, maxType)
			f.AlignedTag = naming.Align(f.Tag, maxTag)
		}
	}
}

// maxFieldAttr computes the maximum name, type, and tag lengths in a group.
// Fields with multi-line comments do not contribute to maxTag.
func maxFieldAttr(fields []*Field) (int, int, int) {
	var maxName, maxType, maxTag int
	for _, f := range fields {
		if len(f.Name) > maxName {
			maxName = len(f.Name)
		}
		if len(f.Type) > maxType {
			maxType = len(f.Type)
		}
		if len(f.Comments) > 1 {
			continue
		}
		if len(f.Tag) > maxTag {
			maxTag = len(f.Tag)
		}
	}
	return maxName, maxType, maxTag
}
