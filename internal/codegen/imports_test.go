package generate

import (
	"reflect"
	"testing"
)

func TestGroupImports(t *testing.T) {
	tests := []struct {
		name   string
		module string
		input  []string
		want   [][]string
	}{
		{
			name:   "empty",
			module: "github.com/user/project",
			input:  nil,
			want:   nil,
		},
		{
			name:   "standard library only",
			module: "github.com/user/project",
			input:  []string{"fmt", "context", "strings"},
			want:   [][]string{{"context", "fmt", "strings"}},
		},
		{
			name:   "third party only",
			module: "github.com/user/project",
			input:  []string{"github.com/uptrace/bun", "github.com/lib/pq"},
			want:   [][]string{{"github.com/lib/pq", "github.com/uptrace/bun"}},
		},
		{
			name:   "local only",
			module: "github.com/user/project",
			input:  []string{"github.com/user/project/internal/types", "github.com/user/project/internal/config"},
			want:   [][]string{{"github.com/user/project/internal/config", "github.com/user/project/internal/types"}},
		},
		{
			name:   "mixed",
			module: "github.com/user/project",
			input:  []string{"fmt", "github.com/uptrace/bun", "github.com/user/project/internal/types"},
			want: [][]string{
				{"fmt"},
				{"github.com/uptrace/bun"},
				{"github.com/user/project/internal/types"},
			},
		},
		{
			name:   "deduplicates",
			module: "github.com/user/project",
			input:  []string{"fmt", "fmt", "context"},
			want:   [][]string{{"context", "fmt"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupImports(tt.module, tt.input...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupImports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]struct{}
		want []string
	}{
		{
			name: "empty",
			m:    map[string]struct{}{},
			want: nil,
		},
		{
			name: "single",
			m:    map[string]struct{}{"a": {}},
			want: []string{"a"},
		},
		{
			name: "multiple",
			m:    map[string]struct{}{"c": {}, "a": {}, "b": {}},
			want: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedKeys(tt.m)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortedKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
