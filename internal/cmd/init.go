package cmd

import (
	"fmt"
	"os"
	"strings"

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
  dir: "{{template}}" # template dir(if not set, use default template)
output:
  dir: "{{output}}"
  package: "{{package}}"
  module: "{{module}}"
custom: [] # custom struct
db:
  - driver: "{{driver}}"
    dsn: "{{dsn}}"{{schema}}
    include: []
    exclude: []
    custom: {}
`

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: fmt.Sprintf("init '%s' config file", defaultConfigFile),
		RunE:  runInitCommand,
	}
	flags := cmd.Flags()
	flags.StringP("template", "t", defaultModelTmpl, "model template dir")
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
	template := mustGetString(flags, "template")
	output := mustGetString(flags, "output")
	pkg := mustGetString(flags, "package")
	module := mustGetString(flags, "module")
	driver := mustGetString(flags, "driver")
	dsn := mustGetString(flags, "dsn")
	schema := mustGetString(flags, "schema")

	if _, err := os.Stat(output); os.IsNotExist(err) {
		if err := os.MkdirAll(output, 0o755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}
	if _, err := os.Stat(defaultConfigFile); err == nil {
		return fmt.Errorf("%s already exists", defaultConfigFile)
	}
	content := defaultConfigContent[1:]
	content = strings.ReplaceAll(content, "{{template}}", template)
	content = strings.ReplaceAll(content, "{{output}}", output)
	content = strings.ReplaceAll(content, "{{package}}", pkg)
	content = strings.ReplaceAll(content, "{{module}}", module)
	content = strings.ReplaceAll(content, "{{driver}}", driver)
	content = strings.ReplaceAll(content, "{{dsn}}", dsn)
	if driver == "postgres" {
		content = strings.ReplaceAll(content, "{{schema}}", fmt.Sprintf("\n    schema: \"%s\"", schema))
	} else {
		content = strings.ReplaceAll(content, "{{schema}}", "")
	}
	if err := os.WriteFile(defaultConfigFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
