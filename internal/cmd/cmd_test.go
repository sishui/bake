package cmd_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/sishui/bake/internal/cmd"
)

func TestRunVersion(t *testing.T) {
	verCmd := cmd.NewVersionCommand()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	verCmd.Run(verCmd, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	r.Close()

	got := buf.String()
	if !strings.Contains(got, "bake version:") {
		t.Errorf("runVersion output = %q, want substring %q", got, "bake version:")
	}
	if !strings.Contains(got, "v0.3.0") {
		t.Errorf("runVersion output = %q, want substring %q", got, "v0.3.0")
	}
}

func TestNewVersionCommand(t *testing.T) {
	c := cmd.NewVersionCommand()

	tests := []struct {
		field string
		want  string
	}{
		{"Use", "version"},
		{"Short", "show bake version"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = c.Use
			case "Short":
				got = c.Short
			}
			if got != tt.want {
				t.Errorf("NewVersionCommand().%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}

	if c.Run == nil {
		t.Error("NewVersionCommand().Run is nil, want non-nil")
	}
}

func TestNewInitCommand_Flags(t *testing.T) {
	c := cmd.NewInitCommand()

	flagTests := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"template", "t", "model"},
		{"output", "o", "model"},
		{"package", "p", "model"},
		{"module", "m", "github.com/username/project"},
		{"driver", "d", "mysql"},
		{"dsn", "n", "root:password@tcp(127.0.0.1:3306)/bake?charset=utf8mb4&parseTime=True&loc=Local"},
		{"schema", "s", "public"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			f := c.Flags().Lookup(tt.name)
			if f == nil {
				t.Fatalf("flag %q not registered", tt.name)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.name, f.Shorthand, tt.shorthand)
			}
			if f.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, f.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestNewInitCommand_Properties(t *testing.T) {
	c := cmd.NewInitCommand()

	if c.Use != "init" {
		t.Errorf("NewInitCommand().Use = %q, want %q", c.Use, "init")
	}
	if c.RunE == nil {
		t.Error("NewInitCommand().RunE is nil, want non-nil")
	}
}

func TestNewRootCommand_Subcommands(t *testing.T) {
	root := newTestRootCommand()
	subcommands := root.Commands()

	names := make(map[string]struct{}, len(subcommands))
	for _, sub := range subcommands {
		names[sub.Name()] = struct{}{}
	}

	for _, want := range []string{"version", "init"} {
		if _, ok := names[want]; !ok {
			present := make([]string, 0, len(subcommands))
			for _, sub := range subcommands {
				present = append(present, sub.Name())
			}
			t.Errorf("root command missing subcommand %q; got %v", want, present)
		}
	}
}

// newTestRootCommand replicates the root command structure from the
// unexported newRootCommand so tests can inspect subcommands.
func newTestRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "bake",
	}
	root.AddCommand(cmd.NewVersionCommand())
	root.AddCommand(cmd.NewInitCommand())
	return root
}
