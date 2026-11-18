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

func TestModelTemplatesTemporalExpressions(t *testing.T) {
	tests := []struct {
		name    string
		driver  string
		table   string
		columns []*schema.Column
		want    []string
		notWant []string
	}{
		{
			name:   "mysql date time timestamp",
			driver: "mysql",
			table:  "events",
			columns: []*schema.Column{
				{Name: "id", DataType: "bigint", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
				{Name: "date_val", DataType: "date", ColumnType: "date", Nullable: "NO"},
				{Name: "time_val", DataType: "time", ColumnType: "time", Nullable: "NO"},
				{Name: "timestamp_val", DataType: "timestamp", ColumnType: "timestamp", Nullable: "NO"},
			},
			want: []string{
				`EventDateValDateExpr`,
				`"CAST(date_val AS DATE)"`,
				`EventTimestampValDateExpr`,
				`"CAST(timestamp_val AS DATE)"`,
				`EventDateValYearExpr`,
				`EventTimestampValYearExpr`,
				`EventTimeValHourExpr`,
				`EventTimestampValHourExpr`,
				`EventAliasDateValDateExpr`,
				`EventAliasTimeValHourExpr`,
			},
			notWant: []string{
				`EventTimeValDateExpr`,
				`EventTimeValYearExpr`,
				`EventTimeValMonthExpr`,
				`EventTimeValDayExpr`,
				`EventDateValHourExpr`,
				`EventDateValMinuteExpr`,
				`EventDateValSecondExpr`,
				`EventAliasTimeValDateExpr`,
				`EventAliasDateValHourExpr`,
			},
		},
		{
			name:   "postgres date time timestamp interval",
			driver: "postgres",
			table:  "events",
			columns: []*schema.Column{
				{Name: "id", DataType: "int8", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
				{Name: "date_val", DataType: "date", ColumnType: "date", Nullable: "NO"},
				{Name: "time_val", DataType: "time", ColumnType: "time", Nullable: "NO"},
				{Name: "timestamp_val", DataType: "timestamp", ColumnType: "timestamp", Nullable: "NO"},
				{Name: "interval_val", DataType: "interval", ColumnType: "interval", Nullable: "NO"},
			},
			want: []string{
				`EventDateValDateExpr`,
				`"CAST(date_val AS DATE)"`,
				`EventTimestampValDateExpr`,
				`"CAST(timestamp_val AS DATE)"`,
				`EventTimeValHourExpr`,
				`EventAliasTimestampValDateExpr`,
			},
			notWant: []string{
				`EventTimeValDateExpr`,
				`EventTimeValYearExpr`,
				`EventDateValHourExpr`,
				`EventIntervalValDateExpr`,
				`EventIntervalValYearExpr`,
				`EventIntervalValHourExpr`,
			},
		},
		{
			name:   "postgres time only",
			driver: "postgres",
			table:  "clocks",
			columns: []*schema.Column{
				{Name: "id", DataType: "int8", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
				{Name: "time_val", DataType: "time", ColumnType: "time", Nullable: "NO"},
			},
			want: []string{
				`ClockTimeValHourExpr`,
				`ClockAliasTimeValHourExpr`,
			},
			notWant: []string{
				`ClockTimeValDateExpr`,
				`ClockTimeValYearExpr`,
				`ClockTimeValMonthExpr`,
				`ClockTimeValDayExpr`,
				`const ()`,
			},
		},
		{
			name:   "postgres date only",
			driver: "postgres",
			table:  "birthdays",
			columns: []*schema.Column{
				{Name: "id", DataType: "int8", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
				{Name: "date_val", DataType: "date", ColumnType: "date", Nullable: "NO"},
			},
			want: []string{
				`BirthdayDateValDateExpr`,
				`BirthdayDateValYearExpr`,
			},
			notWant: []string{
				`BirthdayDateValHourExpr`,
				`BirthdayDateValMinuteExpr`,
				`BirthdayDateValSecondExpr`,
				`const ()`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderTemporalTemplates(t, tt.driver, tt.table, tt.columns)
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("rendered templates do not contain %q", want)
				}
			}
			for _, notWant := range tt.notWant {
				if strings.Contains(out, notWant) {
					t.Errorf("rendered templates contain %q", notWant)
				}
			}
		})
	}
}

func renderTemporalTemplates(t *testing.T, driver string, tableName string, columns []*schema.Column) string {
	t.Helper()
	cfg := &config.Config{
		Output:   &config.Output{Dir: "model", Package: "model", Module: "github.com/x/y"},
		Timezone: "UTC",
	}
	table := &schema.Table{Name: tableName, Columns: columns}
	m, err := NewModel(table, &config.DB{Driver: driver}, cfg, naming.New())
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}
	tmpl, err := parseTemplates("")
	if err != nil {
		t.Fatalf("parseTemplates() error = %v", err)
	}
	var out strings.Builder
	for _, name := range []string{"model", "model.alias"} {
		buf, err := tmpl.render(name, m)
		if err != nil {
			t.Fatalf("render(%s) error = %v", name, err)
		}
		out.WriteString(buf.String())
	}
	return out.String()
}

func TestModelTemplatesSUMOrZero(t *testing.T) {
	db := &config.DB{
		Driver: "mysql",
		Custom: map[string]*config.CustomTable{
			"orders": {
				Fields: map[string]*config.CustomField{
					"bonus": {Type: "int64"},
				},
			},
		},
	}
	cfg := &config.Config{
		Output:   &config.Output{Dir: "model", Package: "model", Module: "github.com/x/y"},
		Timezone: "UTC",
	}
	table := &schema.Table{
		Name: "orders",
		Columns: []*schema.Column{
			{Name: "id", DataType: "bigint", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
			{Name: "amount", DataType: "decimal", ColumnType: "decimal(10,2)", Nullable: "NO"},
			{Name: "year_val", DataType: "year", ColumnType: "year", Nullable: "NO"},
			{Name: "created_at", DataType: "datetime", ColumnType: "datetime", Nullable: "NO"},
			{Name: "name", DataType: "varchar", ColumnType: "varchar(20)", Nullable: "NO"},
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
	var out strings.Builder
	for _, name := range []string{"model", "model.alias"} {
		buf, err := tmpl.render(name, m)
		if err != nil {
			t.Fatalf("render(%s) error = %v", name, err)
		}
		out.WriteString(buf.String())
	}
	rendered := out.String()

	want := []string{
		`OrderAmountSUMExpr`,
		`OrderAmountSUMOrZeroExpr`,
		`COALESCE(SUM(amount), 0)`,
		`OrderAliasAmountSUMOrZeroExpr`,
		`COALESCE(SUM(orders_alias.amount), 0)`,
	}
	for _, w := range want {
		if !strings.Contains(rendered, w) {
			t.Errorf("rendered templates do not contain %q", w)
		}
	}
	notWant := []string{
		`OrderIDSUMOrZeroExpr`,
		`OrderYearValSUMOrZeroExpr`,
		`OrderCreatedAtSUMOrZeroExpr`,
		`OrderNameSUMOrZeroExpr`,
		`OrderBonusSUMExpr`,
		`OrderBonusSUMOrZeroExpr`,
		`OrderAliasBonusSUMExpr`,
		`OrderAliasBonusSUMOrZeroExpr`,
	}
	for _, w := range notWant {
		if strings.Contains(rendered, w) {
			t.Errorf("rendered templates contain %q", w)
		}
	}
	if got := strings.Count(rendered, "SUMOrZeroExpr"); got != 2 {
		t.Errorf("SUMOrZeroExpr count = %d, want 2 (base and alias)", got)
	}

	// A custom numeric field alone must not emit SUMOrZeroExpr, and its absence
	// must not leave an empty const block behind.
	customOnlyDB := &config.DB{
		Driver: "mysql",
		Custom: map[string]*config.CustomTable{
			"ledgers": {
				Fields: map[string]*config.CustomField{
					"bonus": {Type: "int64"},
				},
			},
		},
	}
	customOnly := &schema.Table{
		Name: "ledgers",
		Columns: []*schema.Column{
			{Name: "id", DataType: "bigint", ColumnType: "bigint", Nullable: "NO", Key: schema.ColumnKeyPrimary},
			{Name: "name", DataType: "varchar", ColumnType: "varchar(20)", Nullable: "NO"},
		},
	}
	m2, err := NewModel(customOnly, customOnlyDB, cfg, naming.New())
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}
	if m2.HasArithmetic {
		t.Error("HasArithmetic = true, want false for a custom-only numeric field")
	}
	var customOnlyOut strings.Builder
	for _, name := range []string{"model", "model.alias"} {
		buf, err := tmpl.render(name, m2)
		if err != nil {
			t.Fatalf("render(%s) error = %v", name, err)
		}
		customOnlyOut.WriteString(buf.String())
	}
	for _, expr := range []string{
		"SUMExpr", "SUMOrZeroExpr", "AVGExpr",
		"AddExpr", "SubExpr", "MulExpr", "DivExpr",
	} {
		if strings.Contains(customOnlyOut.String(), expr) {
			t.Errorf("rendered custom-only model contains %s", expr)
		}
	}
	if strings.Contains(customOnlyOut.String(), "const ()") {
		t.Error("rendered custom-only model contains an empty const block")
	}
}
