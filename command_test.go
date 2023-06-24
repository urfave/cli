package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		expectedErr            error
	}{
		// Test normal "not ignoring flags" flow
		{testArgs: []string{"test-cmd", "-break", "blah", "blah"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("flag provided but not defined: -break")},
		{testArgs: []string{"test-cmd", "blah", "blah"}, skipFlagParsing: true, useShortOptionHandling: false, expectedErr: nil},   // Test SkipFlagParsing without any args that look like flags
		{testArgs: []string{"test-cmd", "blah", "-break"}, skipFlagParsing: true, useShortOptionHandling: false, expectedErr: nil}, // Test SkipFlagParsing with random flag arg
		{testArgs: []string{"test-cmd", "blah", "-help"}, skipFlagParsing: true, useShortOptionHandling: false, expectedErr: nil},  // Test SkipFlagParsing with "special" help flag arg
		{testArgs: []string{"test-cmd", "blah", "-h"}, skipFlagParsing: false, useShortOptionHandling: true, expectedErr: nil},     // Test UseShortOptionHandling
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs, " "), func(t *testing.T) {
			cmd := &Command{Writer: io.Discard}
			set := flag.NewFlagSet("test", 0)
			_ = set.Parse(c.testArgs)

			cCtx := NewContext(cmd, set, nil)

			subCmd := &Command{
				Name:            "test-cmd",
				Aliases:         []string{"tc"},
				Usage:           "this is for testing",
				Description:     "testing",
				Action:          func(_ *Context) error { return nil },
				SkipFlagParsing: c.skipFlagParsing,
			}

			ctx, cancel := context.WithTimeout(cCtx.Context, 100*time.Millisecond)
			t.Cleanup(cancel)

			err := subCmd.Run(ctx, c.testArgs)

			expect(t, err, c.expectedErr)
			// expect(t, cCtx.Args().Slice(), c.testArgs)
		})
	}
}

func TestParseAndRunShortOpts(t *testing.T) {
	testCases := []struct {
		testArgs     args
		expectedErr  string
		expectedArgs Args
	}{
		{testArgs: args{"test", "-a"}},
		{testArgs: args{"test", "-c", "arg1", "arg2"}, expectedArgs: &args{"arg1", "arg2"}},
		{testArgs: args{"test", "-f"}, expectedArgs: &args{}},
		{testArgs: args{"test", "-ac", "--fgh"}, expectedArgs: &args{}},
		{testArgs: args{"test", "-af"}, expectedArgs: &args{}},
		{testArgs: args{"test", "-cf"}, expectedArgs: &args{}},
		{testArgs: args{"test", "-acf"}, expectedArgs: &args{}},
		{testArgs: args{"test", "--acf"}, expectedErr: "flag provided but not defined: -acf"},
		{testArgs: args{"test", "-invalid"}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: args{"test", "-acf", "-invalid"}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: args{"test", "--invalid"}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: args{"test", "-acf", "--invalid"}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: args{"test", "-acf", "arg1", "-invalid"}, expectedArgs: &args{"arg1", "-invalid"}},
		{testArgs: args{"test", "-acf", "arg1", "--invalid"}, expectedArgs: &args{"arg1", "--invalid"}},
		{testArgs: args{"test", "-acfi", "not-arg", "arg1", "-invalid"}, expectedArgs: &args{"arg1", "-invalid"}},
		{testArgs: args{"test", "-i", "ivalue"}, expectedArgs: &args{}},
		{testArgs: args{"test", "-i", "ivalue", "arg1"}, expectedArgs: &args{"arg1"}},
		{testArgs: args{"test", "-i"}, expectedErr: "flag needs an argument: -i"},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.testArgs, " "), func(t *testing.T) {
			state := map[string]Args{"args": nil}

			cmd := &Command{
				Name:        "test",
				Usage:       "this is for testing",
				Description: "testing",
				Action: func(c *Context) error {
					state["args"] = c.Args()
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

			err := cmd.Run(buildTestContext(t), tc.testArgs)

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
		Before: func(*Context) error {
			return fmt.Errorf("before error")
		},
		After: func(*Context) error {
			return fmt.Errorf("after error")
		},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"bar"})
	r := require.New(t)

	r.ErrorContains(err, "before error")
	r.ErrorContains(err, "after error")
}

func TestCommand_Run_BeforeSavesMetadata(t *testing.T) {
	var receivedMsgFromAction string
	var receivedMsgFromAfter string

	cmd := &Command{
		Name: "bar",
		Before: func(c *Context) error {
			c.Command.Metadata["msg"] = "hello world"
			return nil
		},
		Action: func(c *Context) error {
			msg, ok := c.Command.Metadata["msg"]
			if !ok {
				return errors.New("msg not found")
			}
			receivedMsgFromAction = msg.(string)
			return nil
		},
		After: func(c *Context) error {
			msg, ok := c.Command.Metadata["msg"]
			if !ok {
				return errors.New("msg not found")
			}
			receivedMsgFromAfter = msg.(string)
			return nil
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "bar"})
	if err != nil {
		t.Fatalf("expected no error from Run, got %s", err)
	}

	expectedMsg := "hello world"

	if receivedMsgFromAction != expectedMsg {
		t.Fatalf("expected msg from Action to match. Given: %q\nExpected: %q",
			receivedMsgFromAction, expectedMsg)
	}
	if receivedMsgFromAfter != expectedMsg {
		t.Fatalf("expected msg from After to match. Given: %q\nExpected: %q",
			receivedMsgFromAction, expectedMsg)
	}
}

func TestCommand_OnUsageError_hasCommandContext(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(c *Context, err error, _ bool) error {
			return fmt.Errorf("intercepted in %s: %s", c.Command.Name, err.Error())
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted in bar") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

func TestCommand_OnUsageError_WithWrongFlagValue(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ *Context, err error, _ bool) error {
			if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
				t.Errorf("Expect an invalid value error, but got \"%v\"", err)
			}
			return errors.New("intercepted: " + err.Error())
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
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
		OnUsageError: func(_ *Context, err error, _ bool) error {
			if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
				t.Errorf("Expect an invalid value error, but got \"%v\"", err)
			}
			return errors.New("intercepted: " + err.Error())
		},
	}

	require.ErrorContains(t, cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"}), "intercepted: invalid value")
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
				Action: func(cCtx *Context) error {
					require.Equal(t, io.Discard, cCtx.Command.Root().ErrWriter)

					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"bar", "baz"}))
}

func TestCommandSkipFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs     args
		expectedArgs *args
		expectedErr  error
	}{
		{testArgs: args{"some-command", "some-arg", "--flag", "foo"}, expectedArgs: &args{"some-arg", "--flag", "foo"}, expectedErr: nil},
		{testArgs: args{"some-command", "some-arg", "--flag=foo"}, expectedArgs: &args{"some-arg", "--flag=foo"}, expectedErr: nil},
	}

	for _, c := range cases {
		var args Args
		cmd := &Command{
			SkipFlagParsing: true,
			Name:            "some-command",
			Flags: []Flag{
				&StringFlag{Name: "flag"},
			},
			Action: func(c *Context) error {
				args = c.Args()
				return nil
			},
			Writer: io.Discard,
		}

		err := cmd.Run(buildTestContext(t), c.testArgs)
		expect(t, err, c.expectedErr)
		expect(t, args, c.expectedArgs)
	}
}

func TestCommand_Run_CustomShellCompleteAcceptsMalformedFlags(t *testing.T) {
	cases := []struct {
		testArgs    args
		expectedOut string
	}{
		{testArgs: args{"--undefined"}, expectedOut: "found 0 args"},
		{testArgs: args{"--number"}, expectedOut: "found 0 args"},
		{testArgs: args{"--number", "forty-two"}, expectedOut: "found 0 args"},
		{testArgs: args{"--number", "42"}, expectedOut: "found 0 args"},
		{testArgs: args{"--number", "42", "newArg"}, expectedOut: "found 1 args"},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs, " "), func(t *testing.T) {
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
				ShellComplete: func(cCtx *Context) {
					fmt.Fprintf(cCtx.Command.Root().Writer, "found %[1]d args", cCtx.NArg())
				},
			}

			osArgs := args{"bar"}
			osArgs = append(osArgs, c.testArgs...)
			osArgs = append(osArgs, "--generate-shell-completion")

			r := require.New(t)

			r.NoError(cmd.Run(buildTestContext(t), osArgs))
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
	expect(t, err, nil)
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

	expect(t, cmd.VisibleCommands(), []*Command{subc1, subc3})
}

func TestCommand_VisibleFlagCategories(t *testing.T) {
	cmd := &Command{
		Name:  "bar",
		Usage: "this is for testing",
		Flags: []Flag{
			&StringFlag{
				Name: "strd", // no category set
			},
			&IntFlag{
				Name:     "intd",
				Aliases:  []string{"altd1", "altd2"},
				Category: "cat1",
			},
		},
	}

	vfc := cmd.VisibleFlagCategories()
	if len(vfc) != 1 {
		t.Fatalf("unexpected visible flag categories %+v", vfc)
	}
	if vfc[0].Name() != "cat1" {
		t.Errorf("expected category name cat1 got %s", vfc[0].Name())
	}
	if len(vfc[0].Flags()) != 1 {
		t.Fatalf("expected flag category to have just one flag got %+v", vfc[0].Flags())
	}

	fl := vfc[0].Flags()[0]
	if !reflect.DeepEqual(fl.Names(), []string{"intd", "altd1", "altd2"}) {
		t.Errorf("unexpected flag %+v", fl.Names())
	}
}

