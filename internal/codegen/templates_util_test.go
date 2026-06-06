package codegen

import (
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "positive numbers", a: 3, b: 5, want: 8},
		{name: "negative numbers", a: -3, b: -5, want: -8},
		{name: "mixed signs", a: -3, b: 5, want: 2},
		{name: "zero", a: 0, b: 0, want: 0},
		{name: "zero and value", a: 10, b: 0, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "positive numbers", a: 10, b: 3, want: 7},
		{name: "negative result", a: 3, b: 10, want: -7},
		{name: "negative numbers", a: -3, b: -5, want: 2},
		{name: "zero", a: 0, b: 0, want: 0},
		{name: "subtract zero", a: 5, b: 0, want: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("sub(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDict(t *testing.T) {
	tests := []struct {
		name    string
		pairs   []any
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "empty args returns empty map",
			pairs: nil,
			want:  map[string]any{},
		},
		{
			name:  "single pair",
			pairs: []any{"key", "value"},
			want:  map[string]any{"key": "value"},
		},
		{
			name:  "multiple pairs",
			pairs: []any{"a", 1, "b", "hello", "c", true},
			want:  map[string]any{"a": 1, "b": "hello", "c": true},
		},
		{
			name:    "odd number of args returns error",
			pairs:   []any{"key", "value", "orphan"},
			wantErr: true,
		},
		{
			name:    "non-string key returns error",
			pairs:   []any{123, "value"},
			wantErr: true,
		},
		{
			name:    "single arg returns error",
			pairs:   []any{"lonely"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dict(tt.pairs...)
			if tt.wantErr {
				if err == nil {
					t.Errorf("dict(%v) expected error, got nil", tt.pairs)
				}
				return
			}
			if err != nil {
				t.Errorf("dict(%v) unexpected error: %v", tt.pairs, err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("dict() got %d entries, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("dict()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestLongest(t *testing.T) {
	fields := []*Field{
		{Name: "ID", Kind: "NUMERIC"},
		{Name: "Username", Kind: "STRING"},
		{Name: "CreatedAt", Kind: "TIME"},
		{Name: "IsActive", Kind: "BOOL"},
		{Name: "VeryLongFieldName", Kind: "STRING"},
	}

	tests := []struct {
		name   string
		fields []*Field
		kinds  []string
		want   int
	}{
		{
			name:   "match fields by kind STRING",
			fields: fields,
			kinds:  []string{"STRING"},
			want:   len("VeryLongFieldName"),
		},
		{
			name:   "match fields by kind NUMERIC",
			fields: fields,
			kinds:  []string{"NUMERIC"},
			want:   len("ID"),
		},
		{
			name:   "no matching fields returns zero",
			fields: fields,
			kinds:  []string{"UNKNOWN"},
			want:   0,
		},
		{
			name:   "multiple kinds matches longest across all",
			fields: fields,
			kinds:  []string{"STRING", "NUMERIC"},
			want:   len("VeryLongFieldName"),
		},
		{
			name:   "empty fields returns zero",
			fields: nil,
			kinds:  []string{"STRING"},
			want:   0,
		},
		{
			name:   "empty kinds returns zero",
			fields: fields,
			kinds:  nil,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longest(tt.fields, tt.kinds...)
			if got != tt.want {
				t.Errorf("longest() = %d, want %d", got, tt.want)
			}
		})
	}
}
