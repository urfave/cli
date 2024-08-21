package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	itesting "github.com/urfave/cli/v3/internal/testing"
)

var (
	lastExitCode = 0
	fakeOsExiter = func(rc int) {
		lastExitCode = rc
	}
	fakeErrWriter = &bytes.Buffer{}
)

func init() {
	OsExiter = fakeOsExiter
	ErrWriter = fakeErrWriter
}

type opCounts struct {
	Total, ShellComplete, OnUsageError, Before, CommandNotFound, Action, After, SubCommand int
}

func buildExtendedTestCommand() *Command {
	cmd := buildMinimalTestCommand()
	cmd.Name = "greet"
	cmd.Flags = []Flag{
		&StringFlag{
			Name:      "socket",
			Aliases:   []string{"s"},
			Usage:     "some 'usage' text",
			Value:     "value",
			TakesFile: true,
		},
		&StringFlag{Name: "flag", Aliases: []string{"fl", "f"}},
		&BoolFlag{
			Name:    "another-flag",
			Aliases: []string{"b"},
			Usage:   "another usage text",
			Sources: EnvVars("EXAMPLE_VARIABLE_NAME"),
		},
		&BoolFlag{
			Name:   "hidden-flag",
			Hidden: true,
		},
	}
	cmd.Commands = []*Command{{
		Aliases: []string{"c"},
		Flags: []Flag{
			&StringFlag{
				Name:      "flag",
				Aliases:   []string{"fl", "f"},
				TakesFile: true,
			},
			&BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
			},
		},
		Name:  "config",
		Usage: "another usage test",
		Commands: []*Command{{
			Aliases: []string{"s", "ss"},
			Flags: []Flag{
				&StringFlag{Name: "sub-flag", Aliases: []string{"sub-fl", "s"}},
				&BoolFlag{
					Name:    "sub-command-flag",
					Aliases: []string{"s"},
					Usage:   "some usage text",
				},
			},
			Name:  "sub-config",
			Usage: "another usage test",
		}},
	}, {
		Aliases: []string{"i", "in"},
		Name:    "info",
		Usage:   "retrieve generic information",
	}, {
		Name: "some-command",
	}, {
		Name:   "hidden-command",
		Hidden: true,
	}, {
		Aliases: []string{"u"},
		Flags: []Flag{
			&StringFlag{
				Name:      "flag",
				Aliases:   []string{"fl", "f"},
				TakesFile: true,
			},
			&BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
			},
		},
		Name:  "usage",
		Usage: "standard usage text",
		UsageText: `
Usage for the usage text
- formatted:  Based on the specified ConfigMap and summon secrets.yml
- list:       Inspect the environment for a specific process running on a Pod
- for_effect: Compare 'namespace' environment with 'local'

` + "```" + `
func() { ... }
` + "```" + `

Should be a part of the same code block
`,
		Commands: []*Command{{
			Aliases: []string{"su"},
			Flags: []Flag{
				&BoolFlag{
					Name:    "sub-command-flag",
					Aliases: []string{"s"},
					Usage:   "some usage text",
				},
			},
			Name:      "sub-usage",
			Usage:     "standard usage text",
			UsageText: "Single line of UsageText",
		}},
	}}
	cmd.UsageText = "app [first_arg] [second_arg]"
	cmd.Description = `Description of the application.`
	cmd.Usage = "Some app"
	cmd.Authors = []any{
		"Harrison <harrison@lolwut.example.com>",
		&mail.Address{Name: "Oliver Allen", Address: "oliver@toyshop.com"},
	}

	return cmd
}

func TestCommandFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            string
	}{
		// Test normal "not ignoring flags" flow
		{testArgs: []string{"test-cmd", "-break", "blah", "blah"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: "flag provided but not defined: -break"},
		{testArgs: []string{"test-cmd", "blah", "blah"}, skipFlagParsing: true, useShortOptionHandling: false},   // Test SkipFlagParsing without any args that look like flags
		{testArgs: []string{"test-cmd", "blah", "-break"}, skipFlagParsing: true, useShortOptionHandling: false}, // Test SkipFlagParsing with random flag arg
		{testArgs: []string{"test-cmd", "blah", "-help"}, skipFlagParsing: true, useShortOptionHandling: false},  // Test SkipFlagParsing with "special" help flag arg
		{testArgs: []string{"test-cmd", "blah", "-h"}, skipFlagParsing: false, useShortOptionHandling: true},     // Test UseShortOptionHandling
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs, " "), func(t *testing.T) {
			cmd := &Command{
				Writer:          io.Discard,
				Name:            "test-cmd",
				Aliases:         []string{"tc"},
				Usage:           "this is for testing",
				Description:     "testing",
				Action:          func(context.Context, *Command) error { return nil },
				SkipFlagParsing: c.skipFlagParsing,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			r := require.New(t)

			err := cmd.Run(ctx, c.testArgs)

			if c.expectedErr != "" {
				r.EqualError(err, c.expectedErr)
			} else {
				r.NoError(err)
			}
		})
	}
}

func TestParseAndRunShortOpts(t *testing.T) {
	testCases := []struct {
		testArgs     *stringSliceArgs
		expectedErr  string
		expectedArgs Args
	}{
		{testArgs: &stringSliceArgs{v: []string{"test", "-a"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-c", "arg1", "arg2"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1", "arg2"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-f"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-ac", "--fgh"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-af"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-cf"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "--acf"}}, expectedErr: "flag provided but not defined: -acf"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "-invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "--invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "--invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "arg1", "-invalid"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1", "-invalid"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "arg1", "--invalid"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1", "--invalid"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acfi", "not-arg", "arg1", "-invalid"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1", "-invalid"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-i", "ivalue"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-i", "ivalue", "arg1"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-i"}}, expectedErr: "flag needs an argument: -i"},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.testArgs.v, " "), func(t *testing.T) {
			state := map[string]Args{"args": nil}

			cmd := &Command{
				Name:        "test",
				Usage:       "this is for testing",
				Description: "testing",
				Action: func(_ context.Context, cmd *Command) error {
					state["args"] = cmd.Args()
					return nil
				},
				UseShortOptionHandling: true,
				Writer:                 io.Discard,
				Flags: []Flag{
					&BoolFlag{Name: "abc", Aliases: []string{"a"}},
					&BoolFlag{Name: "cde", Aliases: []string{"c"}},
					&BoolFlag{Name: "fgh", Aliases: []string{"f"}},
					&StringFlag{Name: "ijk", Aliases: []string{"i"}},
				},
			}

			err := cmd.Run(buildTestContext(t), tc.testArgs.Slice())

			r := require.New(t)

			if tc.expectedErr == "" {
				r.NoError(err)
			} else {
				r.ErrorContains(err, tc.expectedErr)
			}

			if tc.expectedArgs == nil {
				if state["args"] != nil {
					r.Len(state["args"].Slice(), 0)
				} else {
					r.Nil(state["args"])
				}
			} else {
				r.Equal(tc.expectedArgs, state["args"])
			}
		})
	}
}

func TestCommand_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Before: func(context.Context, *Command) error {
			return fmt.Errorf("before error")
		},
		After: func(context.Context, *Command) error {
			return fmt.Errorf("after error")
		},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"bar"})

	itesting.RequireErrorContains(t, err, "before error")
	itesting.RequireErrorContains(t, err, "after error")
}

func TestCommand_Run_BeforeSavesMetadata(t *testing.T) {
	var receivedMsgFromAction string
	var receivedMsgFromAfter string

	cmd := &Command{
		Name: "bar",
		Before: func(_ context.Context, cmd *Command) error {
			cmd.Metadata["msg"] = "hello world"
			return nil
		},
		Action: func(_ context.Context, cmd *Command) error {
			msg, ok := cmd.Metadata["msg"]
			if !ok {
				return errors.New("msg not found")
			}
			receivedMsgFromAction = msg.(string)
			return nil
		},
		After: func(_ context.Context, cmd *Command) error {
			msg, ok := cmd.Metadata["msg"]
			if !ok {
				return errors.New("msg not found")
			}
			receivedMsgFromAfter = msg.(string)
			return nil
		},
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar"}))
	itesting.RequireEqual(t, "hello world", receivedMsgFromAction)
	itesting.RequireEqual(t, "hello world", receivedMsgFromAfter)
}

func TestCommand_OnUsageError_hasCommandContext(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, cmd *Command, err error, _ bool) error {
			return fmt.Errorf("intercepted in %s: %s", cmd.Name, err.Error())
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"})
	itesting.ErrorContains(t, err, "intercepted in bar")
}

func TestCommand_OnUsageError_WithWrongFlagValue(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, _ *Command, err error, _ bool) error {
			itesting.ErrorContains(t, err, "invalid value \"wrong\"")
			return errors.New("intercepted: " + err.Error())
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"})
	itesting.ErrorContains(t, err, "intercepted: invalid value")
}

func TestCommand_OnUsageError_WithSubcommand(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Commands: []*Command{
			{
				Name: "baz",
			},
		},
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, _ *Command, err error, _ bool) error {
			itesting.ErrorContains(t, err, "invalid value \"wrong\"")
			return errors.New("intercepted: " + err.Error())
		},
	}

	itesting.RequireErrorContains(t, cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"}), "intercepted: invalid value")
}

func TestCommand_Run_SubcommandsCanUseErrWriter(t *testing.T) {
	cmd := &Command{
		ErrWriter: io.Discard,
		Name:      "bar",
		Usage:     "this is for testing",
		Commands: []*Command{
			{
				Name:  "baz",
				Usage: "this is for testing",
				Action: func(_ context.Context, cmd *Command) error {
					itesting.RequireEqual(t, io.Discard, cmd.Root().ErrWriter)

					return nil
				},
			},
		},
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"bar", "baz"}))
}

func TestCommandSkipFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs     *stringSliceArgs
		expectedArgs *stringSliceArgs
		expectedErr  error
	}{
		{testArgs: &stringSliceArgs{v: []string{"some-command", "some-arg", "--flag", "foo"}}, expectedArgs: &stringSliceArgs{v: []string{"some-arg", "--flag", "foo"}}, expectedErr: nil},
		{testArgs: &stringSliceArgs{v: []string{"some-command", "some-arg", "--flag=foo"}}, expectedArgs: &stringSliceArgs{v: []string{"some-arg", "--flag=foo"}}, expectedErr: nil},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs.Slice(), " "), func(t *testing.T) {
			var args Args
			cmd := &Command{
				SkipFlagParsing: true,
				Name:            "some-command",
				Flags: []Flag{
					&StringFlag{Name: "flag"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
				Writer: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), c.testArgs.Slice())
			itesting.Equal(t, c.expectedErr, err)
			itesting.Equal(t, c.expectedArgs, args)
		})
	}
}

