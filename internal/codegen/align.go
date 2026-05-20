package generate

import (
	"github.com/sishui/bake/internal/naming"
)

// FieldAlignable is the interface for field types that support alignment.
// Both *Field (model) and *StructField (custom struct) implement this.
type FieldAlignable interface {
	FieldName() string
	FieldType() string
	GetTag() string
	GetComments() []string
	SetAlignedName(string)
	SetAlignedType(string)
	SetAlignedTag(string)
}

// groupFields splits fields into groups, starting a new group before a
// multi-line comment field. This ensures multi-line comment field tags don't
// influence the tag alignment of preceding fields.
func groupFields[T FieldAlignable](fields []T) [][]T {
	groups := make([][]T, 0, len(fields))
	current := make([]T, 0, len(fields))

	for _, f := range fields {
		if len(f.GetComments()) > 1 && len(current) > 0 {
			groups = append(groups, current)
			current = make([]T, 0, len(fields))
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
func alignFields[T FieldAlignable](groups [][]T) {
	for _, group := range groups {
		maxName, maxType, maxTag := maxFieldAttr(group)
		for _, f := range group {
			f.SetAlignedName(naming.Align(f.FieldName(), maxName))
			f.SetAlignedType(naming.Align(f.FieldType(), maxType))
			f.SetAlignedTag(naming.Align(f.GetTag(), maxTag))
		}
	}
}

// maxFieldAttr computes the maximum name, type, and tag lengths in a group.
// Fields with multi-line comments do not contribute to maxTag.
func maxFieldAttr[T FieldAlignable](fields []T) (int, int, int) {
	var maxName, maxType, maxTag int
	for _, f := range fields {
		if len(f.FieldName()) > maxName {
			maxName = len(f.FieldName())
		}
		if len(f.FieldType()) > maxType {
			maxType = len(f.FieldType())
		}
		if len(f.GetComments()) > 1 {
			continue
		}
		if len(f.GetTag()) > maxTag {
			maxTag = len(f.GetTag())
		}
	}
	return maxName, maxType, maxTag
}
