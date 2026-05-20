package generate

import (
	"testing"
)

// ---- groupFields tests ----

func TestGroupFields_Fields(t *testing.T) {
	tests := []struct {
		name   string
		fields []*Field
		want   int // number of groups
	}{
		{
			name:   "empty",
			fields: []*Field{},
			want:   0,
		},
		{
			name: "single field no comment",
			fields: []*Field{
				{Name: "ID"},
			},
			want: 1,
		},
		{
			name: "two fields no multi-line comment",
			fields: []*Field{
				{Name: "ID"},
				{Name: "Name"},
			},
			want: 1,
		},
		{
			name: "field with multi-line comment starts new group",
			fields: []*Field{
				{Name: "ID"},
				{Name: "Name", Comments: []string{"line1", "line2"}},
				{Name: "Email"},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupFields(tt.fields)
			if len(got) != tt.want {
				t.Errorf("groupFields() returned %d groups, want %d", len(got), tt.want)
			}
		})
	}
}

func TestGroupFields_StructFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []*StructField
		want   int // number of expected groups
	}{
		{
			name:   "empty fields",
			fields: []*StructField{},
			want:   0,
		},
		{
			name: "single field without multi-line comment",
			fields: []*StructField{
				{Name: "A", GoType: "string"},
			},
			want: 1,
		},
		{
			name: "single multi-line comment field",
			fields: []*StructField{
				{Name: "A", GoType: "string", Comment: []string{"Line 1", "Line 2"}},
			},
			want: 1,
		},
		{
			name: "multi-line comment splits group",
			fields: []*StructField{
				{Name: "A", GoType: "int"},
				{Name: "B", GoType: "string", Comment: []string{"Line 1", "Line 2"}},
				{Name: "C", GoType: "bool"},
			},
			want: 2,
		},
		{
			name: "consecutive multi-line comments each start own group",
			fields: []*StructField{
				{Name: "A", GoType: "int"},
				{Name: "B", GoType: "string", Comment: []string{"L1", "L2"}},
				{Name: "C", GoType: "bool", Comment: []string{"L3", "L4"}},
			},
			want: 3,
		},
		{
			name: "multi-line comment at start doesn't create empty first group",
			fields: []*StructField{
				{Name: "A", GoType: "string", Comment: []string{"L1", "L2"}},
				{Name: "B", GoType: "bool"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := groupFields(tt.fields)
			if len(groups) != tt.want {
				t.Errorf("len(groups) = %d, want %d", len(groups), tt.want)
			}
		})
	}
}

// ---- maxFieldAttr tests ----

func TestMaxFieldAttr_Fields(t *testing.T) {
	fields := []*Field{
		{Name: "ID", Type: "int32", Tag: `bun:"id,pk"`},
		{Name: "UserName", Type: "string", Tag: `bun:"user_name"`},
		{Name: "Email", Type: "*string", Tag: `bun:"email,nullzero"`},
	}

	maxName, maxType, maxTag := maxFieldAttr(fields)
	if maxName != 8 { // "UserName"
		t.Errorf("maxName = %d, want 8", maxName)
	}
	if maxType != 7 { // "*string"
		t.Errorf("maxType = %d, want 7", maxType)
	}
	if maxTag != 20 { // `bun:"email,nullzero"`
		t.Errorf("maxTag = %d, want 20", maxTag)
	}
}

func TestMaxFieldAttr_StructFields_MultiLineCommentSkipsTag(t *testing.T) {
	// Multi-line comment fields should not contribute to maxTag
	fields := []*StructField{
		{Name: "A", GoType: "int", Tag: "`json:\"a\"`"},
		{Name: "LongDesc", GoType: "string", Tag: "`json:\"long_desc,omitempty\"`with_extra_tag", Comment: []string{"Line 1", "Line 2"}},
		{Name: "B", GoType: "bool", Tag: "`json:\"b\"`"},
	}

	maxName, maxType, maxTag := maxFieldAttr(fields)
	if maxName != 8 { // "LongDesc"
		t.Errorf("maxName = %d, want 8", maxName)
	}
	if maxType != 6 { // "string"
		t.Errorf("maxType = %d, want 6", maxType)
	}
	// maxTag should be based on A and B only (LongDesc's tag is skipped)
	if maxTag != 10 { // `json:"a"` = 10 chars, `json:"b"` = 10 chars (both include backticks)
		t.Errorf("maxTag = %d, want 10 (multi-line field's tag should be skipped)", maxTag)
	}
}

