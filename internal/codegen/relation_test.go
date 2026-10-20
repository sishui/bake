package codegen

import (
	"testing"

	"github.com/sishui/bake/internal/config"
)

func TestHasRelationship(t *testing.T) {
	tests := []struct {
		name string
		tags []*config.Tag
		want bool
	}{
		{
			name: "belongs-to",
			tags: []*config.Tag{{Key: "bun", Options: []string{"rel:belongs-to", "join:user_id=id"}}},
			want: true,
		},
		{
			name: "has-one",
			tags: []*config.Tag{{Key: "bun", Options: []string{"rel:has-one", "join:id=user_id"}}},
			want: true,
		},
		{
			name: "has-many",
			tags: []*config.Tag{{Key: "bun", Options: []string{"rel:has-many", "join:id=user_id"}}},
			want: true,
		},
		{
			name: "m2m with join table",
			tags: []*config.Tag{{Key: "bun", Options: []string{"m2m:user_roles"}}},
			want: true,
		},
		{
			name: "m2m without a join table",
			tags: []*config.Tag{{Key: "bun", Options: []string{"m2m"}}},
			want: false,
		},
		{
			name: "relation option is matched case-insensitively",
			tags: []*config.Tag{{Key: "bun", Options: []string{"REL:Has-Many", "join:id=user_id"}}},
			want: true,
		},
		{
			name: "rel without a value",
			tags: []*config.Tag{{Key: "bun", Options: []string{"rel"}}},
			want: false,
		},
		{
			name: "rel with an unsupported value",
			tags: []*config.Tag{{Key: "bun", Options: []string{"rel:m2m"}}},
			want: false,
		},
		{
			name: "relation option on a non-bun tag",
			tags: []*config.Tag{{Key: "json", Options: []string{"rel:has-many"}}},
			want: false,
		},
		{
			name: "join without a relation option",
			tags: []*config.Tag{{Key: "bun", Options: []string{"join:id=user_id", "notnull"}}},
			want: false,
		},
		{
			name: "relation among unrelated tags",
			tags: []*config.Tag{
				{Key: "json", Options: []string{"posts,omitempty"}},
				{Key: "bun", Options: []string{"rel:has-many", "join:id=user_id"}},
			},
			want: true,
		},
		{
			name: "no tags",
			tags: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := hasRelationship(tt.tags...); got != tt.want {
				t.Errorf("hasRelationship(%v) = %v, want %v", tt.tags, got, tt.want)
			}
		})
	}
}
