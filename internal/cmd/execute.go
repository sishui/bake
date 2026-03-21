// Package cmd contains the CLI commands for the application.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/generate"
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
	configFile, err := flags.GetString("config")
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
	dsn, err := flags.GetString("dsn")
	if err != nil {
		return err
	}
	verbose, err := flags.GetBool("verbose")
	if err != nil {
		return err
	}
	cfg, err := config.Parse(config.Args{
		Config:  configFile,
		Output:  output,
		Package: pkg,
		DSN:     dsn,
		Verbose: verbose,
	})
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	cfg.Version = version
	return generate.Run(cfg)
}