func TestCommand_RunSubcommandWithDefault(t *testing.T) {
	cmd := &Command{
		Version:        "some version",
		Name:           "app",
		DefaultCommand: "foo",
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(ctx *Context) error {
					return errors.New("should not run this subcommand")
				},
			},
			{
				Name:     "bar",
				Usage:    "this is for testing",
				Commands: []*Command{{}}, // some subcommand
				Action: func(*Context) error {
					return nil
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"app", "bar"})
	expect(t, err, nil)

	err = cmd.Run(buildTestContext(t), []string{"app"})
	expect(t, err, errors.New("should not run this subcommand"))
}

func TestCommand_Run(t *testing.T) {
	s := ""

	cmd := &Command{
		Action: func(c *Context) error {
			s = s + c.Args().First()
			return nil
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"command", "foo"})
	expect(t, err, nil)
	err = cmd.Run(buildTestContext(t), []string{"command", "bar"})
	expect(t, err, nil)
	expect(t, s, "foobar")
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
		expect(t, cmd.Command(test.name) != nil, test.expected)
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
			expect(t, err == nil, test.expected)
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
			expect(t, err == nil, test.expected)
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
			expect(t, err == nil, test.expected)
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
	if err != nil {
		t.Error(err)
	}

	if someint != 10 {
		t.Errorf("Expected 10 got %d for someint", someint)
	}

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
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCommand_Setup_defaultsReader(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})
	expect(t, cmd.Reader, os.Stdin)
}

func TestCommand_Setup_defaultsWriter(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})
	expect(t, cmd.Writer, os.Stdout)
}

func TestCommand_RunAsSubcommandParseFlags(t *testing.T) {
	var cCtx *Context

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(c *Context) error {
					cCtx = c
					return nil
				},
				Flags: []Flag{
					&StringFlag{
						Name:  "lang",
						Value: "english",
						Usage: "language for the greeting",
					},
				},
				Before: func(_ *Context) error { return nil },
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "foo", "--lang", "spanish", "abcd"})

	expect(t, cCtx.Args().Get(0), "abcd")
	expect(t, cCtx.String("lang"), "spanish")
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
				Action: func(c *Context) error {
					parsedOption = c.String("option")
					args = c.Args()
					return nil
				},
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"})

	expect(t, parsedOption, "my-option")
	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "--notARealFlag")
}

func TestCommand_CommandWithDash(t *testing.T) {
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(c *Context) error {
					args = c.Args()
					return nil
				},
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "cmd", "my-arg", "-"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "-")
}

func TestCommand_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(c *Context) error {
					args = c.Args()
					return nil
				},
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "cmd", "my-arg", "--", "notAFlagAtAll"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "notAFlagAtAll")
}

func TestCommand_SkipFlagParsing(t *testing.T) {
	var args Args

	cmd := &Command{
		SkipFlagParsing: true,
		Action: func(c *Context) error {
			args = c.Args()
			return nil
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "--", "my-arg", "notAFlagAtAll"})

	expect(t, args.Get(0), "--")
	expect(t, args.Get(1), "my-arg")
	expect(t, args.Get(2), "notAFlagAtAll")
}

func TestCommand_VisibleCommands(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{
				Name:   "frob",
				Action: func(_ *Context) error { return nil },
			},
			{
				Name:   "frib",
				Hidden: true,
				Action: func(_ *Context) error { return nil },
			},
		},
	}

	cmd.setupDefaults([]string{"cli.test"})
	expected := []*Command{
		cmd.Commands[0],
		cmd.Commands[2], // help
	}
	actual := cmd.VisibleCommands()
	expect(t, len(expected), len(actual))
	for i, actualCommand := range actual {
		expectedCommand := expected[i]

		if expectedCommand.Action != nil {
			// comparing func addresses is OK!
			expect(t, fmt.Sprintf("%p", expectedCommand.Action), fmt.Sprintf("%p", actualCommand.Action))
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

			if !reflect.DeepEqual(expectedCommand, actualCommand) {
				t.Errorf("expected\n%#v\n!=\n%#v", expectedCommand, actualCommand)
			}
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
	cmd.Action = func(c *Context) error {
		one = c.Bool("one")
		two = c.Bool("two")
		name = c.String("name")
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "-on", expected})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestCommand_UseShortOptionHandling_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
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
		Action: func(c *Context) error {
			one = c.Bool("one")
			two = c.Bool("two")
			name = c.String("name")
			return nil
		},
		UseShortOptionHandling: true,
		Writer:                 io.Discard,
	}

	r := require.New(t)
	r.Nil(cmd.Run(buildTestContext(t), []string{"cmd", "-on", expected}))
	r.True(one)
	r.False(two)
	r.Equal(expected, name)
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

	require.ErrorContains(
		t,
		cmd.Run(buildTestContext(t), []string{"", "cmd", "-n"}),
		"flag needs an argument: -n",
	)
}

