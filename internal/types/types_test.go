package types

import (
	"errors"
	"testing"

	"github.com/sishui/bake/internal/schema"
)

func TestMySQLDescFunc(t *testing.T) {
	tests := []struct {
		name    string
		column  schema.Column
		want    Desc
		wantErr bool
	}{
		{
			name: "tinyint",
			column: schema.Column{
				DataType:   "tinyint",
				ColumnType: "tinyint(4)",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int8", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "tinyint bool",
			column: schema.Column{
				DataType:   "tinyint",
				ColumnType: "tinyint(1)",
				Nullable:   "NO",
			},
			want:    Desc{Type: "bool", Kind: KindBoolean},
			wantErr: false,
		},
		{
			name: "smallint",
			column: schema.Column{
				DataType:   "smallint",
				ColumnType: "smallint(6)",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int16", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "int",
			column: schema.Column{
				DataType:   "int",
				ColumnType: "int(11)",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int32", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "bigint",
			column: schema.Column{
				DataType:   "bigint",
				ColumnType: "bigint(20)",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int64", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "float",
			column: schema.Column{
				DataType:   "float",
				ColumnType: "float",
				Nullable:   "NO",
			},
			want:    Desc{Type: "float32", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "double",
			column: schema.Column{
				DataType:   "double",
				ColumnType: "double",
				Nullable:   "NO",
			},
			want:    Desc{Type: "float64", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "decimal",
			column: schema.Column{
				DataType:   "decimal",
				ColumnType: "decimal(10,2)",
				Nullable:   "NO",
			},
			want: Desc{
				Type:    "decimal.Decimal",
				Kind:    KindNumeric,
				Imports: []string{"github.com/shopspring/decimal"},
			},
			wantErr: false,
		},
		{
			name: "varchar",
			column: schema.Column{
				DataType:   "varchar",
				ColumnType: "varchar(255)",
				Nullable:   "YES",
			},
			want:    Desc{Type: "*string", Kind: KindString},
			wantErr: false,
		},
		{
			name: "text",
			column: schema.Column{
				DataType:   "text",
				ColumnType: "text",
				Nullable:   "NO",
			},
			want:    Desc{Type: "string", Kind: KindString},
			wantErr: false,
		},
		{
			name: "blob",
			column: schema.Column{
				DataType:   "blob",
				ColumnType: "blob",
				Nullable:   "NO",
			},
			want:    Desc{Type: "[]byte", Kind: KindBytes},
			wantErr: false,
		},
		{
			name: "datetime",
			column: schema.Column{
				DataType:   "datetime",
				ColumnType: "datetime",
				Nullable:   "NO",
			},
			want: Desc{
				Type:    "time.Time",
				Kind:    KindTime,
				Imports: []string{"time"},
			},
			wantErr: false,
		},
		{
			name: "json",
			column: schema.Column{
				DataType:   "json",
				ColumnType: "json",
				Nullable:   "YES",
			},
			want: Desc{
				Type:    "json.RawMessage",
				Kind:    KindBytes,
				Imports: []string{"encoding/json"},
			},
			wantErr: false,
		},
		{
			name: "enum",
			column: schema.Column{
				DataType:   "enum",
				ColumnType: "enum('a','b')",
				Nullable:   "NO",
			},
			want:    Desc{Type: "string", Kind: KindEnum},
			wantErr: false,
		},
		{
			name: "unsupported type",
			column: schema.Column{
				DataType:   "unsupported_type",
				ColumnType: "unsupported_type",
				Nullable:   "NO",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MySQLDescFunc(&tt.column)
			if (err != nil) != tt.wantErr {
				t.Errorf("MySQLDescFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.want.Type || got.Kind != tt.want.Kind {
					t.Errorf("MySQLDescFunc() = %v, want %v", got, tt.want)
				}
			}
			if tt.wantErr && !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("MySQLDescFunc() error = %v, want ErrUnsupportedType", err)
			}
		})
	}
}

func TestPostgresDescFunc(t *testing.T) {
	tests := []struct {
		name    string
		column  schema.Column
		want    Desc
		wantErr bool
	}{
		{
			name: "int2",
			column: schema.Column{
				DataType:   "int2",
				ColumnType: "int2",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int16", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "int4",
			column: schema.Column{
				DataType:   "int4",
				ColumnType: "int4",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int32", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "int8",
			column: schema.Column{
				DataType:   "int8",
				ColumnType: "int8",
				Nullable:   "NO",
			},
			want:    Desc{Type: "int64", Kind: KindNumeric},
			wantErr: false,
		},
		{
			name: "bool",
			column: schema.Column{
				DataType:   "bool",
				ColumnType: "bool",
				Nullable:   "YES",
			},
			want:    Desc{Type: "*bool", Kind: KindBoolean},
			wantErr: false,
		},
		{
			name: "text",
			column: schema.Column{
				DataType:   "text",
				ColumnType: "text",
				Nullable:   "NO",
			},
			want:    Desc{Type: "string", Kind: KindString},
			wantErr: false,
		},
		{
			name: "varchar",
			column: schema.Column{
				DataType:   "varchar",
				ColumnType: "varchar(255)",
				Nullable:   "YES",
			},
			want:    Desc{Type: "*string", Kind: KindString},
			wantErr: false,
		},
		{
			name: "bytea",
			column: schema.Column{
				DataType:   "bytea",
				ColumnType: "bytea",
				Nullable:   "NO",
			},
			want:    Desc{Type: "[]byte", Kind: KindBytes},
			wantErr: false,
		},
		{
			name: "timestamp",
			column: schema.Column{
				DataType:   "timestamp",
				ColumnType: "timestamp",
				Nullable:   "NO",
			},
			want: Desc{
				Type:    "time.Time",
				Kind:    KindTime,
				Imports: []string{"time"},
			},
			wantErr: false,
		},
		{
			name: "json",
			column: schema.Column{
				DataType:   "json",
				ColumnType: "json",
				Nullable:   "YES",
			},
			want:    Desc{Type: "json.RawMessage", Kind: KindBytes},
			wantErr: false,
		},
		{
			name: "jsonb",
			column: schema.Column{
				DataType:   "jsonb",
				ColumnType: "jsonb",
				Nullable:   "NO",
			},
			want:    Desc{Type: "json.RawMessage", Kind: KindBytes},
			wantErr: false,
		},
		{
			name: "uuid",
			column: schema.Column{
				DataType:   "uuid",
				ColumnType: "uuid",
				Nullable:   "YES",
			},
			want: Desc{
				Type:    "*uuid.UUID",
				Kind:    KindStruct,
				Imports: []string{"github.com/google/uuid"},
			},
			wantErr: false,
		},
		{
			name: "inet",
			column: schema.Column{
				DataType:   "inet",
				ColumnType: "inet",
				Nullable:   "NO",
			},
			want: Desc{
				Type:    "net.IP",
				Kind:    KindStruct,
				Imports: []string{"net"},
			},
			wantErr: false,
		},
		{
			name: "interval",
			column: schema.Column{
				DataType:   "interval",
				ColumnType: "interval",
				Nullable:   "YES",
			},
			want: Desc{
				Type:    "*time.Duration",
				Kind:    KindTime,
				Imports: []string{"time"},
			},
			wantErr: false,
		},
		{
			name: "array text",
			column: schema.Column{
				DataType:   "ARRAY",
				ColumnType: "_text",
				Nullable:   "NO",
			},
			want:    Desc{Type: "[]string", Kind: KindStruct},
			wantErr: false,
		},
		{
			name: "array int4",
			column: schema.Column{
				DataType:   "ARRAY",
				ColumnType: "_int4",
				Nullable:   "NO",
			},
			want:    Desc{Type: "[]int32", Kind: KindStruct},
			wantErr: false,
		},
		{
			name: "unsupported type",
			column: schema.Column{
				DataType:   "unsupported_type",
				ColumnType: "unsupported_type",
				Nullable:   "NO",
			},
			wantErr: true,
		},
		{
			name: "unsupported array type",
			column: schema.Column{
				DataType:   "ARRAY",
				ColumnType: "_unsupported",
				Nullable:   "NO",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PostgresDescFunc(&tt.column)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresDescFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.want.Type || got.Kind != tt.want.Kind {
					t.Errorf("PostgresDescFunc() = %v, want %v", got, tt.want)
				}
			}
			if tt.wantErr && !errors.Is(err, ErrUnsupportedType) && !errors.Is(err, ErrUnsupportedArrayType) {
				t.Errorf("PostgresDescFunc() error = %v, want ErrUnsupportedType or ErrUnsupportedArrayType", err)
			}
		})
	}
}

func TestDescFuncs(t *testing.T) {
	if len(DescFuncs) != 2 {
		t.Errorf("DescFuncs length = %d, want 2", len(DescFuncs))
	}

	if _, ok := DescFuncs["mysql"]; !ok {
		t.Error("DescFuncs should contain mysql")
	}

	if _, ok := DescFuncs["postgres"]; !ok {
		t.Error("DescFuncs should contain postgres")
	}
}