func TestCommand_Run_CustomShellCompleteAcceptsMalformedFlags(t *testing.T) {
	cases := []struct {
		testArgs    *stringSliceArgs
		expectedOut string
	}{
		{testArgs: &stringSliceArgs{v: []string{"--undefined"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number", "forty-two"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number", "42"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number", "42", "newArg"}}, expectedOut: "found 1 args"},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs.Slice(), " "), func(t *testing.T) {
			out := &bytes.Buffer{}
			cmd := &Command{
				Writer:                out,
				EnableShellCompletion: true,
				Name:                  "bar",
				Usage:                 "this is for testing",
				Flags: []Flag{
					&IntFlag{
						Name:  "number",
						Usage: "A number to parse",
					},
				},
				ShellComplete: func(_ context.Context, cmd *Command) {
					fmt.Fprintf(cmd.Root().Writer, "found %[1]d args", cmd.NArg())
				},
			}

			osArgs := &stringSliceArgs{v: []string{"bar"}}
			osArgs.v = append(osArgs.v, c.testArgs.Slice()...)
			osArgs.v = append(osArgs.v, "--generate-shell-completion")

			r := require.New(t)

			r.NoError(cmd.Run(buildTestContext(t), osArgs.Slice()))
			r.Equal(c.expectedOut, out.String())
		})
	}
}

func TestCommand_CanAddVFlagOnSubCommands(t *testing.T) {
	cmd := &Command{
		Version: "some version",
		Writer:  io.Discard,
		Name:    "foo",
		Usage:   "this is for testing",
		Commands: []*Command{
			{
				Name: "bar",
				Flags: []Flag{
					&BoolFlag{Name: "v"},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "bar"})
	itesting.NoError(t, err)
}

func TestCommand_VisibleSubcCommands(t *testing.T) {
	subc1 := &Command{
		Name:  "subc1",
		Usage: "subc1 command1",
	}
	subc3 := &Command{
		Name:  "subc3",
		Usage: "subc3 command2",
	}
	cmd := &Command{
		Name:  "bar",
		Usage: "this is for testing",
		Commands: []*Command{
			subc1,
			{
				Name:   "subc2",
				Usage:  "subc2 command2",
				Hidden: true,
			},
			subc3,
		},
	}

	itesting.Equal(t, cmd.VisibleCommands(), []*Command{subc1, subc3})
}

func TestCommand_VisibleFlagCategories(t *testing.T) {
	cmd := &Command{
		Name:  "bar",
		Usage: "this is for testing",
		Flags: []Flag{
			&StringFlag{
				Name: "strd", // no category set
			},
			&StringFlag{
				Name:   "strd1", // no category set and also hidden
				Hidden: true,
			},
			&IntFlag{
				Name:     "intd",
				Aliases:  []string{"altd1", "altd2"},
				Category: "cat1",
			},
			&StringFlag{
				Name:     "sfd",
				Category: "cat2", // category set and hidden
				Hidden:   true,
			},
		},
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{{
			Category: "cat2",
			Flags: [][]Flag{
				{
					&StringFlag{
						Name: "mutex",
					},
				},
			},
		}},
	}

	cmd.MutuallyExclusiveFlags[0].propagateCategory()

	vfc := cmd.VisibleFlagCategories()
	itesting.RequireLen(t, vfc, 3)

	itesting.Equal(t, vfc[0].Name(), "", "expected category name to be empty")
	itesting.Equal(t, vfc[0].Flags()[0].Names(), []string{"strd"})

	itesting.Equal(t, vfc[1].Name(), "cat1", "expected category name cat1")
	itesting.RequireLen(t, vfc[1].Flags(), 1, "expected flag category to have one flag")
	itesting.Equal(t, vfc[1].Flags()[0].Names(), []string{"intd", "altd1", "altd2"})

	itesting.Equal(t, vfc[2].Name(), "cat2", "expected category name cat2")
	itesting.RequireLen(t, vfc[2].Flags(), 1, "expected flag category to have one flag")
	itesting.Equal(t, vfc[2].Flags()[0].Names(), []string{"mutex"})
}

func TestCommand_RunSubcommandWithDefault(t *testing.T) {
	cmd := &Command{
		Version:        "some version",
		Name:           "app",
		DefaultCommand: "foo",
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(context.Context, *Command) error {
					return errors.New("should not run this subcommand")
				},
			},
			{
				Name:     "bar",
				Usage:    "this is for testing",
				Commands: []*Command{{}}, // some subcommand
				Action: func(context.Context, *Command) error {
					return nil
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"app", "bar"})
	itesting.NoError(t, err)

	err = cmd.Run(buildTestContext(t), []string{"app"})
	itesting.EqualError(t, err, "should not run this subcommand")
}

func TestCommand_Run(t *testing.T) {
	s := ""

	cmd := &Command{
		Action: func(_ context.Context, cmd *Command) error {
			s = s + cmd.Args().First()
			return nil
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"command", "foo"})
	itesting.NoError(t, err)
	err = cmd.Run(buildTestContext(t), []string{"command", "bar"})
	itesting.NoError(t, err)
	itesting.Equal(t, s, "foobar")
}

var commandTests = []struct {
	name     string
	expected bool
}{
	{"foobar", true},
	{"batbaz", true},
	{"b", true},
	{"f", true},
	{"bat", false},
	{"nothing", false},
}

func TestCommand_Command(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{Name: "foobar", Aliases: []string{"f"}},
			{Name: "batbaz", Aliases: []string{"b"}},
		},
	}

	for _, test := range commandTests {
		if test.expected {
			itesting.NotEmpty(t, cmd.Command(test.name))
		} else {
			itesting.Empty(t, cmd.Command(test.name))
		}
	}
}

var defaultCommandTests = []struct {
	cmdName    string
	defaultCmd string
	expected   bool
}{
	{"foobar", "foobar", true},
	{"batbaz", "foobar", true},
	{"b", "", true},
	{"f", "", true},
	{"", "foobar", true},
	{"", "", true},
	{" ", "", false},
	{"bat", "batbaz", false},
	{"nothing", "batbaz", false},
	{"nothing", "", false},
}

func TestCommand_RunDefaultCommand(t *testing.T) {
	for _, test := range defaultCommandTests {
		testTitle := fmt.Sprintf("command=%[1]s-default=%[2]s", test.cmdName, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			cmd := &Command{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{Name: "foobar", Aliases: []string{"f"}},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			err := cmd.Run(buildTestContext(t), []string{"c", test.cmdName})
			if test.expected {
				itesting.NoError(t, err)
			} else {
				itesting.Error(t, err)
			}
		})
	}
}

var defaultCommandSubCommandTests = []struct {
	cmdName    string
	subCmd     string
	defaultCmd string
	expected   bool
}{
	{"foobar", "", "foobar", true},
	{"foobar", "carly", "foobar", true},
	{"batbaz", "", "foobar", true},
	{"b", "", "", true},
	{"f", "", "", true},
	{"", "", "foobar", true},
	{"", "", "", true},
	{"", "jimbob", "foobar", true},
	{"", "j", "foobar", true},
	{"", "carly", "foobar", true},
	{"", "jimmers", "foobar", true},
	{"", "jimmers", "", true},
	{" ", "jimmers", "foobar", false},
	{"", "", "", true},
	{" ", "", "", false},
	{" ", "j", "", false},
	{"bat", "", "batbaz", false},
	{"nothing", "", "batbaz", false},
	{"nothing", "", "", false},
	{"nothing", "j", "batbaz", false},
	{"nothing", "carly", "", false},
}

func TestCommand_RunDefaultCommandWithSubCommand(t *testing.T) {
	for _, test := range defaultCommandSubCommandTests {
		testTitle := fmt.Sprintf("command=%[1]s-subcmd=%[2]s-default=%[3]s", test.cmdName, test.subCmd, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			cmd := &Command{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{
						Name:    "foobar",
						Aliases: []string{"f"},
						Commands: []*Command{
							{Name: "jimbob", Aliases: []string{"j"}},
							{Name: "carly"},
						},
					},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			err := cmd.Run(buildTestContext(t), []string{"c", test.cmdName, test.subCmd})
			if test.expected {
				itesting.NoError(t, err)
			} else {
				itesting.Error(t, err)
			}
		})
	}
}

var defaultCommandFlagTests = []struct {
	cmdName    string
	flag       string
	defaultCmd string
	expected   bool
}{
	{"foobar", "", "foobar", true},
	{"foobar", "-c derp", "foobar", true},
	{"batbaz", "", "foobar", true},
	{"b", "", "", true},
	{"f", "", "", true},
	{"", "", "foobar", true},
	{"", "", "", true},
	{"", "-j", "foobar", true},
	{"", "-j", "foobar", true},
	{"", "-c derp", "foobar", true},
	{"", "--carly=derp", "foobar", true},
	{"", "-j", "foobar", true},
	{"", "-j", "", true},
	{" ", "-j", "foobar", false},
	{"", "", "", true},
	{" ", "", "", false},
	{" ", "-j", "", false},
	{"bat", "", "batbaz", false},
	{"nothing", "", "batbaz", false},
	{"nothing", "", "", false},
	{"nothing", "--jimbob", "batbaz", false},
	{"nothing", "--carly", "", false},
}

func TestCommand_RunDefaultCommandWithFlags(t *testing.T) {
	for _, test := range defaultCommandFlagTests {
		testTitle := fmt.Sprintf("command=%[1]s-flag=%[2]s-default=%[3]s", test.cmdName, test.flag, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			cmd := &Command{
				DefaultCommand: test.defaultCmd,
				Flags: []Flag{
					&StringFlag{
						Name:     "carly",
						Aliases:  []string{"c"},
						Required: false,
					},
					&BoolFlag{
						Name:     "jimbob",
						Aliases:  []string{"j"},
						Required: false,
						Value:    true,
					},
				},
				Commands: []*Command{
					{
						Name:    "foobar",
						Aliases: []string{"f"},
					},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			appArgs := []string{"c"}

			if test.flag != "" {
				flags := strings.Split(test.flag, " ")
				if len(flags) > 1 {
					appArgs = append(appArgs, flags...)
				}

				flags = strings.Split(test.flag, "=")
				if len(flags) > 1 {
					appArgs = append(appArgs, flags...)
				}
			}

			appArgs = append(appArgs, test.cmdName)

			err := cmd.Run(buildTestContext(t), appArgs)
			if test.expected {
				itesting.NoError(t, err)
			} else {
				itesting.Error(t, err)
			}
		})
	}
}

func TestCommand_FlagsFromExtPackage(t *testing.T) {
	var someint int
	flag.IntVar(&someint, "epflag", 2, "ext package flag usage")

	// Based on source code we can reset the global flag parsing this way
	defer func() {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	cmd := &Command{
		AllowExtFlags: true,
		Flags: []Flag{
			&StringFlag{
				Name:     "carly",
				Aliases:  []string{"c"},
				Required: false,
			},
			&BoolFlag{
				Name:     "jimbob",
				Aliases:  []string{"j"},
				Required: false,
				Value:    true,
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "-c", "cly", "--epflag", "10"})
	itesting.NoError(t, err)

	itesting.Equal(t, someint, int(10))

	cmd = &Command{
		Flags: []Flag{
			&StringFlag{
				Name:     "carly",
				Aliases:  []string{"c"},
				Required: false,
			},
			&BoolFlag{
				Name:     "jimbob",
				Aliases:  []string{"j"},
				Required: false,
				Value:    true,
			},
		},
	}

	// this should return an error since epflag shouldnt be registered
	err = cmd.Run(buildTestContext(t), []string{"foo", "-c", "cly", "--epflag", "10"})
	itesting.Error(t, err)
}

func TestCommand_Setup_defaultsReader(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})
	itesting.Equal(t, cmd.Reader, os.Stdin)
}

func TestCommand_Setup_defaultsWriter(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})
	itesting.Equal(t, cmd.Writer, os.Stdout)
}

func TestCommand_CommandWithFlagBeforeTerminator(t *testing.T) {
	var parsedOption string
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&StringFlag{Name: "option", Value: "", Usage: "some option"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parsedOption = cmd.String("option")
					args = cmd.Args()
					return nil
				},
			},
		},
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"}))

	itesting.RequireEqual(t, "my-option", parsedOption)
	itesting.RequireEqual(t, "my-arg", args.Get(0))
	itesting.RequireEqual(t, "--", args.Get(1))
	itesting.RequireEqual(t, "--notARealFlag", args.Get(2))
}

func TestCommand_CommandWithDash(t *testing.T) {
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
			},
		},
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "my-arg", "-"}))
	require.NotNil(t, args)
	itesting.RequireEqual(t, "my-arg", args.Get(0))
	itesting.RequireEqual(t, "-", args.Get(1))
}

func TestCommand_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
			},
		},
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "my-arg", "--", "notAFlagAtAll"}))

	require.NotNil(t, args)
	itesting.RequireEqual(t, "my-arg", args.Get(0))
	itesting.RequireEqual(t, "--", args.Get(1))
	itesting.RequireEqual(t, "notAFlagAtAll", args.Get(2))
}

func TestCommand_SkipFlagParsing(t *testing.T) {
	var args Args

	cmd := &Command{
		SkipFlagParsing: true,
		Action: func(_ context.Context, cmd *Command) error {
			args = cmd.Args()
			return nil
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "--", "my-arg", "notAFlagAtAll"})

	itesting.Equal(t, args.Get(0), "--")
	itesting.Equal(t, args.Get(1), "my-arg")
	itesting.Equal(t, args.Get(2), "notAFlagAtAll")
}

