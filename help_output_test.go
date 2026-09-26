package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const helpOutputFile = "testdata/expected-help-output.txt"

type testAuthor struct {
	Name, Email string
}

type testStringerAuthor struct {
	Name string
}

func (a testStringerAuthor) String() string {
	return a.Name + " (stringer)"
}

func buildHelpOutputTestCommand() *Command {
	cmd := buildMinimalTestCommand()
	*cmd = Command{
		Name:    "app",
		Usage:   "a test application\nwith a multi-line usage",
		Version: "1.2.3",
		Description: "This is a long description of the application,\n" +
			"spanning multiple lines.\n\nAnd paragraphs.",
		Authors: []any{
			"Alice <alice@example.com>",
			testAuthor{Name: "Bob", Email: "bob@example.com"},
			&testAuthor{Name: "Carol", Email: "carol@example.com"},
			testStringerAuthor{Name: "Dave"},
			nil,
		},
		Copyright: "(c) 2026 Somebody\nAll rights reserved.",
		Arguments: []Argument{&StringArg{Name: "root-arg"}},
		Flags: []Flag{
			&StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "config `FILE`", Category: "Configuration"},
			&BoolFlag{Name: "debug", Usage: "enable debug output", Category: "Logging"},
			&StringFlag{Name: "log", Usage: "log file\nwith a multi-line usage", Category: "Logging"},
			&IntFlag{Name: "level", Value: 3, Usage: "some level"},
			&BoolFlag{Name: "hidden", Hidden: true},
		},
		Commands: []*Command{
			{
				Name:        "create",
				Aliases:     []string{"cr", "new"},
				Usage:       "create a thing",
				Category:    "Lifecycle",
				Description: "Creates a thing.\nWith a multi-line description.",
				Arguments: []Argument{
					&StringArg{Name: "id", Required: true},
					&StringArg{Name: "bundle"},
					&StringArg{Name: "extra", UsageText: "[extra...]"},
				},
				Flags: []Flag{
					&StringFlag{Name: "pid-file", Usage: "write pid to `FILE`", Local: true},
					&BoolFlag{Name: "detach", Aliases: []string{"d"}, Usage: "detach", Category: "Run options", Local: true},
				},
			},
			{
				Name:     "delete",
				Usage:    "delete a thing",
				Category: "Lifecycle",
				Flags: []Flag{
					&BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force deletion", Local: true},
				},
			},
			{
				Name:      "exec",
				Usage:     "execute a thing",
				ArgsUsage: "<id> <command> [args...]",
			},
			{
				Name:      "list",
				Aliases:   []string{"ls"},
				Usage:     "list things",
				UsageText: "app list [options]\n\napp ls",
			},
			{
				Name:   "hidden-command",
				Hidden: true,
			},
			{
				Name:      "plugin",
				Usage:     "manage plugins",
				UsageText: "app plugin <command>\n\napp plugin help",
				Category:  "Extensions",
				Commands: []*Command{
					{Name: "install", Usage: "install a plugin"},
				},
			},
			{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "manage configuration",
				Commands: []*Command{
					{
						Name:    "show",
						Aliases: []string{"s"},
						Usage:   "show configuration",
					},
					{
						Name:    "set-a-very-long-name",
						Aliases: []string{"set"},
						Usage:   "set\nconfiguration",
						Flags: []Flag{
							&BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "set globally", Local: true},
						},
					},
				},
			},
		},
		EnableShellCompletion: true,
	}
	return cmd
}

func buildHelpOutputMinimalCommand() *Command {
	cmd := buildMinimalTestCommand()
	cmd.Name = "mini"
	cmd.Version = "0.1"
	cmd.HideVersion = true
	cmd.Authors = []any{"Single Author"}
	cmd.HideHelp = true
	return cmd
}

func buildHelpOutputUsageTextCommand() *Command {
	cmd := buildMinimalTestCommand()
	cmd.Name = "ut"
	cmd.UsageText = "ut [options] <thing>\n\nut other-form"
	cmd.Flags = []Flag{&BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "assume yes"}}
	return cmd
}

