package generate

import (
	"reflect"
	"testing"

	"github.com/sishui/bake/internal/schema"
)

func TestBunTag(t *testing.T) {
	type args struct {
		c schema.Column
	}
	tests := []struct {
		name string
		args args
		want *Tag
	}{
		{
			name: "Test pk and autoincrement",
			args: args{
				c: schema.Column{
					Table:           "Users",
					Name:            "id",
					OrdinalPosition: 1,
					Default:         "",
					Nullable:        "NO",
					DataType:        "int",
					ColumnType:      "int(11) unsigned",
					Key:             schema.ColumnKeyPrimary,
					Extra:           schema.ColumnExtraAutoIncrement,
					Comment:         "user id",
				},
			},
			want: NewTag("bun", "id", "pk", "autoincrement", "notnull"),
		},
		{
			name: "Test isnullable",
			args: args{
				c: schema.Column{
					Table:           "Users",
					Name:            "name",
					OrdinalPosition: 2,
					Default:         "",
					Nullable:        "YES",
					DataType:        "string",
					ColumnType:      "varchar(31)",
					Key:             "",
					Extra:           "",
					Comment:         "user name",
				},
			},
			want: NewTag("bun", "name", "nullzero"),
		},
		{
			name: "Test createdAt",
			args: args{
				c: schema.Column{
					Table:           "Users",
					Name:            "createdAt",
					OrdinalPosition: 3,
					Default:         "CURRENT_TIMESTAMP",
					Nullable:        "NO",
					DataType:        "bigint",
					ColumnType:      "bigint",
					Key:             "",
					Extra:           "",
					Comment:         "created time",
				},
			},
			want: NewTag("bun", "createdAt", "notnull", "default:current_timestamp"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := newBunTag(&tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BunTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptPK(t *testing.T) {
	tests := []struct {
		name    string
		col     schema.Column
		wantLen int
	}{
		{
			name:    "primary key",
			col:     schema.Column{Key: schema.ColumnKeyPrimary},
			wantLen: 1,
		},
		{
			name:    "not primary key - empty",
			col:     schema.Column{Key: ""},
			wantLen: 0,
		},
		{
			name:    "not primary key - MUL",
			col:     schema.Column{Key: "MUL"},
			wantLen: 0,
		},
		{
			name:    "not primary key - UNI",
			col:     schema.Column{Key: "UNI"},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			options := make([]string, 0, 8)
			got := optPK(&tt.col, options)
			if len(got) != tt.wantLen {
				t.Errorf("optPK() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0] != "pk" {
				t.Errorf("optPK() = %v, want [pk]", got)
			}
		})
	}
}

func TestOptUnique(t *testing.T) {
	tests := []struct {
		name    string
		col     schema.Column
		want    []string
	}{
		{
			name:    "single unique",
			col:     schema.Column{IndexName: "idx_email", NonUnique: 0, IsMultiKey: false},
			want:    []string{"unique"},
		},
		{
			name:    "composite unique",
			col:     schema.Column{IndexName: "idx_unique_range", NonUnique: 0, IsMultiKey: true},
			want:    []string{"unique:idx_unique_range"},
		},
		{
			name:    "non-unique index",
			col:     schema.Column{IndexName: "idx_name", NonUnique: 1, IsMultiKey: false},
			want:    nil,
		},
		{
			name:    "no index",
			col:     schema.Column{},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optUnique(&tt.col, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("optUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptAutoIncr(t *testing.T) {
	tests := []struct {
		name    string
		col     schema.Column
		wantLen int
	}{
		{
			name:    "has auto_increment",
			col:     schema.Column{Extra: schema.ColumnExtraAutoIncrement},
			wantLen: 1,
		},
		{
			name:    "no auto_increment",
			col:     schema.Column{Extra: ""},
			wantLen: 0,
		},
		{
			name:    "other extra",
			col:     schema.Column{Extra: "DEFAULT_GENERATED"},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optAutoIncr(&tt.col, nil)
			if len(got) != tt.wantLen {
				t.Errorf("optAutoIncr() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0] != "autoincrement" {
				t.Errorf("optAutoIncr() = %v, want [autoincrement]", got)
			}
		})
	}
}

func TestOptDataType(t *testing.T) {
	tests := []struct {
		name string
		col  schema.Column
		want []string
	}{
		{
			name: "decimal type",
			col:  schema.Column{DataType: "decimal", ColumnType: "decimal(10,2)"},
			want: []string{"type:decimal(10,2)"},
		},
		{
			name: "int type",
			col:  schema.Column{DataType: "int", ColumnType: "int(11)"},
			want: nil,
		},
		{
			name: "varchar type",
			col:  schema.Column{DataType: "varchar", ColumnType: "varchar(255)"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optDataType(&tt.col, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("optDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptNullable(t *testing.T) {
	tests := []struct {
		name string
		col  schema.Column
		want []string
	}{
		{
			name: "nullable non-deleted_at",
			col:  schema.Column{Name: "name", Nullable: "YES"},
			want: []string{"nullzero"},
		},
		{
			name: "not null non-deleted_at",
			col:  schema.Column{Name: "name", Nullable: "NO"},
			want: []string{"notnull"},
		},
		{
			name: "deleted_at nullable - skip",
			col:  schema.Column{Name: "deleted_at", Nullable: "YES"},
			want: nil,
		},
		{
			name: "deleted_at not null - skip",
			col:  schema.Column{Name: "deleted_at", Nullable: "NO"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optNullable(&tt.col, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("optNullable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptDefault(t *testing.T) {
	tests := []struct {
		name string
		col  schema.Column
		want []string
	}{
		{
			name: "CURRENT_TIMESTAMP default",
			col:  schema.Column{Default: "CURRENT_TIMESTAMP"},
			want: []string{"default:current_timestamp"},
		},
		{
			name: "lowercase current_timestamp default",
			col:  schema.Column{Default: "current_timestamp"},
			want: []string{"default:current_timestamp"},
		},
		{
			name: "string default",
			col:  schema.Column{Default: "active"},
			want: []string{"default:'active'"},
		},
		{
			name: "default with single quote",
			col:  schema.Column{Default: "it's"},
			want: []string{"default:'it\\'s'"},
		},
		{
			name: "numeric default",
			col:  schema.Column{Default: "0"},
			want: []string{"default:'0'"},
		},
		{
			name: "empty default",
			col:  schema.Column{Default: ""},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optDefault(&tt.col, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("optDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptSoftDelete(t *testing.T) {
	tests := []struct {
		name string
		col  schema.Column
		want []string
	}{
		{
			name: "deleted_at column",
			col:  schema.Column{Name: "deleted_at"},
			want: []string{"soft_delete", "nullzero"},
		},
		{
			name: "other column",
			col:  schema.Column{Name: "created_at"},
			want: nil,
		},
		{
			name: "regular column",
			col:  schema.Column{Name: "name"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optSoftDelete(&tt.col, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("optSoftDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBunTag_DeletedAt(t *testing.T) {
	tests := []struct {
		name string
		col  schema.Column
		want *Tag
	}{
		{
			name: "deleted_at nullable",
			col: schema.Column{
				Name:     "deleted_at",
				Nullable: "YES",
				DataType: "timestamp",
			},
			want: NewTag("bun", "deleted_at", "soft_delete", "nullzero"),
		},
		{
			name: "deleted_at not null with default",
			col: schema.Column{
				Name:     "deleted_at",
				Nullable: "NO",
				DataType: "timestamp",
				Default:  "CURRENT_TIMESTAMP",
			},
			want: NewTag("bun", "deleted_at", "default:current_timestamp", "soft_delete", "nullzero"),
		},
		{
			name: "deleted_at with unique index",
			col: schema.Column{
				Name:      "deleted_at",
				Nullable:  "YES",
				DataType:  "timestamp",
				IndexName: "idx_deleted_at",
				NonUnique: 0,
			},
			want: NewTag("bun", "deleted_at", "unique", "soft_delete", "nullzero"),
		},
		{
			name: "deleted_at composite unique",
			col: schema.Column{
				Name:      "deleted_at",
				Nullable:  "YES",
				DataType:  "timestamp",
				IndexName: "idx_unique",
				NonUnique: 0,
				IsMultiKey: true,
			},
			want: NewTag("bun", "deleted_at", "unique:idx_unique", "soft_delete", "nullzero"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := newBunTag(&tt.col)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newBunTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonTag(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Tag
	}{
		{
			name: "Test json CamelCase",
			args: args{
				name: "CamelCase",
			},
			want: NewTag("json", "CamelCase", "omitempty"),
		},
		{
			name: "Test json snake_case",
			args: args{
				name: "camel_case",
			},
			want: NewTag("json", "camel_case", "omitempty"),
		},
		{
			name: "Test json camelCase",
			args: args{
				name: "camelCase",
			},
			want: NewTag("json", "camelCase", "omitempty"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := newJSONTag(tt.args.name, "omitempty"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagToString(t *testing.T) {
	type args struct {
		tags *Tags
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test tag to string with bun and json",
			args: args{
				tags: NewTags(
					NewTag("bun", "id", "pk", "autoincrement", "nullzero", "notnull"),
					newJSONTag("name", "omitempty"),
				),
			},
			want: "`bun:\"id,pk,autoincrement,nullzero,notnull\" json:\"name,omitempty\"`",
		},
		{
			name: "Test tag to string with bun json csv",
			args: args{
				tags: NewTags(
					NewTag("bun", "created_at", "pk", "autoincrement", "nullzero", "notnull", "default:'CURRENT_TIMESTAMP'"),
					newJSONTag("name", "omitempty"),
					NewTag("csv", "name"),
				),
			},
			want: "`bun:\"created_at,pk,autoincrement,nullzero,notnull,default:'CURRENT_TIMESTAMP'\" csv:\"name\" json:\"name,omitempty\"`",
		},
		{
			name: "Test tag to string with bun json csv xml",
			args: args{
				tags: NewTags(
					NewTag("bun", "created_at", "pk", "autoincrement", "nullzero", "notnull"),
					newJSONTag("name", "omitempty"),
					NewTag("csv", "name"),
					NewTag("xml", "name"),
				),
			},
			want: "`bun:\"created_at,pk,autoincrement,nullzero,notnull\" csv:\"name\" json:\"name,omitempty\" xml:\"name\"`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.args.tags.String(); got != tt.want {
				t.Errorf("TagToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
