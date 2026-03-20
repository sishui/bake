package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultConfigFile = "bake.gen.yaml"
	defaultModelTmpl  = "model"
	defaultOutputDir  = "model"
	defaultPackage    = "model"
	defaultModule     = "github.com/username/project"
	defaultDSN        = "root:password@tcp(127.0.0.1:3306)/bake?charset=utf8mb4&parseTime=True&loc=Local"
	defaultDriver     = "mysql"
	defaultSchema     = "public"
)

const defaultConfigContent = `
#go install github.com/sishui/bake/cmd/bake@latest
#bake init

log:
  file: ""

uncountables: ["sms", "mms", "rls"]

initialisms: ["ID", "URL", "URI", "UUID", "IP"]

timezone: ""

template:
  dir: ""
  model: "%s"
output:
  dir: "%s"
  package: "%s"
  module: "%s"
db:
  - driver: "%s"
    dsn: "%s"%s
    included: []
    excluded: []
`

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: fmt.Sprintf("init '%s' config file", defaultConfigFile),
		RunE:  runInitCommand,
	}
	flags := cmd.Flags()
	flags.StringP("template", "t", defaultModelTmpl, "model template filename(no extension)")
	flags.StringP("output", "o", defaultOutputDir, fmt.Sprintf("%s output dir", defaultConfigFile))
	flags.StringP("package", "p", defaultPackage, "model gen package name")
	flags.StringP("module", "m", defaultModule, "module name")
	flags.StringP("driver", "d", defaultDriver, "db driver: support mysql, postgres")
	flags.StringP("dsn", "n", defaultDSN, "db dsn")
	flags.StringP("schema", "s", defaultSchema, "db schema")

	return cmd
}

func runInitCommand(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	template, err := flags.GetString("template")
	if err != nil {
		return err
	}
	output, err := flags.GetString("output")
	if err != nil {
		return err
	}
	pkg, err := flags.GetString("package")
	if err != nil {
		return err
	}
	module, err := flags.GetString("module")
	if err != nil {
		return err
	}
	driver, err := flags.GetString("driver")
	if err != nil {
		return err
	}
	dsn, err := flags.GetString("dsn")
	if err != nil {
		return err
	}
	schema, err := flags.GetString("schema")
	if err != nil {
		return err
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		if err := os.MkdirAll(output, 0o755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}
	if _, err := os.Stat(defaultConfigFile); err == nil {
		return fmt.Errorf("%s already exists", defaultConfigFile)
	}
	if driver == "postgres" {
		schema = fmt.Sprintf("\n    schema: \"%s\"", schema)
	} else {
		schema = ""
	}
	content := fmt.Sprintf(defaultConfigContent, template, output, pkg, module, driver, dsn, schema)
	if err := os.WriteFile(defaultConfigFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
