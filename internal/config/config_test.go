package config

import (
	"errors"
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
			err := tt.output.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Output.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateValidate(t *testing.T) {
	tests := []struct {
		name     string
		template *Template
		wantErr  bool
	}{
		{
			name: "valid template",
			template: &Template{
				Model: "model",
			},
			wantErr: false,
		},
		{
			name: "empty model",
			template: &Template{
				Model: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Template.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
			wantErr: true,
		},
		{
			name: "unsupported driver",
			db: &DB{
				Driver: "sqlite",
				DSN:    "file.db",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.db.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigNaming(t *testing.T) {
	cfg := &Config{
		Initialisms: []string{"ID", "URL", "API"},
	}

	result := cfg.Naming()

	if len(result) != 3 {
		t.Errorf("Naming() len = %d, want 3", len(result))
	}

	if result["id"] != "ID" {
		t.Errorf("Naming() id = %v, want ID", result["id"])
	}

	if result["url"] != "URL" {
		t.Errorf("Naming() url = %v, want URL", result["url"])
	}

	if result["api"] != "API" {
		t.Errorf("Naming() api = %v, want API", result["api"])
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
			wantErr: true,
		},
		{
			name: "empty type",
			field: &CustomField{
				Name: "CustomField",
				Type: "",
			},
			wantErr: true,
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
			err := tt.field.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CustomField.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
