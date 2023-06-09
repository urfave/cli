package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
			Sources: ValueSources{EnvSource("EXAMPLE_VARIABLE_NAME")},
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
		cmd := &Command{Writer: io.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		cCtx := NewContext(cmd, set, nil)

		command := Command{
			Name:            "test-cmd",
			Aliases:         []string{"tc"},
			Usage:           "this is for testing",
			Description:     "testing",
			Action:          func(_ *Context) error { return nil },
			SkipFlagParsing: c.skipFlagParsing,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		t.Cleanup(cancel)

		ctx = context.WithValue(ctx, contextContextKey, cCtx)
		err := command.Run(ctx, c.testArgs)

		expect(t, err, c.expectedErr)
		// expect(t, cCtx.Args().Slice(), c.testArgs)
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

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, tc.testArgs)

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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"bar"})
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "bar"})
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"bar", "--flag=wrong"})
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"bar", "--flag=wrong"})
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	require.ErrorContains(t, cmd.Run(ctx, []string{"bar", "--flag=wrong"}), "intercepted: invalid value")
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
					require.Equal(t, io.Discard, cCtx.Command.errWriter())

					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	require.NoError(t, cmd.Run(ctx, []string{"bar", "baz"}))
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

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		t.Cleanup(cancel)

		err := cmd.Run(ctx, c.testArgs)
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
					fmt.Fprintf(cCtx.Command.writer(), "found %[1]d args", cCtx.NArg())
				},
			}

			osArgs := args{"bar"}
			osArgs = append(osArgs, c.testArgs...)
			osArgs = append(osArgs, "--generate-shell-completion")

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			r := require.New(t)

			r.NoError(cmd.Run(ctx, osArgs))
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "bar"})
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
			&Int64Flag{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"app", "bar"})
	expect(t, err, nil)

	err = cmd.Run(ctx, []string{"app"})
	expect(t, err, errors.New("should not run this subcommand"))
}