func buildHelpOutputArgsUsageCommand() *Command {
	cmd := buildMinimalTestCommand()
	cmd.Name = "au"
	cmd.ArgsUsage = "<source> <destination>"
	cmd.HideHelp = true
	return cmd
}

// TestHelpOutput checks the output of help and fish completion against
// a golden file. It is run both with and without urfave_cli_no_template
// build tag, making sure both produce the exact same output.
//
// To regenerate the golden file, run with UPDATE_GOLDEN=1 (without the tag).
func TestHelpOutput(t *testing.T) {
	type testCase struct {
		name string
		cmd  func() *Command
		args []string
		wrap int
		// If set, call ShowRootCommandHelp directly instead of cmd.Run.
		direct bool
	}
	tests := []testCase{
		{name: "root", cmd: buildHelpOutputTestCommand, args: []string{"--help"}},
		{name: "root help command", cmd: buildHelpOutputTestCommand, args: []string{"help"}},
		{name: "command", cmd: buildHelpOutputTestCommand, args: []string{"create", "--help"}},
		{name: "help command", cmd: buildHelpOutputTestCommand, args: []string{"help", "create"}},
		{name: "command no categories", cmd: buildHelpOutputTestCommand, args: []string{"delete", "--help"}},
		{name: "command args usage", cmd: buildHelpOutputTestCommand, args: []string{"exec", "--help"}},
		{name: "command usage text", cmd: buildHelpOutputTestCommand, args: []string{"list", "--help"}},
		{name: "subcommand", cmd: buildHelpOutputTestCommand, args: []string{"config", "--help"}},
		{name: "help subcommand", cmd: buildHelpOutputTestCommand, args: []string{"help", "config"}},
		{name: "subcommand usage text and category", cmd: buildHelpOutputTestCommand, args: []string{"plugin", "--help"}},
		{name: "sub-subcommand", cmd: buildHelpOutputTestCommand, args: []string{"config", "set", "--help"}},
		{name: "root wrapped", cmd: buildHelpOutputTestCommand, args: []string{"--help"}, wrap: 40},
		{name: "command wrapped", cmd: buildHelpOutputTestCommand, args: []string{"create", "--help"}, wrap: 40},
		{name: "subcommand wrapped", cmd: buildHelpOutputTestCommand, args: []string{"config", "--help"}, wrap: 30},
		{name: "minimal", cmd: buildHelpOutputMinimalCommand, direct: true},
		{name: "usage text", cmd: buildHelpOutputUsageTextCommand, args: []string{"--help"}},
		{name: "args usage", cmd: buildHelpOutputArgsUsageCommand, direct: true},
	}

	var got strings.Builder
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wrap != 0 {
				defer func(old HelpPrinterFunc) { HelpPrinter = old }(HelpPrinter)
				HelpPrinter = func(w io.Writer, templ string, data any) {
					HelpPrinterCustom(w, templ, data, map[string]any{
						"wrapAt": func() int { return tc.wrap },
					})
				}
			}

			var out bytes.Buffer
			cmd := tc.cmd()
			cmd.Writer = &out
			cmd.ErrWriter = io.Discard
			if tc.direct {
				require.NoError(t, ShowRootCommandHelp(cmd))
			} else {
				require.NoError(t, cmd.Run(context.Background(), append([]string{cmd.Name}, tc.args...)))
			}

			got.WriteString("=== " + tc.name + ": " + strings.Join(tc.args, " ") + "\n")
			got.WriteString(out.String())
		})
	}

	fish, err := buildHelpOutputTestCommand().ToFishCompletion()
	require.NoError(t, err)
	got.WriteString("=== fish completion\n" + fish)

	if os.Getenv("UPDATE_GOLDEN") != "" {
		require.NoError(t, os.WriteFile(helpOutputFile, []byte(got.String()), 0o644))
	}
	expectFileContent(t, helpOutputFile, got.String())
}
