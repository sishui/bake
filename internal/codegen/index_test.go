package codegen

import (
	"reflect"
	"testing"
)

func TestNewIndexes(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []Index
	}{
		{
			name: "empty",
		},
		{
			name:   "drops the idx_ prefix",
			values: []string{"idx_users_email"},
			want:   []Index{{Key: "UsersEmail", Value: "idx_users_email"}},
		},
		{
			name:   "keeps names without the prefix",
			values: []string{"uk_email"},
			want:   []Index{{Key: "UkEmail", Value: "uk_email"}},
		},
		{
			name:   "index named idx alone keeps its name",
			values: []string{"idx_"},
			want:   []Index{{Key: "Idx", Value: "idx_"}},
		},
		{
			name:   "multiple",
			values: []string{"users_email", "idx_users_name"},
			want: []Index{
				{Key: "UsersEmail", Value: "users_email"},
				{Key: "UsersName", Value: "idx_users_name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := newIndexes(tt.values...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newIndexes() = %v, want %v", got, tt.want)
			}
		})
	}
}
