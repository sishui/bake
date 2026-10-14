package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const parseFullConfig = `
output:
  dir: "model"
  package: "model"
  module: "github.com/user/project"
log:
  level: "warn"
db:
  - driver: "postgres"
    dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
`

const parseNoSectionsConfig = `
db:
  - driver: "mysql"
    dsn: "user:pass@tcp(127.0.0.1:3306)/db"
`

func writeTempConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "bake.gen.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

func TestParse(t *testing.T) {
	const mysqlDSN = "root:pw@tcp(127.0.0.1:3306)/db?charset=utf8mb4"

	tests := []struct {
		name    string
		body    string
		args    func(configFile string) Args
		wantErr string
		check   func(t *testing.T, c *Config)
	}{
		{
			name: "file values are used as-is",
			body: parseFullConfig,
			args: func(f string) Args { return Args{Config: f} },
			check: func(t *testing.T, c *Config) {
				if c.Output.Dir != "model" || c.Output.Package != "model" {
					t.Errorf("output = %+v", c.Output)
				}
				if c.Log.Level != "warn" {
					t.Errorf("log level = %q, want %q", c.Log.Level, "warn")
				}
				if len(c.DB) != 1 || c.DB[0].Driver != "postgres" ||
					c.DB[0].DSN != "postgres://user:pass@localhost:5432/db?sslmode=disable" {
					t.Errorf("db = %+v", c.DB)
				}
				if c.Filename == "" {
					t.Error("Filename not set")
				}
			},
		},
		{
			name: "output flag sets dir and package falls back to dir base",
			body: parseNoSectionsConfig,
			args: func(f string) Args { return Args{Config: f, Output: "gen/models"} },
			check: func(t *testing.T, c *Config) {
				if c.Output.Dir != "gen/models" {
					t.Errorf("dir = %q, want %q", c.Output.Dir, "gen/models")
				}
				if c.Output.Package != "models" {
					t.Errorf("package = %q, want %q", c.Output.Package, "models")
				}
			},
		},
		{
			name: "package flag wins over dir base",
			body: parseNoSectionsConfig,
			args: func(f string) Args {
				return Args{Config: f, Output: "gen/models", Package: "entities"}
			},
			check: func(t *testing.T, c *Config) {
				if c.Output.Package != "entities" {
					t.Errorf("package = %q, want %q", c.Output.Package, "entities")
				}
			},
		},
		{
			name: "file package is kept when no package flag",
			body: parseFullConfig,
			args: func(f string) Args { return Args{Config: f, Output: "other"} },
			check: func(t *testing.T, c *Config) {
				if c.Output.Dir != "other" {
					t.Errorf("dir = %q, want %q", c.Output.Dir, "other")
				}
				if c.Output.Package != "model" {
					t.Errorf("package = %q, want %q", c.Output.Package, "model")
				}
			},
		},
		{
			name:    "dsn is required when the file defines no db",
			body:    "output:\n  dir: \"model\"\n",
			args:    func(f string) Args { return Args{Config: f} },
			wantErr: "db.dsn is required",
		},
		{
			name: "dsn flag creates the db entry",
			body: "output:\n  dir: \"model\"\n",
			args: func(f string) Args { return Args{Config: f, DSN: mysqlDSN} },
			check: func(t *testing.T, c *Config) {
				if len(c.DB) != 1 {
					t.Fatalf("db = %+v", c.DB)
				}
				if c.DB[0].Driver != "mysql" || c.DB[0].DSN != mysqlDSN {
					t.Errorf("db[0] = %+v", c.DB[0])
				}
			},
		},
		{
			name: "dsn flag overrides the first db entry",
			body: parseFullConfig,
			args: func(f string) Args { return Args{Config: f, DSN: mysqlDSN} },
			check: func(t *testing.T, c *Config) {
				if len(c.DB) != 1 {
					t.Fatalf("db = %+v", c.DB)
				}
				if c.DB[0].Driver != "mysql" || c.DB[0].DSN != mysqlDSN {
					t.Errorf("db[0] = %+v", c.DB[0])
				}
			},
		},
		{
			name: "file db is kept when no dsn flag",
			body: parseFullConfig,
			args: func(f string) Args { return Args{Config: f} },
			check: func(t *testing.T, c *Config) {
				if len(c.DB) != 1 || c.DB[0].Driver != "postgres" {
					t.Errorf("db = %+v", c.DB)
				}
			},
		},
		{
			name:    "undetectable dsn is rejected",
			body:    "output:\n  dir: \"model\"\n",
			args:    func(f string) Args { return Args{Config: f, DSN: "not-a-dsn"} },
			wantErr: "cannot detect driver",
		},
		{
			name: "verbose forces the debug level",
			body: parseFullConfig,
			args: func(f string) Args { return Args{Config: f, Verbose: true} },
			check: func(t *testing.T, c *Config) {
				if c.Log.Level != "debug" {
					t.Errorf("log level = %q, want %q", c.Log.Level, "debug")
				}
			},
		},
		{
			name: "missing sections get defaults",
			body: parseNoSectionsConfig,
			args: func(f string) Args { return Args{Config: f} },
			check: func(t *testing.T, c *Config) {
				if c.Output == nil || c.Template == nil || c.Log == nil {
					t.Fatalf("nil section: output=%v template=%v log=%v", c.Output, c.Template, c.Log)
				}
				if c.Log.Level != "info" {
					t.Errorf("log level = %q, want %q", c.Log.Level, "info")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := writeTempConfig(t, tt.body)
			cfg, err := Parse(tt.args(file))

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Parse() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			tt.check(t, cfg)
		})
	}
}

func TestParse_MissingConfigFile(t *testing.T) {
	_, err := Parse(Args{Config: filepath.Join(t.TempDir(), "absent.yaml")})
	if err == nil {
		t.Fatal("Parse() error = nil, want an error for a missing config file")
	}
}