func TestCommand_UseShortOptionHandlingSubCommand(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
	}
	subCommand := &Command{
		Name: "sub",
		Flags: []Flag{
			&BoolFlag{Name: "one", Aliases: []string{"o"}},
			&BoolFlag{Name: "two", Aliases: []string{"t"}},
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Action: func(c *Context) error {
			one = c.Bool("one")
			two = c.Bool("two")
			name = c.String("name")
			return nil
		},
	}
	command.Commands = []*Command{subCommand}
	cmd.Commands = []*Command{command}

	err := cmd.Run(buildTestContext(t), []string{"", "cmd", "sub", "-on", expected})
	expect(t, err, nil)
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
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
	expect(t, err, errors.New("flag needs an argument: -n"))
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
	cmd.Action = func(c *Context) error {
		sliceVal = c.StringSlice("env")
		one = c.Bool("one")
		two = c.Bool("two")
		name = c.String("name")
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "-e", "foo", "-on", expected})
	expect(t, sliceVal, []string{"foo"})
	expect(t, sliceValDest, []string{"foo"})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestCommand_Float64Flag(t *testing.T) {
	var meters float64

	cmd := &Command{
		Flags: []Flag{
			&FloatFlag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
		},
		Action: func(c *Context) error {
			meters = c.Float("height")
			return nil
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "--height", "1.93"})
	expect(t, meters, 1.93)
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
				Action: func(c *Context) error {
					parsedIntSlice = c.IntSlice("p")
					parsedStringSlice = c.StringSlice("ip")
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
				Action: func(c *Context) error {
					parsedIntSlice = c.IntSlice("a")
					parsedStringSlice = c.StringSlice("str")
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

	if cmd.Reader != os.Stdin {
		t.Error("Default input reader not set.")
	}
}

func TestCommand_DefaultStdout(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"cli.test"})

	if cmd.Writer != os.Stdout {
		t.Error("Default output writer not set.")
	}
}

func TestCommand_SetStdin(t *testing.T) {
	buf := make([]byte, 12)

	cmd := &Command{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Action: func(c *Context) error {
			_, err := c.Command.Reader.Read(buf)
			return err
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"help"})
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if string(buf) != "Hello World!" {
		t.Error("Command did not read input from desired reader.")
	}
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
						Action: func(c *Context) error {
							_, err := c.Command.Root().Reader.Read(buf)
							return err
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"test", "command", "subcommand"})
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if string(buf) != "Hello World!" {
		t.Error("Command did not read input from desired reader.")
	}
}

func TestCommand_SetStdout(t *testing.T) {
	var w bytes.Buffer

	cmd := &Command{
		Name:   "test",
		Writer: &w,
	}

	err := cmd.Run(buildTestContext(t), []string{"help"})
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if w.Len() == 0 {
		t.Error("Command did not write output to desired writer.")
	}
}

