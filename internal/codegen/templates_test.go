package codegen

import (
	"strings"
	"testing"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
)

// TestModelTemplatesCastDate guards the cross-driver form of DateExpr.
// DATE(col) is MySQL-only; CAST(col AS DATE) works on both MySQL and PostgreSQL.
func TestModelTemplatesCastDate(t *testing.T) {
	db := &config.DB{Driver: "postgres"}
	cfg := &config.Config{
		Output:   &config.Output{Dir: "model", Package: "model", Module: "github.com/x/y"},
		Timezone: "UTC",
	}
	table := &schema.Table{
		Name: "mails",
		Columns: []*schema.Column{
			{Name: "id", DataType: "int8", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
			{Name: "created_at", DataType: "timestamptz", ColumnType: "timestamptz", Nullable: "NO"},
		},
	}
	m, err := NewModel(table, db, cfg, naming.New())
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}
	tmpl, err := parseTemplates("")
	if err != nil {
		t.Fatalf("parseTemplates() error = %v", err)
	}

	tests := []struct {
		template string
		want     string
	}{
		{
			template: "model",
			want:     `MailCreatedAtDateExpr = "CAST(created_at AS DATE)"`,
		},
		{
			template: "model.alias",
			want:     `MailAliasCreatedAtDateExpr = "CAST(mails_alias.created_at AS DATE)"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			buf, err := tmpl.render(tt.template, m)
			if err != nil {
				t.Fatalf("render(%s) error = %v", tt.template, err)
			}
			out := buf.String()
			if !strings.Contains(out, tt.want) {
				t.Errorf("render(%s) does not contain %q", tt.template, tt.want)
			}
			if strings.Contains(out, `"DATE(`) {
				t.Errorf("render(%s) still emits MySQL-only DATE()", tt.template)
			}
		})
	}
}
