package config

import (
	"errors"
	"strings"
	"testing"
)

func TestOutputValidate(t *testing.T) {
	tests := []struct {
		name    string
		output  *Output
		wantErr bool
		errType error
	}{
		{
			name: "valid output",
			output: &Output{
				Dir:     "model",
				Package: "model",
				Module:  "github.com/test/module",
			},
			wantErr: false,
		},
		{
			name: "empty dir",
			output: &Output{
				Dir:     "",
				Package: "model",
				Module:  "github.com/test/module",
			},
			wantErr: true,
			errType: errors.New("output.dir is required"),
		},
		{
			name: "invalid dir with spaces",
			output: &Output{
				Dir:     "model dir",
				Package: "model",
				Module:  "github.com/test/module",
			},
			wantErr: true,
		},
		{
			name: "empty package",
			output: &Output{
				Dir:     "model",
				Package: "",
				Module:  "github.com/test/module",
			},
			wantErr: true,
		},
		{
			name: "package with uppercase",
			output: &Output{
				Dir:     "model",
				Package: "Model",
				Module:  "github.com/test/module",
			},
			wantErr: true,
		},
		{
			name: "package with underscore",
			output: &Output{
				Dir:     "model",
				Package: "model_name",
				Module:  "github.com/test/module",
			},
			wantErr: true,
		},
		{
			name: "invalid module path",
			output: &Output{
				Dir:     "model",
				Package: "model",
				Module:  "-invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.output.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Output.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateValidate(t *testing.T) {
	testTemplate := &Template{}
	if err := testTemplate.Validate(); err != nil {
		t.Errorf("Template.Validate() error = %v, want nil", err)
	}
}

func TestDBValidate(t *testing.T) {
	tests := []struct {
		name    string
		db      *DB
		wantErr bool
	}{
		{
			name: "valid mysql",
			db: &DB{
				Driver: "mysql",
				DSN:    "user:pass@tcp(localhost:3306)/db",
				Schema: "test",
			},
			wantErr: false,
		},
		{
			name: "valid postgres",
			db: &DB{
				Driver: "postgres",
				DSN:    "postgres://user:pass@localhost:5432/db",
				Schema: "public",
			},
			wantErr: false,
		},
		{
			name: "empty driver",
			db: &DB{
				Driver: "",
				DSN:    "user:pass@tcp(localhost:3306)/db",
			},
			wantErr: true,
		},
		{
			name: "empty dsn",
			db: &DB{
				Driver: "mysql",
				DSN:    "",
			},
			wantErr: true,
		},
		{
			name: "postgres without schema",
			db: &DB{
				Driver: "postgres",
				DSN:    "postgres://user:pass@localhost:5432/db",
				Schema: "",
			},
			wantErr: false,
		},
		{
			name: "any driver name is valid at config level",
			db: &DB{
				Driver: "sqlite",
				DSN:    "file.db",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.db.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomFieldValidate(t *testing.T) {
	tests := []struct {
		name    string
		field   *CustomField
		wantErr bool
	}{
		{
			name: "valid field",
			field: &CustomField{
				Name: "CustomField",
				Type: "string",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			field: &CustomField{
				Name: "",
				Type: "string",
			},
			wantErr: false,
		},
		{
			name: "empty type",
			field: &CustomField{
				Name: "CustomField",
				Type: "",
			},
			wantErr: false,
		},
		{
			name: "invalid name (number start)",
			field: &CustomField{
				Name: "1Field",
				Type: "string",
			},
			wantErr: true,
		},
		{
			name: "keyword as name",
			field: &CustomField{
				Name: "func",
				Type: "string",
			},
			wantErr: true,
		},
		{
			name: "empty type",
			field: &CustomField{
				Name: "CustomField",
				Type: "CustomField",
			},
			wantErr: false,
		},
		{
			name: "compound type decimal.Decimal",
			field: &CustomField{
				Name: "Amount",
				Type: "decimal.Decimal",
			},
			wantErr: false,
		},
		{
			name: "slice type []string",
			field: &CustomField{
				Name: "Tags",
				Type: "[]string",
			},
			wantErr: false,
		},
		{
			name: "pointer type *User",
			field: &CustomField{
				Name: "User",
				Type: "*User",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.field.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CustomField.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImportValidate(t *testing.T) {
	tests := []struct {
		name    string
		imp     Import
		wantErr string
	}{
		{
			name: "plain package",
			imp:  Import{Package: "github.com/google/uuid", Types: []string{"uuid.UUID"}},
		},
		{
			name: "package whose name differs from the path",
			imp:  Import{Package: "github.com/hashicorp/go-version", Types: []string{"version.Version"}},
		},
		{
			name: "with alias",
			imp:  Import{Package: "github.com/shopspring/decimal", Alias: "dec", Types: []string{"dec.Decimal"}},
		},
		{
			name:    "unqualified type",
			imp:     Import{Package: "github.com/google/uuid", Types: []string{"UUID"}},
			wantErr: "must be qualified",
		},
		{
			name:    "type does not use the alias",
			imp:     Import{Package: "github.com/shopspring/decimal", Alias: "dec", Types: []string{"decimal.Decimal"}},
			wantErr: "must use the alias",
		},
		{
			name:    "missing package",
			imp:     Import{},
			wantErr: "import.package is required",
		},
		{
			name:    "package with a quote",
			imp:     Import{Package: `github.com/x/"y"`},
			wantErr: "contains invalid characters",
		},
		{
			name:    "invalid alias",
			imp:     Import{Package: "github.com/x/y", Alias: "1bad"},
			wantErr: "import.alias",
		},
		{
			name:    "invalid type",
			imp:     Import{Package: "github.com/x/y", Types: []string{"pkg.not a type"}},
			wantErr: "import.types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.imp.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidateChecksImports(t *testing.T) {
	c := &Config{
		Output:  &Output{Dir: "model", Package: "model", Module: "github.com/x/y"},
		Imports: []*Import{{Package: "github.com/x/y", Alias: "1bad"}},
	}
	if err := c.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error for the invalid import alias")
	}
}