func TestCommand_BeforeFunc(t *testing.T) {
	counts := &opCounts{}
	beforeError := fmt.Errorf("fail")
	var err error

	cmd := &Command{
		Before: func(c *Context) error {
			counts.Total++
			counts.Before = counts.Total
			s := c.String("opt")
			if s == "fail" {
				return beforeError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(*Context) error {
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

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	if counts.SubCommand != 2 {
		t.Errorf("Subcommand not executed when expected")
	}

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != beforeError {
		t.Errorf("Run error expected, but not received")
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand executed when NOT expected")
	}

	// reset
	counts = &opCounts{}

	afterError := errors.New("fail again")
	cmd.After = func(_ *Context) error {
		return afterError
	}

	// run with the Before() func failing, wrapped by After()
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if _, ok := err.(MultiError); !ok {
		t.Errorf("MultiError expected, but not received")
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand executed when NOT expected")
	}
}

func TestCommand_BeforeAfterFuncShellCompletion(t *testing.T) {
	t.Skip("TODO: is '--generate-shell-completion' (flag) still supported?")

	counts := &opCounts{}

	cmd := &Command{
		EnableShellCompletion: true,
		Before: func(*Context) error {
			counts.Total++
			counts.Before = counts.Total
			return nil
		},
		After: func(*Context) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(*Context) error {
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
		After: func(c *Context) error {
			counts.Total++
			counts.After = counts.Total
			s := c.String("opt")
			if s == "fail" {
				return afterError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(*Context) error {
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

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.After != 2 {
		t.Errorf("After() not executed when expected")
	}

	if counts.SubCommand != 1 {
		t.Errorf("Subcommand not executed when expected")
	}

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != afterError {
		t.Errorf("Run error expected, but not received")
	}

	if counts.After != 2 {
		t.Errorf("After() not executed when expected")
	}

	if counts.SubCommand != 1 {
		t.Errorf("Subcommand not executed when expected")
	}

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
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.After != 1 {
		t.Errorf("After() not executed when expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand not executed when expected")
	}
}

func TestCommandNoHelpFlag(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = nil

	cmd := &Command{Writer: io.Discard}

	err := cmd.Run(buildTestContext(t), []string{"test", "-h"})

	if err != flag.ErrHelp {
		t.Errorf("expected error about missing help flag, but got: %s (%T)", err, err)
	}
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
					Action: func(c *Context) error {
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
			if test.expectedAnError && err == nil {
				t.Errorf("expected an error, but there was none")
			}
			if _, ok := err.(requiredFlagsErr); test.expectedAnError && !ok {
				t.Errorf("expected a requiredFlagsErr, but got: %s", err)
			}
			if !test.expectedAnError && err != nil {
				t.Errorf("did not expected an error, but there was one: %s", err)
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

	if wasCalled == false {
		t.Errorf("Help printer expected to be called, but was not")
	}
}

func TestCommand_VersionPrinter(t *testing.T) {
	oldPrinter := VersionPrinter
	defer func() {
		VersionPrinter = oldPrinter
	}()

	wasCalled := false
	VersionPrinter = func(*Context) {
		wasCalled = true
	}

	cmd := &Command{}
	cCtx := NewContext(cmd, nil, nil)
	ShowVersion(cCtx)

	if wasCalled == false {
		t.Errorf("Version printer expected to be called, but was not")
	}
}

func TestCommand_CommandNotFound(t *testing.T) {
	counts := &opCounts{}
	cmd := &Command{
		CommandNotFound: func(*Context, string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		},
		Commands: []*Command{
			{
				Name: "bar",
				Action: func(*Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"command", "foo"})

	expect(t, counts.CommandNotFound, 1)
	expect(t, counts.SubCommand, 0)
	expect(t, counts.Total, 1)
}

func TestCommand_OrderOfOperations(t *testing.T) {
	counts := &opCounts{}

	resetCounts := func() { counts = &opCounts{} }

	cmd := &Command{
		EnableShellCompletion: true,
		ShellComplete: func(*Context) {
			counts.Total++
			counts.ShellComplete = counts.Total
		},
		OnUsageError: func(*Context, error, bool) error {
			counts.Total++
			counts.OnUsageError = counts.Total
			return errors.New("hay OnUsageError")
		},
		Writer: io.Discard,
	}

	beforeNoError := func(*Context) error {
		counts.Total++
		counts.Before = counts.Total
		return nil
	}

	beforeError := func(*Context) error {
		counts.Total++
		counts.Before = counts.Total
		return errors.New("hay Before")
	}

	cmd.Before = beforeNoError
	cmd.CommandNotFound = func(*Context, string) {
		counts.Total++
		counts.CommandNotFound = counts.Total
	}

	afterNoError := func(*Context) error {
		counts.Total++
		counts.After = counts.Total
		return nil
	}

	afterError := func(*Context) error {
		counts.Total++
		counts.After = counts.Total
		return errors.New("hay After")
	}

	cmd.After = afterNoError
	cmd.Commands = []*Command{
		{
			Name: "bar",
			Action: func(*Context) error {
				counts.Total++
				counts.SubCommand = counts.Total
				return nil
			},
		},
	}

	cmd.Action = func(*Context) error {
		counts.Total++
		counts.Action = counts.Total
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"command", "--nope"})
	expect(t, counts.OnUsageError, 1)
	expect(t, counts.Total, 1)

	resetCounts()

	_ = cmd.Run(buildTestContext(t), []string{"command", fmt.Sprintf("--%s", "generate-shell-completion")})
	expect(t, counts.ShellComplete, 1)
	expect(t, counts.Total, 1)

	resetCounts()

	oldOnUsageError := cmd.OnUsageError
	cmd.OnUsageError = nil
	_ = cmd.Run(buildTestContext(t), []string{"command", "--nope"})
	expect(t, counts.Total, 0)
	cmd.OnUsageError = oldOnUsageError

	resetCounts()

	_ = cmd.Run(buildTestContext(t), []string{"command", "foo"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.CommandNotFound, 0)
	expect(t, counts.Action, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)

	resetCounts()

	cmd.Before = beforeError
	_ = cmd.Run(buildTestContext(t), []string{"command", "bar"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.After, 2)
	expect(t, counts.Total, 2)
	cmd.Before = beforeNoError

	resetCounts()

	cmd.After = nil
	_ = cmd.Run(buildTestContext(t), []string{"command", "bar"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.SubCommand, 2)
	expect(t, counts.Total, 2)
	cmd.After = afterNoError

	resetCounts()

	cmd.After = afterError
	err := cmd.Run(buildTestContext(t), []string{"command", "bar"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.SubCommand, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)
	cmd.After = afterNoError

	resetCounts()

	oldCommands := cmd.Commands
	cmd.Commands = nil
	_ = cmd.Run(buildTestContext(t), []string{"command"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.Action, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)
	cmd.Commands = oldCommands
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
				Action:      func(c *Context) error { return nil },
				Writer:      buf,
			}

			err := cmd.Run(buildTestContext(t), flagSet)
			if err != nil {
				t.Error(err)
			}

			output := buf.String()

			if strings.Contains(output, "No help topic for") {
				t.Errorf("expect a help topic, got none: \n%q", output)
			}

			for _, shouldContain := range []string{
				cmd.Name, cmd.Description,
				subCmdBar.Name, subCmdBar.Usage,
				subCmdBaz.Name, subCmdBaz.Usage,
			} {
				if !strings.Contains(output, shouldContain) {
					t.Errorf("want help to contain %q, did not: \n%q", shouldContain, output)
				}
			}
		})
	}
}

func TestCommand_Run_SubcommandFullPath(t *testing.T) {
	out := &bytes.Buffer{}

	subCmd := &Command{
		Name:  "bar",
		Usage: "does bar things",
	}

	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Commands:    []*Command{subCmd},
		Writer:      out,
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "bar", "--help"}))

	outString := out.String()
	r.Contains(outString, "foo bar - does bar things")
	r.Contains(outString, "foo bar [command [command options]] [arguments...]")
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
				Action: func(*Context) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), tt.helpArguments)
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("want err: %s, did note %s\n", tt.wantErr, err)
			}

			output := buf.String()

			if !strings.Contains(output, tt.wantContains) {
				t.Errorf("want help to contain %q, did not: \n%q", "boom - make an explosive entrance", output)
			}
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
				Action: func(*Context) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), args)
			if err != nil {
				t.Error(err)
			}

			output := buf.String()

			if !strings.Contains(output, "0.1.0") {
				t.Errorf("want version to contain %q, did not: \n%q", "0.1.0", output)
			}
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

	if !reflect.DeepEqual(cmd.categories, &expect) {
		t.Fatalf("expected categories %#v, to equal %#v", cmd.categories, &expect)
	}

	output := buf.String()

	if !strings.Contains(output, "1:\n     command1") {
		t.Errorf("want buffer to include category %q, did not: \n%q", "1:\n     command1", output)
	}
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
	expect(t, expected, cmd.VisibleCategories())

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
	expect(t, expected, cmd.VisibleCategories())

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
	expect(t, []CommandCategory{}, cmd.VisibleCategories())
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
				Before: func(c *Context) error { return fmt.Errorf("before error") },
				After:  func(c *Context) error { return fmt.Errorf("after error") },
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.Contains(err.Error(), "before error") {
		t.Errorf("expected text of error from Before method, but got none in \"%v\"", err)
	}
	if !strings.Contains(err.Error(), "after error") {
		t.Errorf("expected text of error from After method, but got none in \"%v\"", err)
	}
}

func TestCommand_OnUsageError_WithWrongFlagValue_ForSubcommand(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ *Context, err error, isSubcommand bool) error {
			if isSubcommand {
				t.Errorf("Expect subcommand")
			}
			if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
				t.Errorf("Expect an invalid value error, but got \"%v\"", err)
			}
			return errors.New("intercepted: " + err.Error())
		},
		Commands: []*Command{
			{
				Name: "bar",
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--flag=wrong", "bar"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
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

func (c *customBoolFlag) RunAction(*Context) error {
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
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestCustomFlagsUsed(t *testing.T) {
	cmd := &Command{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--custom=bar"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
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
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHandleExitCoder_Default(t *testing.T) {
	app := buildMinimalTestCommand()
	fs, err := flagSet(app.Name, app.Flags)
	if err != nil {
		t.Errorf("error creating FlagSet: %s", err)
	}

	cCtx := NewContext(app, fs, nil)
	_ = app.handleExitCoder(cCtx, Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	if !strings.Contains(output, "Default") {
		t.Fatalf("Expected Default Behavior from Error Handler but got: %s", output)
	}
}

func TestHandleExitCoder_Custom(t *testing.T) {
	cmd := buildMinimalTestCommand()
	fs, err := flagSet(cmd.Name, cmd.Flags)
	if err != nil {
		t.Errorf("error creating FlagSet: %s", err)
	}

	cmd.ExitErrHandler = func(_ *Context, _ error) {
		_, _ = fmt.Fprintln(ErrWriter, "I'm a Custom error handler, I print what I want!")
	}

	ctx := NewContext(cmd, fs, nil)
	_ = cmd.handleExitCoder(ctx, Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	if !strings.Contains(output, "Custom") {
		t.Fatalf("Expected Custom Behavior from Error Handler but got: %s", output)
	}
}

func TestShellCompletionForIncompleteFlags(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{
				Name: "test-completion",
			},
		},
		EnableShellCompletion: true,
		ShellComplete: func(ctx *Context) {
			for _, command := range ctx.Command.Commands {
				if command.Hidden {
					continue
				}

				for _, name := range command.Names() {
					_, _ = fmt.Fprintln(ctx.Command.Writer, name)
				}
			}

			for _, fl := range ctx.Command.Flags {
				for _, name := range fl.Names() {
					if name == GenerateShellCompletionFlag.Names()[0] {
						continue
					}

					switch name = strings.TrimSpace(name); len(name) {
					case 0:
					case 1:
						_, _ = fmt.Fprintln(ctx.Command.Writer, "-"+name)
					default:
						_, _ = fmt.Fprintln(ctx.Command.Writer, "--"+name)
					}
				}
			}
		},
		Action: func(ctx *Context) error {
			return fmt.Errorf("should not get here")
		},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--test-completion", "--" + "generate-shell-completion"})
	if err != nil {
		t.Errorf("app should not return an error: %s", err)
	}
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
					Action: func(c *Context) error {
						return Exit("exit error", testCode)
					},
				},
			},
		},
	}

	// set user function as ExitErrHandler
	exitCodeFromExitErrHandler := int(0)
	cmd.ExitErrHandler = func(_ *Context, err error) {
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
	return &Command{Writer: io.Discard}
}

func TestSetupInitializesBothWriters(t *testing.T) {
	cmd := &Command{}

	cmd.setupDefaults([]string{"cli.test"})

	if cmd.ErrWriter != os.Stderr {
		t.Errorf("expected a.ErrWriter to be os.Stderr")
	}

	if cmd.Writer != os.Stdout {
		t.Errorf("expected a.Writer to be os.Stdout")
	}
}

func TestSetupInitializesOnlyNilWriters(t *testing.T) {
	wr := &bytes.Buffer{}
	cmd := &Command{
		ErrWriter: wr,
	}

	cmd.setupDefaults([]string{"cli.test"})

	if cmd.ErrWriter != wr {
		t.Errorf("expected a.ErrWriter to be a *bytes.Buffer instance")
	}

	if cmd.Writer != os.Stdout {
		t.Errorf("expected a.Writer to be os.Stdout")
	}
}

func TestFlagAction(t *testing.T) {
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
			args: []string{"app", "--f_timestamp", "2022-05-01 02:26:20"},
			exp:  "2022-05-01T02:26:20Z ",
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
				Action: func(cCtx *Context, v string) error {
					if v == "" {
						return fmt.Errorf("empty string")
					}
					_, err := cCtx.Command.Root().Writer.Write([]byte(v + " "))
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
						Action: func(cCtx *Context) error { return nil },
						Commands: []*Command{
							{
								Name:   "sub1",
								Action: func(cCtx *Context) error { return nil },
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
						Action: func(cCtx *Context, v []string) error {
							if v[0] == "err" {
								return fmt.Errorf("error string slice")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&BoolFlag{
						Name: "f_bool",
						Action: func(cCtx *Context, v bool) error {
							if !v {
								return fmt.Errorf("value is false")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%t ", v)))
							return err
						},
					},
					&DurationFlag{
						Name: "f_duration",
						Action: func(cCtx *Context, v time.Duration) error {
							if v == 0 {
								return fmt.Errorf("empty duration")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(v.String() + " "))
							return err
						},
					},
					&FloatFlag{
						Name: "f_float64",
						Action: func(cCtx *Context, v float64) error {
							if v < 0 {
								return fmt.Errorf("negative float64")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64) + " "))
							return err
						},
					},
					&FloatSliceFlag{
						Name: "f_float64_slice",
						Action: func(cCtx *Context, v []float64) error {
							if len(v) > 0 && v[0] < 0 {
								return fmt.Errorf("invalid float64 slice")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&IntFlag{
						Name: "f_int",
						Action: func(cCtx *Context, v int64) error {
							if v < 0 {
								return fmt.Errorf("negative int")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&IntSliceFlag{
						Name: "f_int_slice",
						Action: func(cCtx *Context, v []int64) error {
							if len(v) > 0 && v[0] < 0 {
								return fmt.Errorf("invalid int slice")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&TimestampFlag{
						Name: "f_timestamp",
						Config: TimestampConfig{
							Layout: "2006-01-02 15:04:05",
						},
						Action: func(cCtx *Context, v time.Time) error {
							if v.IsZero() {
								return fmt.Errorf("zero timestamp")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(v.Format(time.RFC3339) + " "))
							return err
						},
					},
					&UintFlag{
						Name: "f_uint",
						Action: func(cCtx *Context, v uint64) error {
							if v == 0 {
								return fmt.Errorf("zero uint64")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%v ", v)))
							return err
						},
					},
					&StringMapFlag{
						Name: "f_string_map",
						Action: func(cCtx *Context, v map[string]string) error {
							if _, ok := v["err"]; ok {
								return fmt.Errorf("error string map")
							}
							_, err := cCtx.Command.Root().Writer.Write([]byte(fmt.Sprintf("%v", v)))
							return err
						},
					},
				},
				Action: func(cCtx *Context) error { return nil },
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
				Action: func(ctx *Context, s string) error {
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
						Action: func(ctx *Context) error {
							appSliceFloat64 = ctx.FloatSlice("persistentCommandFloatSliceFlag")
							return nil
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"app",
		"--persistentCommandFlag", "hello",
		"--persistentCommandSliceFlag", "100",
		"--persistentCommandOverrideFlag", "102",
		"cmd",
		"--cmdFlag", "12",
		"--persistentCommandSliceFlag", "102",
		"--persistentCommandFloatSliceFlag", "102.455",
		"--paof", "105",
		"subcmd",
		"--cmdPersistentFlag", "20",
		"--cmdFlag", "11",
		"--persistentCommandFlag", "bar",
		"--persistentCommandSliceFlag", "130",
		"--persistentCommandFloatSliceFlag", "3.1445",
	})

	if err != nil {
		t.Fatal(err)
	}

	if appFlag != "bar" {
		t.Errorf("Expected 'bar' got %s", appFlag)
	}

	if topInt != 12 {
		t.Errorf("Expected 12 got %d", topInt)
	}

	if topPersistentInt != 20 {
		t.Errorf("Expected 20 got %d", topPersistentInt)
	}

	// this should be changed from app since
	// cmd overrides it
	if appOverrideInt != 102 {
		t.Errorf("Expected 102 got %d", appOverrideInt)
	}

	if subCommandInt != 11 {
		t.Errorf("Expected 11 got %d", subCommandInt)
	}

	if appOverrideCmdInt != 105 {
		t.Errorf("Expected 105 got %d", appOverrideCmdInt)
	}

	expectedInt := []int64{100, 102, 130}
	if !reflect.DeepEqual(persistentCommandSliceInt, expectedInt) {
		t.Errorf("Expected %v got %d", expectedInt, persistentCommandSliceInt)
	}

	expectedFloat := []float64{102.455, 3.1445}
	if !reflect.DeepEqual(appSliceFloat64, expectedFloat) {
		t.Errorf("Expected %f got %f", expectedFloat, appSliceFloat64)
	}

	if persistentFlagActionCount != 2 {
		t.Errorf("Expected persistent flag action to be called 2 times instead called %d", persistentFlagActionCount)
	}
}

func TestFlagDuplicates(t *testing.T) {
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
		Action: func(ctx *Context) error {
			return nil
		},
	}

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
			err := cmd.Run(buildTestContext(t), test.args)
			if test.errExpected && err == nil {
				t.Error("expected error")
			} else if !test.errExpected && err != nil {
				t.Error(err)
			}
		})
	}
}

func TestShorthandCommand(t *testing.T) {
	af := func(p *int) ActionFunc {
		return func(ctx *Context) error {
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
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 1 && cmd2 != 0 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthd"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 1 && cmd2 != 0 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthe"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 1 && cmd2 != 0 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthert"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 0 && cmd2 != 1 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthet"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 0 && cmd2 != 1 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}
}
