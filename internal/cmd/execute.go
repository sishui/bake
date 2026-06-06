// Package cmd contains the CLI commands for the application.
package cmd

import (
	"fmt"
	"os"

	"github.com/sishui/bake/internal/codegen"
	"github.com/sishui/bake/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bake",
		Short: "generate bun orm models from database schema",
		RunE:  runRootCommand,
	}

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewInitCommand())

	return cmd
}

func Execute() {
	rootCmd := newRootCommand()
	flags := rootCmd.Flags()
	flags.StringP("config", "c", defaultConfigFile, "config file")
	flags.StringP("output", "o", "", "model output directory")
	flags.StringP("package", "p", "", "model package name")
	flags.StringP("dsn", "n", "", "data source name")
	flags.BoolP("verbose", "v", false, "verbose output")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runRootCommand(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	cfg, err := config.Parse(config.Args{
		Config:  mustGetString(flags, "config"),
		Output:  mustGetString(flags, "output"),
		Package: mustGetString(flags, "package"),
		DSN:     mustGetString(flags, "dsn"),
		Verbose: mustGetBool(flags, "verbose"),
	})
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	cfg.Version = version
	return codegen.Run(cfg)
}

func mustGetString(flags *pflag.FlagSet, name string) string {
	v, err := flags.GetString(name)
	if err != nil {
		panic(fmt.Sprintf("flag %s: %v", name, err))
	}
	return v
}

func mustGetBool(flags *pflag.FlagSet, name string) bool {
	v, err := flags.GetBool(name)
	if err != nil {
		panic(fmt.Sprintf("flag %s: %v", name, err))
	}
	return v
}