// ---- alignFields tests ----

func TestAlignFields_StructFields(t *testing.T) {
	tests := []struct {
		name   string
		groups [][]*StructField
		want   [][]*StructField
	}{
		{
			name:   "empty group",
			groups: [][]*StructField{{}},
			want:   [][]*StructField{{}},
		},
		{
			name: "single field",
			groups: [][]*StructField{{
				{Name: "Value", GoType: "string", Tag: "`json:\"value,omitempty\"`"},
			}},
			want: [][]*StructField{{
				{Name: "Value", AlignedName: "Value", GoType: "string", AlignedType: "string", Tag: "`json:\"value,omitempty\"`", AlignedTag: "`json:\"value,omitempty\"`"},
			}},
		},
		{
			name: "fields with varying widths",
			groups: [][]*StructField{{
				{Name: "A", GoType: "int", Tag: "`json:\"a\"`"},
				{Name: "LongName", GoType: "string", Tag: "`json:\"long_name,omitempty\"`"},
				{Name: "B", GoType: "bool", Tag: "`json:\"b\"`"},
			}},
			want: [][]*StructField{{
				{Name: "A", GoType: "int", Tag: "`json:\"a\"`",
					AlignedName: "A       ", AlignedType: "int   ", AlignedTag: "`json:\"a\"`                  "},
				{Name: "LongName", GoType: "string", Tag: "`json:\"long_name,omitempty\"`",
					AlignedName: "LongName", AlignedType: "string", AlignedTag: "`json:\"long_name,omitempty\"`"},
				{Name: "B", GoType: "bool", Tag: "`json:\"b\"`",
					AlignedName: "B       ", AlignedType: "bool  ", AlignedTag: "`json:\"b\"`                  "},
			}},
		},
		{
			name: "multi-line comment field skips tag alignment",
			groups: [][]*StructField{{
				{Name: "A", GoType: "int", Tag: "`json:\"a\"`"},
				{Name: "LongDesc", GoType: "string", Tag: "`json:\"long_desc,omitempty\"`with_extra_tag", Comment: []string{"Line 1", "Line 2"}},
				{Name: "B", GoType: "bool", Tag: "`json:\"b\"`"},
			}},
			want: [][]*StructField{{
				{Name: "A", GoType: "int", Tag: "`json:\"a\"`",
					AlignedName: "A       ", AlignedType: "int   ", AlignedTag: "`json:\"a\"`"},
				{Name: "LongDesc", GoType: "string", Tag: "`json:\"long_desc,omitempty\"`with_extra_tag",
					AlignedName: "LongDesc", AlignedType: "string", AlignedTag: "`json:\"long_desc,omitempty\"`with_extra_tag"},
				{Name: "B", GoType: "bool", Tag: "`json:\"b\"`",
					AlignedName: "B       ", AlignedType: "bool  ", AlignedTag: "`json:\"b\"`"},
			}},
		},
		{
			name: "multiple groups respect independent alignment",
			groups: [][]*StructField{
				{
					{Name: "Short", GoType: "int", Tag: "`json:\"short\"`"},
				},
				{
					{Name: "VeryLongFieldName", GoType: "bool", Tag: "`json:\"very_long_field_name\"`"},
				},
			},
			want: [][]*StructField{
				{
					{Name: "Short", GoType: "int", Tag: "`json:\"short\"`",
						AlignedName: "Short", AlignedType: "int", AlignedTag: "`json:\"short\"`"},
				},
				{
					{Name: "VeryLongFieldName", GoType: "bool", Tag: "`json:\"very_long_field_name\"`",
						AlignedName: "VeryLongFieldName", AlignedType: "bool", AlignedTag: "`json:\"very_long_field_name\"`"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alignFields(tt.groups)
			for g, group := range tt.groups {
				wantGroup := tt.want[g]
				for i, f := range group {
					w := wantGroup[i]
					if f.AlignedName != w.AlignedName {
						t.Errorf("Group[%d] Fields[%d].AlignedName = %q, want %q", g, i, f.AlignedName, w.AlignedName)
					}
					if f.AlignedType != w.AlignedType {
						t.Errorf("Group[%d] Fields[%d].AlignedType = %q, want %q", g, i, f.AlignedType, w.AlignedType)
					}
					if f.AlignedTag != w.AlignedTag {
						t.Errorf("Group[%d] Fields[%d].AlignedTag = %q, want %q", g, i, f.AlignedTag, w.AlignedTag)
					}
				}
			}
		})
	}
}
