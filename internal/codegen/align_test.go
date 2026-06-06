package codegen

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
			t.Parallel()
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
		fields []*Field
		want   int // number of expected groups
	}{
		{
			name:   "empty fields",
			fields: []*Field{},
			want:   0,
		},
		{
			name: "single field without multi-line comment",
			fields: []*Field{
				{Name: "A", Type: "string"},
			},
			want: 1,
		},
		{
			name: "single multi-line comment field",
			fields: []*Field{
				{Name: "A", Type: "string", Comments: []string{"Line 1", "Line 2"}},
			},
			want: 1,
		},
		{
			name: "multi-line comment splits group",
			fields: []*Field{
				{Name: "A", Type: "int"},
				{Name: "B", Type: "string", Comments: []string{"Line 1", "Line 2"}},
				{Name: "C", Type: "bool"},
			},
			want: 2,
		},
		{
			name: "consecutive multi-line comments each start own group",
			fields: []*Field{
				{Name: "A", Type: "int"},
				{Name: "B", Type: "string", Comments: []string{"L1", "L2"}},
				{Name: "C", Type: "bool", Comments: []string{"L3", "L4"}},
			},
			want: 3,
		},
		{
			name: "multi-line comment at start doesn't create empty first group",
			fields: []*Field{
				{Name: "A", Type: "string", Comments: []string{"L1", "L2"}},
				{Name: "B", Type: "bool"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	fields := []*Field{
		{Name: "A", Type: "int", Tag: "`json:\"a\"`"},
		{Name: "LongDesc", Type: "string", Tag: "`json:\"long_desc,omitempty\"`with_extra_tag", Comments: []string{"Line 1", "Line 2"}},
		{Name: "B", Type: "bool", Tag: "`json:\"b\"`"},
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
		groups [][]*Field
		want   [][]*Field
	}{
		{
			name:   "empty group",
			groups: [][]*Field{{}},
			want:   [][]*Field{{}},
		},
		{
			name: "single field",
			groups: [][]*Field{{
				{Name: "Value", Type: "string", Tag: "`json:\"value,omitempty\"`"},
			}},
			want: [][]*Field{{
				{Name: "Value", AlignedName: "Value", Type: "string", AlignedType: "string", Tag: "`json:\"value,omitempty\"`", AlignedTag: "`json:\"value,omitempty\"`"},
			}},
		},
		{
			name: "fields with varying widths",
			groups: [][]*Field{{
				{Name: "A", Type: "int", Tag: "`json:\"a\"`"},
				{Name: "LongName", Type: "string", Tag: "`json:\"long_name,omitempty\"`"},
				{Name: "B", Type: "bool", Tag: "`json:\"b\"`"},
			}},
			want: [][]*Field{{
				{Name: "A", Type: "int", Tag: "`json:\"a\"`",
					AlignedName: "A       ", AlignedType: "int   ", AlignedTag: "`json:\"a\"`                  "},
				{Name: "LongName", Type: "string", Tag: "`json:\"long_name,omitempty\"`",
					AlignedName: "LongName", AlignedType: "string", AlignedTag: "`json:\"long_name,omitempty\"`"},
				{Name: "B", Type: "bool", Tag: "`json:\"b\"`",
					AlignedName: "B       ", AlignedType: "bool  ", AlignedTag: "`json:\"b\"`                  "},
			}},
		},
		{
			name: "multi-line comment field skips tag alignment",
			groups: [][]*Field{{
				{Name: "A", Type: "int", Tag: "`json:\"a\"`"},
				{Name: "LongDesc", Type: "string", Tag: "`json:\"long_desc,omitempty\"`with_extra_tag", Comments: []string{"Line 1", "Line 2"}},
				{Name: "B", Type: "bool", Tag: "`json:\"b\"`"},
			}},
			want: [][]*Field{{
				{Name: "A", Type: "int", Tag: "`json:\"a\"`",
					AlignedName: "A       ", AlignedType: "int   ", AlignedTag: "`json:\"a\"`"},
				{Name: "LongDesc", Type: "string", Tag: "`json:\"long_desc,omitempty\"`with_extra_tag",
					AlignedName: "LongDesc", AlignedType: "string", AlignedTag: "`json:\"long_desc,omitempty\"`with_extra_tag"},
				{Name: "B", Type: "bool", Tag: "`json:\"b\"`",
					AlignedName: "B       ", AlignedType: "bool  ", AlignedTag: "`json:\"b\"`"},
			}},
		},
		{
			name: "multiple groups respect independent alignment",
			groups: [][]*Field{
				{
					{Name: "Short", Type: "int", Tag: "`json:\"short\"`"},
				},
				{
					{Name: "VeryLongFieldName", Type: "bool", Tag: "`json:\"very_long_field_name\"`"},
				},
			},
			want: [][]*Field{
				{
					{Name: "Short", Type: "int", Tag: "`json:\"short\"`",
						AlignedName: "Short", AlignedType: "int", AlignedTag: "`json:\"short\"`"},
				},
				{
					{Name: "VeryLongFieldName", Type: "bool", Tag: "`json:\"very_long_field_name\"`",
						AlignedName: "VeryLongFieldName", AlignedType: "bool", AlignedTag: "`json:\"very_long_field_name\"`"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