func TestCommand_VisibleCommands(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{
				Name:   "frob",
				Action: func(context.Context, *Command) error { return nil },
			},
			{
				Name:   "frib",
				Hidden: true,
				Action: func(context.Context, *Command) error { return nil },
			},
		},
	}

	cmd.setupDefaults([]string{"cli.test"})
	expected := []*Command{
		cmd.Commands[0],
		cmd.Commands[2], // help
	}
	actual := cmd.VisibleCommands()
	itesting.Len(t, actual, len(expected))
	for i, actualCommand := range actual {
		expectedCommand := expected[i]

		if expectedCommand.Action != nil {
			// comparing func addresses is OK!
			itesting.Equal(t, fmt.Sprintf("%p", expectedCommand.Action), fmt.Sprintf("%p", actualCommand.Action))
		}

		func() {
			// nil out funcs, as they cannot be compared
			// (https://github.com/golang/go/issues/8554)
			expectedAction := expectedCommand.Action
			actualAction := actualCommand.Action
			defer func() {
				expectedCommand.Action = expectedAction
				actualCommand.Action = actualAction
			}()
			expectedCommand.Action = nil
			actualCommand.Action = nil

			itesting.Equal(t, expectedCommand, actualCommand)
		}()
	}
}

func TestCommand_UseShortOptionHandling(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	cmd.Action = func(_ context.Context, cmd *Command) error {
		one = cmd.Bool("one")
		two = cmd.Bool("two")
		name = cmd.String("name")
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "-on", expected})
	itesting.True(t, one)
	itesting.False(t, two)
	itesting.Equal(t, name, expected)
}

func TestCommand_UseShortOptionHandling_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "-n"})
	itesting.EqualError(t, err, "flag needs an argument: -n")
}

func TestCommand_UseShortOptionHandlingCommand(t *testing.T) {
	var (
		one, two bool
		name     string
		expected = "expectedName"
	)

	cmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&BoolFlag{Name: "one", Aliases: []string{"o"}},
			&BoolFlag{Name: "two", Aliases: []string{"t"}},
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			one = cmd.Bool("one")
			two = cmd.Bool("two")
			name = cmd.String("name")
			return nil
		},
		UseShortOptionHandling: true,
		Writer:                 io.Discard,
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"cmd", "-on", expected}))
	require.True(t, one)
	require.False(t, two)
	itesting.RequireEqual(t, expected, name)
}

func TestCommand_UseShortOptionHandlingCommand_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	cmd.Commands = []*Command{command}

	itesting.RequireEqualError(
		t,
		cmd.Run(buildTestContext(t), []string{"", "cmd", "-n"}),
		"flag needs an argument: -n",
	)
}

func TestCommand_UseShortOptionHandlingSubCommand(t *testing.T) {
	var one, two bool
	var name string

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Commands = []*Command{
		{
			Name: "cmd",
			Commands: []*Command{
				{
					Name: "sub",
					Flags: []Flag{
						&BoolFlag{Name: "one", Aliases: []string{"o"}},
						&BoolFlag{Name: "two", Aliases: []string{"t"}},
						&StringFlag{Name: "name", Aliases: []string{"n"}},
					},
					Action: func(_ context.Context, cmd *Command) error {
						one = cmd.Bool("one")
						two = cmd.Bool("two")
						name = cmd.String("name")
						return nil
					},
				},
			},
		},
	}

	expected := "expectedName"

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "sub", "-on", expected}))
	require.True(t, one)
	require.False(t, two)
	itesting.RequireEqual(t, expected, name)
}

func TestCommand_UseShortOptionHandlingSubCommand_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
	}
	subCommand := &Command{
		Name: "sub",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	command.Commands = []*Command{subCommand}
	cmd.Commands = []*Command{command}

	err := cmd.Run(buildTestContext(t), []string{"", "cmd", "sub", "-n"})
	itesting.EqualError(t, err, "flag needs an argument: -n")
}

func TestCommand_UseShortOptionAfterSliceFlag(t *testing.T) {
	var one, two bool
	var name string
	var sliceValDest []string
	var sliceVal []string
	expected := "expectedName"

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringSliceFlag{Name: "env", Aliases: []string{"e"}, Destination: &sliceValDest},
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	cmd.Action = func(_ context.Context, cmd *Command) error {
		sliceVal = cmd.StringSlice("env")
		one = cmd.Bool("one")
		two = cmd.Bool("two")
		name = cmd.String("name")
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "-e", "foo", "-on", expected})
	itesting.Equal(t, sliceVal, []string{"foo"})
	itesting.Equal(t, sliceValDest, []string{"foo"})
	itesting.True(t, one)
	itesting.False(t, two)
	itesting.Equal(t, expected, name)
}

func TestCommand_Float64Flag(t *testing.T) {
	var meters float64

	cmd := &Command{
		Flags: []Flag{
			&FloatFlag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
		},
		Action: func(_ context.Context, cmd *Command) error {
			meters = cmd.Float("height")
			return nil
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "--height", "1.93"})
	itesting.Equal(t, 1.93, meters)
}

func TestCommand_ParseSliceFlags(t *testing.T) {
	var parsedIntSlice []int64
	var parsedStringSlice []string

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntSliceFlag{Name: "p", Value: []int64{}, Usage: "set one or more ip addr"},
					&StringSliceFlag{Name: "ip", Value: []string{}, Usage: "set one or more ports to open"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parsedIntSlice = cmd.IntSlice("p")
					parsedStringSlice = cmd.StringSlice("ip")
					return nil
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"", "cmd", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4"}))
	r.Equal([]int64{22, 80}, parsedIntSlice)
	r.Equal([]string{"8.8.8.8", "8.8.4.4"}, parsedStringSlice)
}

func TestCommand_ParseSliceFlagsWithMissingValue(t *testing.T) {
	var parsedIntSlice []int64
	var parsedStringSlice []string

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntSliceFlag{Name: "a", Usage: "set numbers"},
					&StringSliceFlag{Name: "str", Usage: "set strings"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parsedIntSlice = cmd.IntSlice("a")
					parsedStringSlice = cmd.StringSlice("str")
					return nil
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"", "cmd", "-a", "2", "-str", "A"}))
	r.Equal([]int64{2}, parsedIntSlice)
	r.Equal([]string{"A"}, parsedStringSlice)
}

func TestCommand_DefaultStdin(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})

	itesting.Equal(t, cmd.Reader, os.Stdin, "Default input reader not set.")
}

func TestCommand_DefaultStdout(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})

	itesting.Equal(t, cmd.Writer, os.Stdout, "Default output writer not set.")
}

func TestCommand_SetStdin(t *testing.T) {
	buf := make([]byte, 12)

	cmd := &Command{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Action: func(_ context.Context, cmd *Command) error {
			_, err := cmd.Reader.Read(buf)
			return err
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"help"})
	itesting.RequireNoError(t, err)
	itesting.Equal(t, "Hello World!", string(buf), "Command did not read input from desired reader.")
}

func TestCommand_SetStdin_Subcommand(t *testing.T) {
	buf := make([]byte, 12)

	cmd := &Command{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Commands: []*Command{
			{
				Name: "command",
				Commands: []*Command{
					{
						Name: "subcommand",
						Action: func(_ context.Context, cmd *Command) error {
							_, err := cmd.Root().Reader.Read(buf)
							return err
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"test", "command", "subcommand"})
	itesting.RequireNoError(t, err)
	itesting.Equal(t, "Hello World!", string(buf), "Command did not read input from desired reader.")
}

func TestCommand_SetStdout(t *testing.T) {
	var w bytes.Buffer

	cmd := &Command{
		Name:   "test",
		Writer: &w,
	}

	err := cmd.Run(buildTestContext(t), []string{"help"})
	itesting.RequireNoError(t, err)
	itesting.NotZero(t, w.Len(), "Command did not write output to desired writer.")
}

func TestCommand_BeforeFunc(t *testing.T) {
	counts := &opCounts{}
	beforeError := fmt.Errorf("fail")
	var err error

	cmd := &Command{
		Before: func(_ context.Context, cmd *Command) error {
			counts.Total++
			counts.Before = counts.Total
			s := cmd.String("opt")
			if s == "fail" {
				return beforeError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: io.Discard,
	}

	// run with the Before() func succeeding
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "succeed", "sub"})
	itesting.RequireNoError(t, err)

	itesting.Equal(t, 1, counts.Before, "Before() not executed when expected")
	itesting.Equal(t, 2, counts.SubCommand, "Subcommand not executed when expected")

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	itesting.ErrorIs(t, err, beforeError, "Run error expected, but not received")
	itesting.Equal(t, 1, counts.Before, "Before() not executed when expected")
	itesting.Equal(t, 0, counts.SubCommand, "Subcommand executed when NOT expected")

	// reset
	counts = &opCounts{}

	afterError := errors.New("fail again")
	cmd.After = func(context.Context, *Command) error {
		return afterError
	}

	// run with the Before() func failing, wrapped by After()
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if _, ok := err.(MultiError); !ok {
		t.Errorf("MultiError expected, but not received")
	}

	itesting.Equal(t, 1, counts.Before, "Before() not executed when expected")
	itesting.Zero(t, counts.SubCommand, "Subcommand executed when NOT expected")
}

func TestCommand_BeforeAfterFuncShellCompletion(t *testing.T) {
	t.Skip("TODO: is '--generate-shell-completion' (flag) still supported?")

	counts := &opCounts{}

	cmd := &Command{
		EnableShellCompletion: true,
		Before: func(context.Context, *Command) error {
			counts.Total++
			counts.Before = counts.Total
			return nil
		},
		After: func(context.Context, *Command) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: io.Discard,
	}

	r := require.New(t)

	// run with the Before() func succeeding
	r.NoError(
		cmd.Run(
			buildTestContext(t),
			[]string{
				"command",
				"--opt", "succeed",
				"sub", "--generate-shell-completion",
			},
		),
	)

	r.Equalf(0, counts.Before, "Before was run")
	r.Equal(0, counts.After, "After was run")
	r.Equal(0, counts.SubCommand, "SubCommand was run")
}

func TestCommand_AfterFunc(t *testing.T) {
	counts := &opCounts{}
	afterError := fmt.Errorf("fail")
	var err error

	cmd := &Command{
		After: func(_ context.Context, cmd *Command) error {
			counts.Total++
			counts.After = counts.Total
			s := cmd.String("opt")
			if s == "fail" {
				return afterError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
	}

	// run with the After() func succeeding
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "succeed", "sub"})
	itesting.RequireNoError(t, err)
	itesting.Equal(t, 2, counts.After, "After() not executed when expected")
	itesting.Equal(t, 1, counts.SubCommand, "Subcommand not executed when expected")

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	itesting.ErrorIs(t, err, afterError, "Run error expected, but not received")
	itesting.Equal(t, 2, counts.After, "After() not executed when expected")
	itesting.Equal(t, 1, counts.SubCommand, "Subcommand not executed when expected")

	/*
		reset
	*/
	counts = &opCounts{}
	// reset the flags since they are set previously
	cmd.Flags = []Flag{
		&StringFlag{Name: "opt"},
	}

	// run with none args
	err = cmd.Run(buildTestContext(t), []string{"command"})

	// should be the same error produced by the Before func
	itesting.RequireNoError(t, err)

	itesting.Equal(t, 1, counts.After, "After() not executed when expected")
	itesting.Equal(t, 0, counts.SubCommand, "Subcommand not executed when expected")
}

func TestCommandNoHelpFlag(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = nil

	cmd := &Command{Writer: io.Discard}

	err := cmd.Run(buildTestContext(t), []string{"test", "-h"})

	itesting.ErrorIs(t, err, flag.ErrHelp, "expected error about missing help flag")
}

func TestRequiredFlagCommandRunBehavior(t *testing.T) {
	tdata := []struct {
		testCase        string
		appFlags        []Flag
		appRunInput     []string
		appCommands     []*Command
		expectedAnError bool
	}{
		// assertion: empty input, when a required flag is present, errors
		{
			testCase:        "error_case_empty_input_with_required_flag_on_app",
			appRunInput:     []string{"myCLI"},
			appFlags:        []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_empty_input_with_required_flag_on_command",
			appRunInput: []string{"myCLI", "myCommand"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_empty_input_with_required_flag_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand"},
			appCommands: []*Command{{
				Name: "myCommand",
				Commands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
				}},
			}},
			expectedAnError: true,
		},
		// assertion: inputting --help, when a required flag is present, does not error
		{
			testCase:    "valid_case_help_input_with_required_flag_on_app",
			appRunInput: []string{"myCLI", "--help"},
			appFlags:    []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
		},
		{
			testCase:    "valid_case_help_input_with_required_flag_on_command",
			appRunInput: []string{"myCLI", "myCommand", "--help"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			}},
		},
		{
			testCase:    "valid_case_help_input_with_required_flag_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand", "--help"},
			appCommands: []*Command{{
				Name: "myCommand",
				Commands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
				}},
			}},
		},
		// assertion: giving optional input, when a required flag is present, errors
		{
			testCase:        "error_case_optional_input_with_required_flag_on_app",
			appRunInput:     []string{"myCLI", "--optional", "cats"},
			appFlags:        []Flag{&StringFlag{Name: "requiredFlag", Required: true}, &StringFlag{Name: "optional"}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_optional_input_with_required_flag_on_command",
			appRunInput: []string{"myCLI", "myCommand", "--optional", "cats"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}, &StringFlag{Name: "optional"}},
			}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_optional_input_with_required_flag_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand", "--optional", "cats"},
			appCommands: []*Command{{
				Name: "myCommand",
				Commands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}, &StringFlag{Name: "optional"}},
				}},
			}},
			expectedAnError: true,
		},
		// assertion: when a required flag is present, inputting that required flag does not error
		{
			testCase:    "valid_case_required_flag_input_on_app",
			appRunInput: []string{"myCLI", "--requiredFlag", "cats"},
			appFlags:    []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
		},
		{
			testCase:    "valid_case_required_flag_input_on_command",
			appRunInput: []string{"myCLI", "myCommand", "--requiredFlag", "cats"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			}},
		},
		{
			testCase:    "valid_case_required_flag_input_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand", "--requiredFlag", "cats"},
			appCommands: []*Command{{
				Name: "myCommand",
				Commands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
					Action: func(context.Context, *Command) error {
						return nil
					},
				}},
			}},
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			cmd := buildMinimalTestCommand()
			cmd.Flags = test.appFlags
			cmd.Commands = test.appCommands

			// logic under test
			err := cmd.Run(buildTestContext(t), test.appRunInput)

			// assertions
			if test.expectedAnError {
				itesting.Error(t, err)
				if _, ok := err.(requiredFlagsErr); test.expectedAnError && !ok {
					t.Errorf("expected a requiredFlagsErr, but got: %s", err)
				}
			} else {
				itesting.NoError(t, err)
			}
		})
	}
}

