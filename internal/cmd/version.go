package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v0.0.1"

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show bake version",
		Run:   runVersion,
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Println("bake version:", version)
}
