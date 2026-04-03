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
					Key:             "PRI",
					Extra:           "auto_increment",
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
			if got := newBunTag(&tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BunTag() = %v, want %v", got, tt.want)
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
			if got := tt.args.tags.String(); got != tt.want {
				t.Errorf("TagToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