func TestCommandHelpPrinter(t *testing.T) {
	oldPrinter := HelpPrinter
	defer func() {
		HelpPrinter = oldPrinter
	}()

	wasCalled := false
	HelpPrinter = func(io.Writer, string, interface{}) {
		wasCalled = true
	}

	cmd := &Command{}

	_ = cmd.Run(buildTestContext(t), []string{"-h"})

	itesting.True(t, wasCalled, "Help printer expected to be called, but was not")
}

func TestCommand_VersionPrinter(t *testing.T) {
	oldPrinter := VersionPrinter
	defer func() {
		VersionPrinter = oldPrinter
	}()

	wasCalled := false
	VersionPrinter = func(*Command) {
		wasCalled = true
	}

	cmd := &Command{}
	ShowVersion(cmd)

	itesting.True(t, wasCalled, "Version printer expected to be called, but was not")
}

func TestCommand_CommandNotFound(t *testing.T) {
	counts := &opCounts{}
	cmd := &Command{
		CommandNotFound: func(context.Context, *Command, string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		},
		Commands: []*Command{
			{
				Name: "bar",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"command", "foo"})

	itesting.Equal(t, 1, counts.CommandNotFound, 1)
	itesting.Equal(t, 0, counts.SubCommand)
	itesting.Equal(t, 1, counts.Total)
}

func TestCommand_OrderOfOperations(t *testing.T) {
	buildCmdCounts := func() (*Command, *opCounts) {
		counts := &opCounts{}

		cmd := &Command{
			EnableShellCompletion: true,
			ShellComplete: func(context.Context, *Command) {
				counts.Total++
				counts.ShellComplete = counts.Total
			},
			OnUsageError: func(context.Context, *Command, error, bool) error {
				counts.Total++
				counts.OnUsageError = counts.Total
				return errors.New("hay OnUsageError")
			},
			Writer: io.Discard,
		}

		beforeNoError := func(context.Context, *Command) error {
			counts.Total++
			counts.Before = counts.Total
			return nil
		}

		cmd.Before = beforeNoError
		cmd.CommandNotFound = func(context.Context, *Command, string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		}

		afterNoError := func(context.Context, *Command) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		}

		cmd.After = afterNoError
		cmd.Commands = []*Command{
			{
				Name: "bar",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		}

		cmd.Action = func(context.Context, *Command) error {
			counts.Total++
			counts.Action = counts.Total
			return nil
		}

		return cmd, counts
	}

	t.Run("on usage error", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "--nope"})
		r.Equal(1, counts.OnUsageError)
		r.Equal(1, counts.Total)
	})

	t.Run("shell complete", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "--" + "generate-shell-completion"})
		r.Equal(1, counts.ShellComplete)
		r.Equal(1, counts.Total)
	})

	t.Run("nil on usage error", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.OnUsageError = nil

		_ = cmd.Run(buildTestContext(t), []string{"command", "--nope"})
		itesting.RequireEqual(t, 0, counts.Total)
	})

	t.Run("before after action hooks", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "foo"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(0, counts.CommandNotFound)
		r.Equal(2, counts.Action)
		r.Equal(3, counts.After)
		r.Equal(3, counts.Total)
	})

	t.Run("before with error", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.Before = func(context.Context, *Command) error {
			counts.Total++
			counts.Before = counts.Total
			return errors.New("hay Before")
		}

		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "bar"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.After)
		r.Equal(2, counts.Total)
	})

	t.Run("nil after", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.After = nil
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "bar"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.SubCommand)
		r.Equal(2, counts.Total)
	})

	t.Run("after errors", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.After = func(context.Context, *Command) error {
			counts.Total++
			counts.After = counts.Total
			return errors.New("hay After")
		}

		r := require.New(t)

		err := cmd.Run(buildTestContext(t), []string{"command", "bar"})
		r.Error(err)
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.SubCommand)
		r.Equal(3, counts.After)
		r.Equal(3, counts.Total)
	})

	t.Run("nil commands", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.Commands = nil
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.Action)
		r.Equal(3, counts.After)
		r.Equal(3, counts.Total)
	})
}

func TestCommand_Run_CommandWithSubcommandHasHelpTopic(t *testing.T) {
	subcommandHelpTopics := [][]string{
		{"foo", "--help"},
		{"foo", "-h"},
		{"foo", "help"},
	}

	for _, flagSet := range subcommandHelpTopics {
		t.Run(fmt.Sprintf("checking with flags %v", flagSet), func(t *testing.T) {
			buf := new(bytes.Buffer)

			subCmdBar := &Command{
				Name:  "bar",
				Usage: "does bar things",
			}
			subCmdBaz := &Command{
				Name:  "baz",
				Usage: "does baz things",
			}
			cmd := &Command{
				Name:        "foo",
				Description: "descriptive wall of text about how it does foo things",
				Commands:    []*Command{subCmdBar, subCmdBaz},
				Action:      func(context.Context, *Command) error { return nil },
				Writer:      buf,
			}

			err := cmd.Run(buildTestContext(t), flagSet)
			itesting.NoError(t, err)

			output := buf.String()

			itesting.NotContains(t, output, "No help topic for", "expect a help topic, got none")

			for _, shouldContain := range []string{
				cmd.Name, cmd.Description,
				subCmdBar.Name, subCmdBar.Usage,
				subCmdBaz.Name, subCmdBaz.Usage,
			} {
				itesting.Contains(t, output, shouldContain, "want help to contain %q, did not: \n%q", shouldContain, output)
			}
		})
	}
}

func TestCommand_Run_SubcommandFullPath(t *testing.T) {
	out := &bytes.Buffer{}

	subCmd := &Command{
		Name:      "bar",
		Usage:     "does bar things",
		ArgsUsage: "[arguments...]",
	}

	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Commands:    []*Command{subCmd},
		Writer:      out,
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar", "--help"}))

	outString := out.String()
	require.Contains(t, outString, "foo bar - does bar things")
	require.Contains(t, outString, "foo bar [command [command options]] [arguments...]")
}

func TestCommand_Run_Help(t *testing.T) {
	tests := []struct {
		helpArguments []string
		hideHelp      bool
		wantContains  string
		wantErr       error
	}{
		{
			helpArguments: []string{"boom", "--help"},
			hideHelp:      false,
			wantContains:  "boom - make an explosive entrance",
		},
		{
			helpArguments: []string{"boom", "-h"},
			hideHelp:      false,
			wantContains:  "boom - make an explosive entrance",
		},
		{
			helpArguments: []string{"boom", "help"},
			hideHelp:      false,
			wantContains:  "boom - make an explosive entrance",
		},
		{
			helpArguments: []string{"boom", "--help"},
			hideHelp:      true,
			wantErr:       fmt.Errorf("flag: help requested"),
		},
		{
			helpArguments: []string{"boom", "-h"},
			hideHelp:      true,
			wantErr:       fmt.Errorf("flag: help requested"),
		},
		{
			helpArguments: []string{"boom", "help"},
			hideHelp:      true,
			wantContains:  "boom I say!",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("checking with arguments %v", tt.helpArguments), func(t *testing.T) {
			buf := new(bytes.Buffer)

			cmd := &Command{
				Name:     "boom",
				Usage:    "make an explosive entrance",
				Writer:   buf,
				HideHelp: tt.hideHelp,
				Action: func(context.Context, *Command) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), tt.helpArguments)
			if tt.wantErr != nil {
				itesting.ErrorContains(t, err, tt.wantErr.Error())
			}

			output := buf.String()

			itesting.Contains(t, output, tt.wantContains, "want help to contain %q, did not: \n%q", "boom - make an explosive entrance", output)
		})
	}
}

func TestCommand_Run_Version(t *testing.T) {
	versionArguments := [][]string{{"boom", "--version"}, {"boom", "-v"}}

	for _, args := range versionArguments {
		t.Run(fmt.Sprintf("checking with arguments %v", args), func(t *testing.T) {
			buf := new(bytes.Buffer)

			cmd := &Command{
				Name:    "boom",
				Usage:   "make an explosive entrance",
				Version: "0.1.0",
				Writer:  buf,
				Action: func(context.Context, *Command) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), args)
			itesting.NoError(t, err)
			itesting.Contains(t, buf.String(), "0.1.0", "want version to contain 0.1.0")
		})
	}
}

func TestCommand_Run_Categories(t *testing.T) {
	buf := new(bytes.Buffer)

	cmd := &Command{
		Name:     "categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
			},
			{
				Name:     "command2",
				Category: "1",
			},
			{
				Name:     "command3",
				Category: "2",
			},
		},
		Writer: buf,
	}

	_ = cmd.Run(buildTestContext(t), []string{"categories"})

	expect := commandCategories([]*commandCategory{
		{
			name: "1",
			commands: []*Command{
				cmd.Commands[0],
				cmd.Commands[1],
			},
		},
		{
			name: "2",
			commands: []*Command{
				cmd.Commands[2],
			},
		},
	})

	itesting.RequireEqual(t, &expect, cmd.categories)

	output := buf.String()

	itesting.Contains(t, output, "1:\n     command1", "want buffer to include category %q, did not: \n%q", "1:\n     command1", output)
}

