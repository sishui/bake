package codegen

import (
	"reflect"
	"testing"
)

func TestGroupImports(t *testing.T) {
	tests := []struct {
		name   string
		module string
		input  []Import
		want   [][]Import
	}{
		{
			name:   "empty",
			module: "github.com/user/project",
		},
		{
			name:   "standard library only",
			module: "github.com/user/project",
			input:  []Import{{Path: "fmt"}, {Path: "context"}, {Path: "strings"}},
			want:   [][]Import{{{Path: "context"}, {Path: "fmt"}, {Path: "strings"}}},
		},
		{
			name:   "third party only",
			module: "github.com/user/project",
			input:  []Import{{Path: "github.com/uptrace/bun"}, {Path: "github.com/lib/pq"}},
			want:   [][]Import{{{Path: "github.com/lib/pq"}, {Path: "github.com/uptrace/bun"}}},
		},
		{
			name:   "local only",
			module: "github.com/user/project",
			input: []Import{
				{Path: "github.com/user/project/internal/types"},
				{Path: "github.com/user/project/internal/config"},
			},
			want: [][]Import{{
				{Path: "github.com/user/project/internal/config"},
				{Path: "github.com/user/project/internal/types"},
			}},
		},
		{
			name:   "mixed",
			module: "github.com/user/project",
			input: []Import{
				{Path: "fmt"},
				{Path: "github.com/uptrace/bun"},
				{Path: "github.com/user/project/internal/types"},
			},
			want: [][]Import{
				{{Path: "fmt"}},
				{{Path: "github.com/uptrace/bun"}},
				{{Path: "github.com/user/project/internal/types"}},
			},
		},
		{
			name:   "deduplicates",
			module: "github.com/user/project",
			input:  []Import{{Path: "fmt"}, {Path: "fmt"}, {Path: "context"}},
			want:   [][]Import{{{Path: "context"}, {Path: "fmt"}}},
		},
		{
			name:   "short module name classified as local not std",
			module: "myproject",
			input:  []Import{{Path: "myproject/internal/types"}, {Path: "fmt"}},
			want:   [][]Import{{{Path: "fmt"}}, {{Path: "myproject/internal/types"}}},
		},
		{
			name:   "alias does not change the group",
			module: "github.com/user/project",
			input:  []Import{{Path: "fmt"}, {Path: "github.com/shopspring/decimal", Alias: "dec"}},
			want: [][]Import{
				{{Path: "fmt"}},
				{{Path: "github.com/shopspring/decimal", Alias: "dec"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := groupImports(tt.module, tt.input...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupImports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortedImports(t *testing.T) {
	in := map[string]Import{
		"c": {Path: "c"},
		"a": {Path: "a", Alias: "x"},
		"b": {Path: "b"},
	}
	want := []Import{{Path: "a", Alias: "x"}, {Path: "b"}, {Path: "c"}}
	if got := sortedImports(in); !reflect.DeepEqual(got, want) {
		t.Errorf("sortedImports() = %v, want %v", got, want)
	}
}