func TestCommand_VisibleCategories(t *testing.T) {
	cmd := &Command{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
			},
			{
				Name:     "command3",
				Category: "3",
			},
		},
	}

	expected := []CommandCategory{
		&commandCategory{
			name: "2",
			commands: []*Command{
				cmd.Commands[1],
			},
		},
		&commandCategory{
			name: "3",
			commands: []*Command{
				cmd.Commands[2],
			},
		},
	}

	cmd.setupDefaults([]string{"cli.test"})
	itesting.Equal(t, expected, cmd.VisibleCategories())

	cmd = &Command{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				Hidden:   true,
			},
			{
				Name:     "command3",
				Category: "3",
			},
		},
	}

	expected = []CommandCategory{
		&commandCategory{
			name: "3",
			commands: []*Command{
				cmd.Commands[2],
			},
		},
	}

	cmd.setupDefaults([]string{"cli.test"})
	itesting.Equal(t, expected, cmd.VisibleCategories())

	cmd = &Command{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				Hidden:   true,
			},
			{
				Name:     "command3",
				Category: "3",
				Hidden:   true,
			},
		},
	}

	cmd.setupDefaults([]string{"cli.test"})
	itesting.Empty(t, cmd.VisibleCategories())
}

func TestCommand_Run_SubcommandDoesNotOverwriteErrorFromBefore(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{
				Commands: []*Command{
					{
						Name: "sub",
					},
				},
				Name:   "bar",
				Before: func(context.Context, *Command) error { return fmt.Errorf("before error") },
				After:  func(context.Context, *Command) error { return fmt.Errorf("after error") },
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "bar"})
	itesting.ErrorContains(t, err, "before error")
	itesting.ErrorContains(t, err, "after error")
}

func TestCommand_OnUsageError_WithWrongFlagValue_ForSubcommand(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, _ *Command, err error, isSubcommand bool) error {
			itesting.False(t, isSubcommand, "Expect subcommand")
			itesting.ErrorContains(t, err, "invalid value \"wrong\"")
			return errors.New("intercepted: " + err.Error())
		},
		Commands: []*Command{
			{
				Name: "bar",
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--flag=wrong", "bar"})
	itesting.ErrorContains(t, err, "intercepted: invalid value", "Expect an intercepted error")
}

// A custom flag that conforms to the relevant interfaces, but has none of the
// fields that the other flag types do.
type customBoolFlag struct {
	Nombre string
}

// Don't use the normal FlagStringer
func (c *customBoolFlag) String() string {
	return "***" + c.Nombre + "***"
}

func (c *customBoolFlag) Names() []string {
	return []string{c.Nombre}
}

func (c *customBoolFlag) TakesValue() bool {
	return false
}

func (c *customBoolFlag) GetValue() string {
	return "value"
}

func (c *customBoolFlag) GetUsage() string {
	return "usage"
}

func (c *customBoolFlag) Apply(set *flag.FlagSet) error {
	set.String(c.Nombre, c.Nombre, "")
	return nil
}

func (c *customBoolFlag) RunAction(context.Context, *Command) error {
	return nil
}

func (c *customBoolFlag) IsSet() bool {
	return false
}

func (c *customBoolFlag) IsRequired() bool {
	return false
}

func (c *customBoolFlag) IsVisible() bool {
	return false
}

func (c *customBoolFlag) GetCategory() string {
	return ""
}

func (c *customBoolFlag) GetEnvVars() []string {
	return nil
}

func (c *customBoolFlag) GetDefaultText() string {
	return ""
}

func TestCustomFlagsUnused(t *testing.T) {
	cmd := &Command{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo"})
	itesting.NoError(t, err, "Run returned unexpected error")
}

func TestCustomFlagsUsed(t *testing.T) {
	cmd := &Command{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--custom=bar"})
	itesting.NoError(t, err, "Run returned unexpected error")
}

func TestCustomHelpVersionFlags(t *testing.T) {
	cmd := &Command{
		Writer: io.Discard,
	}

	// Be sure to reset the global flags
	defer func(helpFlag Flag, versionFlag Flag) {
		HelpFlag = helpFlag.(*BoolFlag)
		VersionFlag = versionFlag.(*BoolFlag)
	}(HelpFlag, VersionFlag)

	HelpFlag = &customBoolFlag{"help-custom"}
	VersionFlag = &customBoolFlag{"version-custom"}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--help-custom=bar"})
	itesting.NoError(t, err, "Run returned unexpected error")
}

func TestHandleExitCoder_Default(t *testing.T) {
	app := buildMinimalTestCommand()
	fs, err := newFlagSet(app.Name, app.Flags)
	itesting.NoError(t, err, "error creating FlagSet")

	app.flagSet = fs

	_ = app.handleExitCoder(context.Background(), Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	itesting.Contains(t, output, "Default", "Expected Default Behavior from Error Handler")
}

func TestHandleExitCoder_Custom(t *testing.T) {
	cmd := buildMinimalTestCommand()

	cmd.ExitErrHandler = func(context.Context, *Command, error) {
		_, _ = fmt.Fprintln(ErrWriter, "I'm a Custom error handler, I print what I want!")
	}

	_ = cmd.handleExitCoder(context.Background(), Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	itesting.Contains(t, output, "Custom", "Expected Custom Behavior from Error Handler")
}

func TestShellCompletionForIncompleteFlags(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{
				Name: "test-completion",
			},
		},
		EnableShellCompletion: true,
		ShellComplete: func(_ context.Context, cmd *Command) {
			for _, command := range cmd.Commands {
				if command.Hidden {
					continue
				}

				for _, name := range command.Names() {
					_, _ = fmt.Fprintln(cmd.Writer, name)
				}
			}

			for _, fl := range cmd.Flags {
				for _, name := range fl.Names() {
					if name == GenerateShellCompletionFlag.Names()[0] {
						continue
					}

					switch name = strings.TrimSpace(name); len(name) {
					case 0:
					case 1:
						_, _ = fmt.Fprintln(cmd.Writer, "-"+name)
					default:
						_, _ = fmt.Fprintln(cmd.Writer, "--"+name)
					}
				}
			}
		},
		Action: func(context.Context, *Command) error {
			return fmt.Errorf("should not get here")
		},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--test-completion", "--" + "generate-shell-completion"})
	itesting.NoError(t, err, "app should not return an error")
}

func TestWhenExitSubCommandWithCodeThenCommandQuitUnexpectedly(t *testing.T) {
	testCode := 104

	cmd := buildMinimalTestCommand()
	cmd.Commands = []*Command{
		{
			Name: "cmd",
			Commands: []*Command{
				{
					Name: "subcmd",
					Action: func(context.Context, *Command) error {
						return Exit("exit error", testCode)
					},
				},
			},
		},
	}

	// set user function as ExitErrHandler
	exitCodeFromExitErrHandler := int(0)
	cmd.ExitErrHandler = func(_ context.Context, _ *Command, err error) {
		if exitErr, ok := err.(ExitCoder); ok {
			exitCodeFromExitErrHandler = exitErr.ExitCode()
		}
	}

	// keep and restore original OsExiter
	origExiter := OsExiter
	t.Cleanup(func() { OsExiter = origExiter })

	// set user function as OsExiter
	exitCodeFromOsExiter := int(0)
	OsExiter = func(exitCode int) {
		exitCodeFromOsExiter = exitCode
	}

	r := require.New(t)

	r.Error(cmd.Run(buildTestContext(t), []string{
		"myapp",
		"cmd",
		"subcmd",
	}))

	r.Equal(0, exitCodeFromOsExiter)
	r.Equal(testCode, exitCodeFromExitErrHandler)
}

func buildMinimalTestCommand() *Command {
	// reset the help flag because tests may have set it
	HelpFlag.(*BoolFlag).hasBeenSet = false
	return &Command{Writer: io.Discard}
}

func TestSetupInitializesBothWriters(t *testing.T) {
	cmd := &Command{}

	cmd.setupDefaults([]string{"cli.test"})

	itesting.Equal(t, cmd.ErrWriter, os.Stderr, "expected a.ErrWriter to be os.Stderr")
	itesting.Equal(t, cmd.Writer, os.Stdout, "expected a.Writer to be os.Stdout")
}

func TestSetupInitializesOnlyNilWriters(t *testing.T) {
	wr := &bytes.Buffer{}
	cmd := &Command{
		ErrWriter: wr,
	}

	cmd.setupDefaults([]string{"cli.test"})

	itesting.Equal(t, cmd.ErrWriter, wr, "expected a.ErrWriter to be a *bytes.Buffer instance")
	itesting.Equal(t, cmd.Writer, os.Stdout, "expected a.Writer to be os.Stdout")
}

func TestFlagAction(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Minute)
	testCases := []struct {
		name string
		args []string
		err  string
		exp  string
	}{
		{
			name: "flag_string",
			args: []string{"app", "--f_string=string"},
			exp:  "string ",
		},
		{
			name: "flag_string_error",
			args: []string{"app", "--f_string="},
			err:  "empty string",
		},
		{
			name: "flag_string_slice",
			args: []string{"app", "--f_string_slice=s1,s2,s3"},
			exp:  "[s1 s2 s3] ",
		},
		{
			name: "flag_string_slice_error",
			args: []string{"app", "--f_string_slice=err"},
			err:  "error string slice",
		},
		{
			name: "flag_bool",
			args: []string{"app", "--f_bool"},
			exp:  "true ",
		},
		{
			name: "flag_bool_error",
			args: []string{"app", "--f_bool=false"},
			err:  "value is false",
		},
		{
			name: "flag_duration",
			args: []string{"app", "--f_duration=1h30m20s"},
			exp:  "1h30m20s ",
		},
		{
			name: "flag_duration_error",
			args: []string{"app", "--f_duration=0"},
			err:  "empty duration",
		},
		{
			name: "flag_float64",
			args: []string{"app", "--f_float64=3.14159"},
			exp:  "3.14159 ",
		},
		{
			name: "flag_float64_error",
			args: []string{"app", "--f_float64=-1"},
			err:  "negative float64",
		},
		{
			name: "flag_float64_slice",
			args: []string{"app", "--f_float64_slice=1.1,2.2,3.3"},
			exp:  "[1.1 2.2 3.3] ",
		},
		{
			name: "flag_float64_slice_error",
			args: []string{"app", "--f_float64_slice=-1"},
			err:  "invalid float64 slice",
		},
		{
			name: "flag_int",
			args: []string{"app", "--f_int=1"},
			exp:  "1 ",
		},
		{
			name: "flag_int_error",
			args: []string{"app", "--f_int=-1"},
			err:  "negative int",
		},
		{
			name: "flag_int_slice",
			args: []string{"app", "--f_int_slice=1,2,3"},
			exp:  "[1 2 3] ",
		},
		{
			name: "flag_int_slice_error",
			args: []string{"app", "--f_int_slice=-1"},
			err:  "invalid int slice",
		},
		{
			name: "flag_timestamp",
			args: []string{"app", "--f_timestamp", now.Format(time.DateTime)},
			exp:  now.UTC().Format(time.RFC3339) + " ",
		},
		{
			name: "flag_timestamp_error",
			args: []string{"app", "--f_timestamp", "0001-01-01 00:00:00"},
			err:  "zero timestamp",
		},
		{
			name: "flag_uint",
			args: []string{"app", "--f_uint=1"},
			exp:  "1 ",
		},
		{
			name: "flag_uint_error",
			args: []string{"app", "--f_uint=0"},
			err:  "zero uint64",
		},
		{
			name: "flag_no_action",
			args: []string{"app", "--f_no_action="},
			exp:  "",
		},
		{
			name: "command_flag",
			args: []string{"app", "c1", "--f_string=c1"},
			exp:  "c1 ",
		},
		{
			name: "subCommand_flag",
			args: []string{"app", "c1", "sub1", "--f_string=sub1"},
			exp:  "sub1 ",
		},
		{
			name: "mixture",
			args: []string{"app", "--f_string=app", "--f_uint=1", "--f_int_slice=1,2,3", "--f_duration=1h30m20s", "c1", "--f_string=c1", "sub1", "--f_string=sub1"},
			exp:  "app 1h30m20s [1 2 3] 1 c1 sub1 ",
		},
		{
			name: "flag_string_map",
			args: []string{"app", "--f_string_map=s1=s2,s3="},
			exp:  "map[s1:s2 s3:]",
		},
		{
			name: "flag_string_map_error",
			args: []string{"app", "--f_string_map=err="},
			err:  "error string map",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			out := &bytes.Buffer{}

			stringFlag := &StringFlag{
				Name: "f_string",
				Action: func(_ context.Context, cmd *Command, v string) error {
					if v == "" {
						return fmt.Errorf("empty string")
					}
					_, err := cmd.Root().Writer.Write([]byte(v + " "))
					return err
				},
			}

			cmd := &Command{
				Writer: out,
				Name:   "app",
				Commands: []*Command{
					{
						Name:   "c1",
						Flags:  []Flag{stringFlag},
						Action: func(_ context.Context, cmd *Command) error { return nil },
						Commands: []*Command{
							{
								Name:   "sub1",
								Action: func(context.Context, *Command) error { return nil },
								Flags:  []Flag{stringFlag},
							},
						},
					},
				},
				Flags: []Flag{
					stringFlag,
					&StringFlag{
						Name: "f_no_action",
					},
					&StringSliceFlag{
						Name: "f_string_slice",
						Action: func(_ context.Context, cmd *Command, v []string) error {
							if v[0] == "err" {
								return fmt.Errorf("error string slice")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&BoolFlag{
						Name: "f_bool",
						Action: func(_ context.Context, cmd *Command, v bool) error {
							if !v {
								return fmt.Errorf("value is false")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%t ", v)))
							return err
						},
					},
					&DurationFlag{
						Name: "f_duration",
						Action: func(_ context.Context, cmd *Command, v time.Duration) error {
							if v == 0 {
								return fmt.Errorf("empty duration")
							}
							_, err := cmd.Root().Writer.Write([]byte(v.String() + " "))
							return err
						},
					},
					&FloatFlag{
						Name: "f_float64",
						Action: func(_ context.Context, cmd *Command, v float64) error {
							if v < 0 {
								return fmt.Errorf("negative float64")
							}
							_, err := cmd.Root().Writer.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64) + " "))
							return err
						},
					},
					&FloatSliceFlag{
						Name: "f_float64_slice",
						Action: func(_ context.Context, cmd *Command, v []float64) error {
							if len(v) > 0 && v[0] < 0 {
								return fmt.Errorf("invalid float64 slice")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&IntFlag{
						Name: "f_int",
						Action: func(_ context.Context, cmd *Command, v int64) error {
							if v < 0 {
								return fmt.Errorf("negative int")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&IntSliceFlag{
						Name: "f_int_slice",
						Action: func(_ context.Context, cmd *Command, v []int64) error {
							if len(v) > 0 && v[0] < 0 {
								return fmt.Errorf("invalid int slice")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&TimestampFlag{
						Name: "f_timestamp",
						Config: TimestampConfig{
							Timezone: time.UTC,
							Layouts:  []string{time.DateTime},
						},
						Action: func(_ context.Context, cmd *Command, v time.Time) error {
							if v.IsZero() {
								return fmt.Errorf("zero timestamp")
							}

							_, err := cmd.Root().Writer.Write([]byte(v.Format(time.RFC3339) + " "))
							return err
						},
					},
					&UintFlag{
						Name: "f_uint",
						Action: func(_ context.Context, cmd *Command, v uint64) error {
							if v == 0 {
								return fmt.Errorf("zero uint64")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&StringMapFlag{
						Name: "f_string_map",
						Action: func(_ context.Context, cmd *Command, v map[string]string) error {
							if _, ok := v["err"]; ok {
								return fmt.Errorf("error string map")
							}
							_, err := cmd.Root().Writer.Write([]byte(fmt.Sprintf("%v", v)))
							return err
						},
					},
				},
				Action: func(context.Context, *Command) error { return nil },
			}

			err := cmd.Run(buildTestContext(t), test.args)

			r := require.New(t)

			if test.err != "" {
				r.EqualError(err, test.err)
				return
			}

			r.NoError(err)
			r.Equal(test.exp, out.String())
		})
	}
}

func TestPersistentFlag(t *testing.T) {
	var topInt, topPersistentInt, subCommandInt, appOverrideInt int64
	var appFlag string
	var appRequiredFlag string
	var appOverrideCmdInt int64
	var appSliceFloat64 []float64
	var persistentCommandSliceInt []int64
	var persistentFlagActionCount int64

	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:        "persistentCommandFlag",
				Persistent:  true,
				Destination: &appFlag,
				Action: func(context.Context, *Command, string) error {
					persistentFlagActionCount++
					return nil
				},
			},
			&IntSliceFlag{
				Name:        "persistentCommandSliceFlag",
				Persistent:  true,
				Destination: &persistentCommandSliceInt,
			},
			&FloatSliceFlag{
				Name:       "persistentCommandFloatSliceFlag",
				Persistent: true,
				Value:      []float64{11.3, 12.5},
			},
			&IntFlag{
				Name:        "persistentCommandOverrideFlag",
				Persistent:  true,
				Destination: &appOverrideInt,
			},
			&StringFlag{
				Name:        "persistentRequiredCommandFlag",
				Persistent:  true,
				Required:    true,
				Destination: &appRequiredFlag,
			},
		},
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntFlag{
						Name:        "cmdFlag",
						Destination: &topInt,
					},
					&IntFlag{
						Name:        "cmdPersistentFlag",
						Persistent:  true,
						Destination: &topPersistentInt,
					},
					&IntFlag{
						Name:        "paof",
						Aliases:     []string{"persistentCommandOverrideFlag"},
						Destination: &appOverrideCmdInt,
					},
				},
				Commands: []*Command{
					{
						Name: "subcmd",
						Flags: []Flag{
							&IntFlag{
								Name:        "cmdFlag",
								Destination: &subCommandInt,
							},
						},
						Action: func(_ context.Context, cmd *Command) error {
							appSliceFloat64 = cmd.FloatSlice("persistentCommandFloatSliceFlag")
							return nil
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{
		"app",
		"--persistentCommandFlag", "hello",
		"--persistentCommandSliceFlag", "100",
		"--persistentCommandOverrideFlag", "102",
		"cmd",
		"--cmdFlag", "12",
		"--persistentCommandSliceFlag", "102",
		"--persistentCommandFloatSliceFlag", "102.455",
		"--paof", "105",
		"--persistentRequiredCommandFlag", "hellor",
		"subcmd",
		"--cmdPersistentFlag", "20",
		"--cmdFlag", "11",
		"--persistentCommandFlag", "bar",
		"--persistentCommandSliceFlag", "130",
		"--persistentCommandFloatSliceFlag", "3.1445",
	})

	itesting.RequireNoError(t, err)

	itesting.Equal(t, "bar", appFlag)
	itesting.Equal(t, "hellor", appRequiredFlag)
	itesting.Equal(t, int64(12), topInt)
	itesting.Equal(t, int64(20), topPersistentInt)

	// this should be changed from app since
	// cmd overrides it
	itesting.Equal(t, int64(102), appOverrideInt)
	itesting.Equal(t, int64(11), subCommandInt)
	itesting.Equal(t, int64(105), appOverrideCmdInt)
	itesting.Equal(t, []int64{100, 102, 130}, persistentCommandSliceInt)
	itesting.Equal(t, []float64{102.455, 3.1445}, appSliceFloat64)
	itesting.Equal(t, int64(2), persistentFlagActionCount, "Expected persistent flag action to be called 2 times")
}

func TestPersistentFlagIsSet(t *testing.T) {
	result := ""
	resultIsSet := false

	app := &Command{
		Name: "root",
		Flags: []Flag{
			&StringFlag{
				Name:       "result",
				Persistent: true,
			},
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(_ context.Context, cmd *Command) error {
					result = cmd.String("result")
					resultIsSet = cmd.IsSet("result")
					return nil
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"root", "--result", "before", "sub"})
	itesting.RequireNoError(t, err)
	itesting.RequireEqual(t, "before", result)
	require.True(t, resultIsSet)

	err = app.Run(context.Background(), []string{"root", "sub", "--result", "after"})
	itesting.RequireNoError(t, err)
	itesting.RequireEqual(t, "after", result)
	require.True(t, resultIsSet)
}

func TestRequiredPersistentFlag(t *testing.T) {
	app := &Command{
		Name: "root",
		Flags: []Flag{
			&StringFlag{
				Name:       "result",
				Persistent: true,
				Required:   true,
			},
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(ctx context.Context, c *Command) error {
					return nil
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"root", "sub"})
	itesting.RequireError(t, err)

	err = app.Run(context.Background(), []string{"root", "sub", "--result", "after"})
	itesting.RequireNoError(t, err)
}

func TestFlagDuplicates(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errExpected bool
	}{
		{
			name: "all args present once",
			args: []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10"},
		},
		{
			name: "duplicate non slice flag(duplicatable)",
			args: []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10", "--iflag", "20"},
		},
		{
			name:        "duplicate non slice flag(non duplicatable)",
			args:        []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10", "--sflag", "trip"},
			errExpected: true,
		},
		{
			name:        "duplicate slice flag(non duplicatable)",
			args:        []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--fsflag", "3.0", "--iflag", "10"},
			errExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &Command{
				Flags: []Flag{
					&StringFlag{
						Name:     "sflag",
						OnlyOnce: true,
					},
					&IntSliceFlag{
						Name: "isflag",
					},
					&FloatSliceFlag{
						Name:     "fsflag",
						OnlyOnce: true,
					},
					&IntFlag{
						Name: "iflag",
					},
				},
				Action: func(context.Context, *Command) error {
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), test.args)
			if test.errExpected {
				itesting.NotNil(t, err)
			} else {
				itesting.Nil(t, err)
			}
		})
	}
}

func TestShorthandCommand(t *testing.T) {
	af := func(p *int) ActionFunc {
		return func(context.Context, *Command) error {
			*p = *p + 1
			return nil
		}
	}

	var cmd1, cmd2 int

	cmd := &Command{
		PrefixMatchCommands: true,
		Commands: []*Command{
			{
				Name:    "cthdisd",
				Aliases: []string{"cth"},
				Action:  af(&cmd1),
			},
			{
				Name:    "cthertoop",
				Aliases: []string{"cer"},
				Action:  af(&cmd2),
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "cth"})
	itesting.NoError(t, err)
	itesting.True(t, cmd1 == 1 && cmd2 == 0, "Expected command1 to be trigerred once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthd"})
	itesting.NoError(t, err)
	itesting.True(t, cmd1 == 1 && cmd2 == 0, "Expected command1 to be trigerred once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthe"})
	itesting.NoError(t, err)
	itesting.True(t, cmd1 == 1 && cmd2 == 0, "Expected command1 to be trigerred once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthert"})
	itesting.NoError(t, err)
	itesting.True(t, cmd1 == 0 && cmd2 == 1, "Expected command1 to be trigerred once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthet"})
	itesting.NoError(t, err)
	itesting.True(t, cmd1 == 0 && cmd2 == 1, "Expected command1 to be trigerred once")
}

func TestCommand_Int(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int64("myflag", 12, "doc")

	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Int64("top-flag", 13, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}

	itesting.RequireEqual(t, int64(12), cmd.Int("myflag"))
	itesting.RequireEqual(t, int64(13), cmd.Int("top-flag"))
}

func TestCommand_Uint(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Uint64("myflagUint", uint64(13), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Uint64("top-flag", uint64(14), "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}

	itesting.RequireEqual(t, uint64(13), cmd.Uint("myflagUint"))
	itesting.RequireEqual(t, uint64(14), cmd.Uint("top-flag"))
}

func TestCommand_Float64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Float64("myflag", float64(17), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Float64("top-flag", float64(18), "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}

	r := require.New(t)
	r.Equal(float64(17), cmd.Float("myflag"))
	r.Equal(float64(18), cmd.Float("top-flag"))
}

func TestCommand_Duration(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Duration("myflag", 12*time.Second, "doc")

	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Duration("top-flag", 13*time.Second, "doc")
	pCmd := &Command{flagSet: parentSet}

	cmd := &Command{flagSet: set, parent: pCmd}

	r := require.New(t)
	r.Equal(12*time.Second, cmd.Duration("myflag"))
	r.Equal(13*time.Second, cmd.Duration("top-flag"))
}

func TestCommand_Timestamp(t *testing.T) {
	t1 := time.Time{}.Add(12 * time.Second)
	t2 := time.Time{}.Add(13 * time.Second)

	cmd := &Command{
		Name: "hello",
		Flags: []Flag{
			&TimestampFlag{
				Name:  "myflag",
				Value: t1,
			},
		},
		Action: func(ctx context.Context, c *Command) error {
			return nil
		},
	}

	pCmd := &Command{
		Flags: []Flag{
			&TimestampFlag{
				Name:  "top-flag",
				Value: t2,
			},
		},
		Commands: []*Command{
			cmd,
		},
	}

	err := pCmd.Run(context.Background(), []string{"foo", "hello"})
	itesting.NoError(t, err)

	r := require.New(t)
	r.Equal(t1, cmd.Timestamp("myflag"))
	r.Equal(t2, cmd.Timestamp("top-flag"))
}

func TestCommand_String(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("myflag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.String("top-flag", "hai veld", "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}

	r := require.New(t)
	r.Equal("hello world", cmd.String("myflag"))
	r.Equal("hai veld", cmd.String("top-flag"))

	cmd = &Command{parent: pCmd}
	r.Equal("hai veld", cmd.String("top-flag"))
}

func TestCommand_Bool(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}

	r := require.New(t)
	r.False(cmd.Bool("myflag"))
	r.True(cmd.Bool("top-flag"))
}

func TestCommand_Value(t *testing.T) {
	subCmd := &Command{
		Name: "test",
		Flags: []Flag{
			&IntFlag{
				Name:    "myflag",
				Usage:   "doc",
				Aliases: []string{"m", "mf"},
			},
		},
		Action: func(ctx context.Context, c *Command) error {
			return nil
		},
	}

	cmd := &Command{
		Flags: []Flag{
			&IntFlag{
				Name:    "top-flag",
				Usage:   "doc",
				Aliases: []string{"t", "tf"},
			},
		},
		Commands: []*Command{
			subCmd,
		},
	}
	t.Run("flag name", func(t *testing.T) {
		r := require.New(t)
		err := cmd.Run(buildTestContext(t), []string{"main", "--top-flag", "13", "test", "--myflag", "14"})

		r.NoError(err)
		r.Equal(int64(13), cmd.Value("top-flag"))
		r.Equal(int64(13), cmd.Value("t"))
		r.Equal(int64(13), cmd.Value("tf"))

		r.Equal(int64(14), subCmd.Value("myflag"))
		r.Equal(int64(14), subCmd.Value("m"))
		r.Equal(int64(14), subCmd.Value("mf"))
	})

	t.Run("flag aliases", func(t *testing.T) {
		r := require.New(t)
		err := cmd.Run(buildTestContext(t), []string{"main", "-tf", "15", "test", "-m", "16"})

		r.NoError(err)
		r.Equal(int64(15), cmd.Value("top-flag"))
		r.Equal(int64(15), cmd.Value("t"))
		r.Equal(int64(15), cmd.Value("tf"))

		r.Equal(int64(16), subCmd.Value("myflag"))
		r.Equal(int64(16), subCmd.Value("m"))
		r.Equal(int64(16), subCmd.Value("mf"))
		r.Nil(cmd.Value("unknown-flag"))
	})
}

func TestCommand_Value_InvalidFlagAccessHandler(t *testing.T) {
	var flagName string
	cmd := &Command{
		InvalidFlagAccessHandler: func(_ context.Context, _ *Command, name string) {
			flagName = name
		},
		Commands: []*Command{
			{
				Name: "command",
				Commands: []*Command{
					{
						Name: "subcommand",
						Action: func(_ context.Context, cmd *Command) error {
							cmd.Value("missing")
							return nil
						},
					},
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"run", "command", "subcommand"}))
	r.Equal("missing", flagName)
}

func TestCommand_Args(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	cmd := &Command{flagSet: set}
	_ = set.Parse([]string{"--myflag", "bat", "baz"})

	r := require.New(t)
	r.Equal(2, cmd.Args().Len())
	r.True(cmd.Bool("myflag"))
}

func TestCommand_NArg(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	cmd := &Command{flagSet: set}
	_ = set.Parse([]string{"--myflag", "bat", "baz"})

	itesting.RequireEqual(t, 2, cmd.NArg())
}

func TestCommand_IsSet(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.Bool("two-flag", false, "doc")
	set.String("three-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}

	_ = set.Parse([]string{"--one-flag", "--two-flag", "--three-flag", "frob"})
	_ = parentSet.Parse([]string{"--top-flag"})

	r := require.New(t)

	r.True(cmd.IsSet("one-flag"))
	r.True(cmd.IsSet("two-flag"))
	r.True(cmd.IsSet("three-flag"))
	r.True(cmd.IsSet("top-flag"))
	r.False(cmd.IsSet("bogus"))
}

// XXX Corresponds to hack in context.IsSet for flags with EnvVar field
// Should be moved to `flag_test` in v2
func TestCommand_IsSet_fromEnv(t *testing.T) {
	var (
		timeoutIsSet, tIsSet    bool
		noEnvVarIsSet, nIsSet   bool
		passwordIsSet, pIsSet   bool
		unparsableIsSet, uIsSet bool
	)

	t.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	t.Setenv("APP_PASSWORD", "")

	cmd := &Command{
		Flags: []Flag{
			&FloatFlag{Name: "timeout", Aliases: []string{"t"}, Sources: EnvVars("APP_TIMEOUT_SECONDS")},
			&StringFlag{Name: "password", Aliases: []string{"p"}, Sources: EnvVars("APP_PASSWORD")},
			&FloatFlag{Name: "unparsable", Aliases: []string{"u"}, Sources: EnvVars("APP_UNPARSABLE")},
			&FloatFlag{Name: "no-env-var", Aliases: []string{"n"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			timeoutIsSet = cmd.IsSet("timeout")
			tIsSet = cmd.IsSet("t")
			passwordIsSet = cmd.IsSet("password")
			pIsSet = cmd.IsSet("p")
			unparsableIsSet = cmd.IsSet("unparsable")
			uIsSet = cmd.IsSet("u")
			noEnvVarIsSet = cmd.IsSet("no-env-var")
			nIsSet = cmd.IsSet("n")
			return nil
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"run"}))
	r.True(timeoutIsSet)
	r.True(tIsSet)
	r.True(passwordIsSet)
	r.True(pIsSet)
	r.False(noEnvVarIsSet)
	r.False(nIsSet)

	t.Setenv("APP_UNPARSABLE", "foobar")

	r.Error(cmd.Run(buildTestContext(t), []string{"run"}))
	r.False(unparsableIsSet)
	r.False(uIsSet)
}

func TestCommand_NumFlags(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	set.String("otherflag", "hello world", "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("myflagGlobal", true, "doc")
	globalCmd := &Command{flagSet: globalSet}
	cmd := &Command{flagSet: set, parent: globalCmd}
	_ = set.Parse([]string{"--myflag", "--otherflag=foo"})
	_ = globalSet.Parse([]string{"--myflagGlobal"})
	itesting.RequireEqual(t, 2, cmd.NumFlags())
}

func TestCommand_Set(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int64("int", int64(5), "an int")
	cmd := &Command{flagSet: set}

	r := require.New(t)

	r.False(cmd.IsSet("int"))
	r.NoError(cmd.Set("int", "1"))
	r.Equal(int64(1), cmd.Int("int"))
	r.True(cmd.IsSet("int"))
}

func TestCommand_Set_InvalidFlagAccessHandler(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	var flagName string
	cmd := &Command{
		InvalidFlagAccessHandler: func(_ context.Context, _ *Command, name string) {
			flagName = name
		},
		flagSet: set,
	}

	r := require.New(t)

	r.True(cmd.Set("missing", "") != nil)
	r.Equal("missing", flagName)
}

func TestCommand_LocalFlagNames(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.String("two-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}
	_ = set.Parse([]string{"--one-flag", "--two-flag=foo"})
	_ = parentSet.Parse([]string{"--top-flag"})

	actualFlags := cmd.LocalFlagNames()
	sort.Strings(actualFlags)

	itesting.RequireEqual(t, []string{"one-flag", "two-flag"}, actualFlags)
}

func TestCommand_FlagNames(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.String("two-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}
	_ = set.Parse([]string{"--one-flag", "--two-flag=foo"})
	_ = parentSet.Parse([]string{"--top-flag"})

	actualFlags := cmd.FlagNames()
	sort.Strings(actualFlags)

	itesting.RequireEqual(t, []string{"one-flag", "top-flag", "two-flag"}, actualFlags)
}

func TestCommand_Lineage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("local-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}
	_ = set.Parse([]string{"--local-flag"})
	_ = parentSet.Parse([]string{"--top-flag"})

	lineage := cmd.Lineage()

	r := require.New(t)
	r.Equal(2, len(lineage))
	r.Equal(cmd, lineage[0])
	r.Equal(pCmd, lineage[1])
}

func TestCommand_lookupFlagSet(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("local-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	pCmd := &Command{flagSet: parentSet}
	cmd := &Command{flagSet: set, parent: pCmd}
	_ = set.Parse([]string{"--local-flag"})
	_ = parentSet.Parse([]string{"--top-flag"})

	r := require.New(t)

	fs := cmd.lookupFlagSet("top-flag")
	r.Equal(pCmd.flagSet, fs)

	fs = cmd.lookupFlagSet("local-flag")
	r.Equal(cmd.flagSet, fs)
	r.Nil(cmd.lookupFlagSet("frob"))
}

func TestCommandAttributeAccessing(t *testing.T) {
	tdata := []struct {
		testCase     string
		setBoolInput string
		ctxBoolInput string
		parent       *Command
	}{
		{
			testCase:     "empty",
			setBoolInput: "",
			ctxBoolInput: "",
		},
		{
			testCase:     "empty_with_background_context",
			setBoolInput: "",
			ctxBoolInput: "",
			parent:       &Command{},
		},
		{
			testCase:     "empty_set_bool_and_present_ctx_bool",
			setBoolInput: "",
			ctxBoolInput: "ctx-bool",
		},
		{
			testCase:     "present_set_bool_and_present_ctx_bool_with_background_context",
			setBoolInput: "",
			ctxBoolInput: "ctx-bool",
			parent:       &Command{},
		},
		{
			testCase:     "present_set_bool_and_present_ctx_bool",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "ctx-bool",
		},
		{
			testCase:     "present_set_bool_and_present_ctx_bool_with_background_context",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "ctx-bool",
			parent:       &Command{},
		},
		{
			testCase:     "present_set_bool_and_different_ctx_bool",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "not-ctx-bool",
		},
		{
			testCase:     "present_set_bool_and_different_ctx_bool_with_background_context",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "not-ctx-bool",
			parent:       &Command{},
		},
	}

	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			set := flag.NewFlagSet("some-flag-set-name", 0)
			set.Bool(test.setBoolInput, false, "usage documentation")
			cmd := &Command{flagSet: set, parent: test.parent}

			require.False(t, cmd.Bool(test.ctxBoolInput))
		})
	}
}

func TestCheckRequiredFlags(t *testing.T) {
	tdata := []struct {
		testCase              string
		parseInput            []string
		envVarInput           [2]string
		flags                 []Flag
		expectedAnError       bool
		expectedErrorContents []string
	}{
		{
			testCase: "empty",
		},
		{
			testCase: "optional",
			flags: []Flag{
				&StringFlag{Name: "optionalFlag"},
			},
		},
		{
			testCase: "required",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlag"},
		},
		{
			testCase: "required_and_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "required_and_present_via_env_var",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true, Sources: EnvVars("REQUIRED_FLAG")},
			},
			envVarInput: [2]string{"REQUIRED_FLAG", "true"},
		},
		{
			testCase: "required_and_optional",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			parseInput:      []string{"--optionalFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present_via_env_var",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag", Sources: EnvVars("OPTIONAL_FLAG")},
			},
			envVarInput:     [2]string{"OPTIONAL_FLAG", "true"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_required_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "two_required",
			flags: []Flag{
				&StringFlag{Name: "requiredFlagOne", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlagOne", "requiredFlagTwo"},
		},
		{
			testCase: "two_required_and_one_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput:      []string{"--requiredFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "two_required_and_both_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput", "--requiredFlagTwo", "myinput"},
		},
		{
			testCase: "required_flag_with_short_name",
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"N"}, Required: true},
			},
			parseInput: []string{"-N", "asd", "-N", "qwe"},
		},
		{
			testCase: "required_flag_with_multiple_short_names",
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"N", "n"}, Required: true},
			},
			parseInput: []string{"-n", "asd", "-n", "qwe"},
		},
		{
			testCase:              "required_flag_with_short_alias_not_printed_on_error",
			expectedAnError:       true,
			expectedErrorContents: []string{"Required flag \"names\" not set"},
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"n"}, Required: true},
			},
		},
		{
			testCase:              "required_flag_with_one_character",
			expectedAnError:       true,
			expectedErrorContents: []string{"Required flag \"n\" not set"},
			flags: []Flag{
				&StringFlag{Name: "n", Required: true},
			},
		},
	}

	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			if test.envVarInput[0] != "" {
				defer resetEnv(os.Environ())
				os.Clearenv()
				_ = os.Setenv(test.envVarInput[0], test.envVarInput[1])
			}

			set := flag.NewFlagSet("test", 0)
			for _, flags := range test.flags {
				_ = flags.Apply(set)
			}
			_ = set.Parse(test.parseInput)

			cmd := &Command{
				Flags:   test.flags,
				flagSet: set,
			}

			err := cmd.checkRequiredFlags()

			// assertions
			if test.expectedAnError {
				itesting.NotNil(t, err)
			} else {
				itesting.Nil(t, err)
			}
			for _, errString := range test.expectedErrorContents {
				if err != nil {
					itesting.ErrorContains(t, err, errString)
				}
			}
		})
	}
}

func TestCommand_ParentCommand_Set(t *testing.T) {
	parentSet := flag.NewFlagSet("parent", flag.ContinueOnError)
	parentSet.String("Name", "", "")

	cmd := &Command{
		flagSet: flag.NewFlagSet("child", flag.ContinueOnError),
		parent: &Command{
			flagSet: parentSet,
		},
	}

	err := cmd.Set("Name", "aaa")
	itesting.NoError(t, err)
}

func TestCommandReadArgsFromStdIn(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		args          []string
		expectedInt   int64
		expectedFloat float64
		expectedSlice []string
		expectError   bool
	}{
		{
			name:          "empty",
			input:         "",
			args:          []string{"foo"},
			expectedInt:   0,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name: "empty2",
			input: `
			
			`,
			args:          []string{"foo"},
			expectedInt:   0,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name:          "intflag-from-input",
			input:         "--if=100",
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name: "intflag-from-input2",
			input: `
			--if 

			100`,
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name: "multiflag-from-input",
			input: `
			--if

			100
			--ff      100.1

			--ssf hello
			--ssf

			"hello	
  123
44"
			`,
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 100.1,
			expectedSlice: []string{"hello", "hello\t\n  123\n44"},
		},
		{
			name: "end-args",
			input: `
			--if

			100
			--
			--ff      100.1

			--ssf hello
			--ssf

			hell02
			`,
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 0,
			expectedSlice: []string{},
		},
		{
			name: "invalid string",
			input: `
			"
			`,
			args:          []string{"foo"},
			expectedInt:   0,
			expectedFloat: 0,
			expectedSlice: []string{},
		},
		{
			name: "invalid string2",
			input: `
			--if
			"
			`,
			args:        []string{"foo"},
			expectError: true,
		},
		{
			name: "incomplete string",
			input: `
			--ssf
			"
			hello
			`,
			args:          []string{"foo"},
			expectedSlice: []string{"hello"},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			r := require.New(t)

			fp, err := os.CreateTemp("", "readargs")
			r.NoError(err)
			_, err = fp.Write([]byte(tst.input))
			r.NoError(err)
			fp.Close()

			cmd := buildMinimalTestCommand()
			cmd.ReadArgsFromStdin = true
			cmd.Reader, err = os.Open(fp.Name())
			r.NoError(err)
			cmd.Flags = []Flag{
				&IntFlag{
					Name: "if",
				},
				&FloatFlag{
					Name: "ff",
				},
				&StringSliceFlag{
					Name: "ssf",
				},
			}

			actionCalled := false
			cmd.Action = func(ctx context.Context, c *Command) error {
				r.Equal(tst.expectedInt, c.Int("if"))
				r.Equal(tst.expectedFloat, c.Float("ff"))
				r.Equal(tst.expectedSlice, c.StringSlice("ssf"))
				actionCalled = true
				return nil
			}

			err = cmd.Run(context.Background(), tst.args)
			if !tst.expectError {
				r.NoError(err)
				r.True(actionCalled)
			} else {
				r.Error(err)
			}
		})
	}
}

func TestJSONExportCommand(t *testing.T) {
	cmd := buildExtendedTestCommand()
	cmd.Arguments = []Argument{
		&IntArg{
			Name: "fooi",
		},
	}

	out, err := json.Marshal(cmd)
	itesting.RequireNoError(t, err)

	expected := `{
		"name": "greet",
		"aliases": null,
		"usage": "Some app",
		"usageText": "app [first_arg] [second_arg]",
		"argsUsage": "",
		"version": "",
		"description": "Description of the application.",
		"defaultCommand": "",
		"category": "",
		"commands": [
		  {
			"name": "config",
			"aliases": [
			  "c"
			],
			"usage": "another usage test",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": [
			  {
				"name": "sub-config",
				"aliases": [
				  "s",
				  "ss"
				],
				"usage": "another usage test",
				"usageText": "",
				"argsUsage": "",
				"version": "",
				"description": "",
				"defaultCommand": "",
				"category": "",
				"commands": null,
				"flags": [
				  {
					"name": "sub-flag",
					"category": "",
					"defaultText": "",
					"usage": "",
					"required": false,
					"hidden": false,
					"hideDefault": false,
					"persistent": false,
					"defaultValue": "",
					"aliases": [
					  "sub-fl",
					  "s"
					],
					"takesFileArg": false,
					"config": {
					  "TrimSpace": false
					},
					"onlyOnce": false
				  },
				  {
					"name": "sub-command-flag",
					"category": "",
					"defaultText": "",
					"usage": "some usage text",
					"required": false,
					"hidden": false,
					"hideDefault": false,
					"persistent": false,
					"defaultValue": false,
					"aliases": [
					  "s"
					],
					"takesFileArg": false,
					"config": {
					  "Count": null
					},
					"onlyOnce": false
				  }
				],
				"hideHelp": false,
				"hideHelpCommand": false,
				"hideVersion": false,
				"hidden": false,
				"authors": null,
				"copyright": "",
				"metadata": null,
				"sliceFlagSeparator": "",
				"disableSliceFlagSeparator": false,
				"useShortOptionHandling": false,
				"suggest": false,
				"allowExtFlags": false,
				"skipFlagParsing": false,
				"prefixMatchCommands": false,
				"mutuallyExclusiveFlags": null,
				"arguments": null,
				"readArgsFromStdin": false
			  }
			],
			"flags": [
			  {
				"name": "flag",
				"category": "",
				"defaultText": "",
				"usage": "",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"persistent": false,
				"defaultValue": "",
				"aliases": [
				  "fl",
				  "f"
				],
				"takesFileArg": true,
				"config": {
				  "TrimSpace": false
				},
				"onlyOnce": false
			  },
			  {
				"name": "another-flag",
				"category": "",
				"defaultText": "",
				"usage": "another usage text",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"persistent": false,
				"defaultValue": false,
				"aliases": [
				  "b"
				],
				"takesFileArg": false,
				"config": {
				  "Count": null
				},
				"onlyOnce": false
			  }
			],
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "info",
			"aliases": [
			  "i",
			  "in"
			],
			"usage": "retrieve generic information",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": null,
			"flags": null,
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "some-command",
			"aliases": null,
			"usage": "",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": null,
			"flags": null,
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "hidden-command",
			"aliases": null,
			"usage": "",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": null,
			"flags": null,
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": true,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "usage",
			"aliases": [
			  "u"
			],
			"usage": "standard usage text",
			"usageText": "\nUsage for the usage text\n- formatted:  Based on the specified ConfigMap and summon secrets.yml\n- list:       Inspect the environment for a specific process running on a Pod\n- for_effect: Compare 'namespace' environment with 'local'\n\n` + "```\\nfunc() { ... }\\n```" + `\n\nShould be a part of the same code block\n",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": [
			  {
				"name": "sub-usage",
				"aliases": [
				  "su"
				],
				"usage": "standard usage text",
				"usageText": "Single line of UsageText",
				"argsUsage": "",
				"version": "",
				"description": "",
				"defaultCommand": "",
				"category": "",
				"commands": null,
				"flags": [
				  {
					"name": "sub-command-flag",
					"category": "",
					"defaultText": "",
					"usage": "some usage text",
					"required": false,
					"hidden": false,
					"hideDefault": false,
					"persistent": false,
					"defaultValue": false,
					"aliases": [
					  "s"
					],
					"takesFileArg": false,
					"config": {
					  "Count": null
					},
					"onlyOnce": false
				  }
				],
				"hideHelp": false,
				"hideHelpCommand": false,
				"hideVersion": false,
				"hidden": false,
				"authors": null,
				"copyright": "",
				"metadata": null,
				"sliceFlagSeparator": "",
				"disableSliceFlagSeparator": false,
				"useShortOptionHandling": false,
				"suggest": false,
				"allowExtFlags": false,
				"skipFlagParsing": false,
				"prefixMatchCommands": false,
				"mutuallyExclusiveFlags": null,
				"arguments": null,
				"readArgsFromStdin": false
			  }
			],
			"flags": [
			  {
				"name": "flag",
				"category": "",
				"defaultText": "",
				"usage": "",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"persistent": false,
				"defaultValue": "",
				"aliases": [
				  "fl",
				  "f"
				],
				"takesFileArg": true,
				"config": {
				  "TrimSpace": false
				},
				"onlyOnce": false
			  },
			  {
				"name": "another-flag",
				"category": "",
				"defaultText": "",
				"usage": "another usage text",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"persistent": false,
				"defaultValue": false,
				"aliases": [
				  "b"
				],
				"takesFileArg": false,
				"config": {
				  "Count": null
				},
				"onlyOnce": false
			  }
			],
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  }
		],
		"flags": [
		  {
			"name": "socket",
			"category": "",
			"defaultText": "",
			"usage": "some 'usage' text",
			"required": false,
			"hidden": false,
			"hideDefault": false,
			"persistent": false,
			"defaultValue": "value",
			"aliases": [
			  "s"
			],
			"takesFileArg": true,
			"config": {
			  "TrimSpace": false
			},
			"onlyOnce": false
		  },
		  {
			"name": "flag",
			"category": "",
			"defaultText": "",
			"usage": "",
			"required": false,
			"hidden": false,
			"hideDefault": false,
			"persistent": false,
			"defaultValue": "",
			"aliases": [
			  "fl",
			  "f"
			],
			"takesFileArg": false,
			"config": {
			  "TrimSpace": false
			},
			"onlyOnce": false
		  },
		  {
			"name": "another-flag",
			"category": "",
			"defaultText": "",
			"usage": "another usage text",
			"required": false,
			"hidden": false,
			"hideDefault": false,
			"persistent": false,
			"defaultValue": false,
			"aliases": [
			  "b"
			],
			"takesFileArg": false,
			"config": {
			  "Count": null
			},
			"onlyOnce": false
		  },
		  {
			"name": "hidden-flag",
			"category": "",
			"defaultText": "",
			"usage": "",
			"required": false,
			"hidden": true,
			"hideDefault": false,
			"persistent": false,
			"defaultValue": false,
			"aliases": null,
			"takesFileArg": false,
			"config": {
			  "Count": null
			},
			"onlyOnce": false
		  }
		],
		"hideHelp": false,
		"hideHelpCommand": false,
		"hideVersion": false,
		"hidden": false,
		"authors": [
		  "Harrison <harrison@lolwut.example.com>",
		  {
			"Name": "Oliver Allen",
			"Address": "oliver@toyshop.com"
		  }
		],
		"copyright": "",
		"metadata": null,
		"sliceFlagSeparator": "",
		"disableSliceFlagSeparator": false,
		"useShortOptionHandling": false,
		"suggest": false,
		"allowExtFlags": false,
		"skipFlagParsing": false,
		"prefixMatchCommands": false,
		"mutuallyExclusiveFlags": null,
		"arguments": [
		  {
			"name": "fooi",
			"value": 0,
			"usageText": "",
			"minTimes": 0,
			"maxTimes": 0,
			"config": {
			  "Base": 0
			}
		  }
		],
		"readArgsFromStdin": false
	  }
`
	itesting.JSONEq(t, expected, string(out))
}
