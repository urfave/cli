package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

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
		app := &App{Writer: io.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		cCtx := NewContext(app, set, nil)

		command := Command{
			Name:            "test-cmd",
			Aliases:         []string{"tc"},
			Usage:           "this is for testing",
			Description:     "testing",
			Action:          func(_ *Context) error { return nil },
			SkipFlagParsing: c.skipFlagParsing,
			isRoot:          true,
		}

		err := command.Run(cCtx, c.testArgs...)

		expect(t, err, c.expectedErr)
		//expect(t, cCtx.Args().Slice(), c.testArgs)
	}
}

func TestParseAndRunShortOpts(t *testing.T) {
	cases := []struct {
		testArgs     args
		expectedErr  error
		expectedArgs Args
	}{
		{testArgs: args{"foo", "test", "-a"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "-c", "arg1", "arg2"}, expectedErr: nil, expectedArgs: &args{"arg1", "arg2"}},
		{testArgs: args{"foo", "test", "-f"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "-ac", "--fgh"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "-af"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "-cf"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "-acf"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "--acf"}, expectedErr: errors.New("flag provided but not defined: -acf"), expectedArgs: nil},
		{testArgs: args{"foo", "test", "-invalid"}, expectedErr: errors.New("flag provided but not defined: -invalid"), expectedArgs: nil},
		{testArgs: args{"foo", "test", "-acf", "-invalid"}, expectedErr: errors.New("flag provided but not defined: -invalid"), expectedArgs: nil},
		{testArgs: args{"foo", "test", "--invalid"}, expectedErr: errors.New("flag provided but not defined: -invalid"), expectedArgs: nil},
		{testArgs: args{"foo", "test", "-acf", "--invalid"}, expectedErr: errors.New("flag provided but not defined: -invalid"), expectedArgs: nil},
		{testArgs: args{"foo", "test", "-acf", "arg1", "-invalid"}, expectedErr: nil, expectedArgs: &args{"arg1", "-invalid"}},
		{testArgs: args{"foo", "test", "-acf", "arg1", "--invalid"}, expectedErr: nil, expectedArgs: &args{"arg1", "--invalid"}},
		{testArgs: args{"foo", "test", "-acfi", "not-arg", "arg1", "-invalid"}, expectedErr: nil, expectedArgs: &args{"arg1", "-invalid"}},
		{testArgs: args{"foo", "test", "-i", "ivalue"}, expectedErr: nil, expectedArgs: &args{}},
		{testArgs: args{"foo", "test", "-i", "ivalue", "arg1"}, expectedErr: nil, expectedArgs: &args{"arg1"}},
		{testArgs: args{"foo", "test", "-i"}, expectedErr: errors.New("flag needs an argument: -i"), expectedArgs: nil},
	}

	for _, c := range cases {
		var args Args
		cmd := &Command{
			Name:        "test",
			Usage:       "this is for testing",
			Description: "testing",
			Action: func(c *Context) error {
				args = c.Args()
				return nil
			},
			UseShortOptionHandling: true,
			Flags: []Flag{
				&BoolFlag{Name: "abc", Aliases: []string{"a"}},
				&BoolFlag{Name: "cde", Aliases: []string{"c"}},
				&BoolFlag{Name: "fgh", Aliases: []string{"f"}},
				&StringFlag{Name: "ijk", Aliases: []string{"i"}},
			},
		}

		app := newTestApp()
		app.Commands = []*Command{cmd}

		err := app.Run(c.testArgs)

		expect(t, err, c.expectedErr)
		expect(t, args, c.expectedArgs)
	}
}

func TestCommand_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "bar",
				Before: func(c *Context) error {
					return fmt.Errorf("before error")
				},
				After: func(c *Context) error {
					return fmt.Errorf("after error")
				},
			},
		},
		Writer: io.Discard,
	}

	err := app.Run([]string{"foo", "bar"})
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

func TestCommand_Run_BeforeSavesMetadata(t *testing.T) {
	var receivedMsgFromAction string
	var receivedMsgFromAfter string

	app := &App{
		Commands: []*Command{
			{
				Name: "bar",
				Before: func(c *Context) error {
					c.App.Metadata["msg"] = "hello world"
					return nil
				},
				Action: func(c *Context) error {
					msg, ok := c.App.Metadata["msg"]
					if !ok {
						return errors.New("msg not found")
					}
					receivedMsgFromAction = msg.(string)
					return nil
				},
				After: func(c *Context) error {
					msg, ok := c.App.Metadata["msg"]
					if !ok {
						return errors.New("msg not found")
					}
					receivedMsgFromAfter = msg.(string)
					return nil
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar"})
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
	app := &App{
		Commands: []*Command{
			{
				Name: "bar",
				Flags: []Flag{
					&IntFlag{Name: "flag"},
				},
				OnUsageError: func(c *Context, err error, _ bool) error {
					return fmt.Errorf("intercepted in %s: %s", c.Command.Name, err.Error())
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted in bar") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

func TestCommand_OnUsageError_WithWrongFlagValue(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "bar",
				Flags: []Flag{
					&IntFlag{Name: "flag"},
				},
				OnUsageError: func(c *Context, err error, _ bool) error {
					if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
						t.Errorf("Expect an invalid value error, but got \"%v\"", err)
					}
					return errors.New("intercepted: " + err.Error())
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

func TestCommand_OnUsageError_WithSubcommand(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "bar",
				Subcommands: []*Command{
					{
						Name: "baz",
					},
				},
				Flags: []Flag{
					&IntFlag{Name: "flag"},
				},
				OnUsageError: func(c *Context, err error, _ bool) error {
					if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
						t.Errorf("Expect an invalid value error, but got \"%v\"", err)
					}
					return errors.New("intercepted: " + err.Error())
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

func TestCommand_Run_SubcommandsCanUseErrWriter(t *testing.T) {
	app := &App{
		ErrWriter: io.Discard,
		Commands: []*Command{
			{
				Name:  "bar",
				Usage: "this is for testing",
				Subcommands: []*Command{
					{
						Name:  "baz",
						Usage: "this is for testing",
						Action: func(c *Context) error {
							if c.App.ErrWriter != io.Discard {
								return fmt.Errorf("ErrWriter not passed")
							}

							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar", "baz"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommandSkipFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs     args
		expectedArgs *args
		expectedErr  error
	}{
		{testArgs: args{"some-exec", "some-command", "some-arg", "--flag", "foo"}, expectedArgs: &args{"some-arg", "--flag", "foo"}, expectedErr: nil},
		{testArgs: args{"some-exec", "some-command", "some-arg", "--flag=foo"}, expectedArgs: &args{"some-arg", "--flag=foo"}, expectedErr: nil},
	}

	for _, c := range cases {
		var args Args
		app := &App{
			Commands: []*Command{
				{
					SkipFlagParsing: true,
					Name:            "some-command",
					Flags: []Flag{
						&StringFlag{Name: "flag"},
					},
					Action: func(c *Context) error {
						args = c.Args()
						return nil
					},
				},
			},
			Writer: io.Discard,
		}

		err := app.Run(c.testArgs)
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
		var outputBuffer bytes.Buffer
		app := &App{
			Writer:               &outputBuffer,
			EnableBashCompletion: true,
			Commands: []*Command{
				{
					Name:  "bar",
					Usage: "this is for testing",
					Flags: []Flag{
						&IntFlag{
							Name:  "number",
							Usage: "A number to parse",
						},
					},
					BashComplete: func(c *Context) {
						fmt.Fprintf(c.App.Writer, "found %d args", c.NArg())
					},
				},
			},
		}

		osArgs := args{"foo", "bar"}
		osArgs = append(osArgs, c.testArgs...)
		osArgs = append(osArgs, "--generate-bash-completion")

		err := app.Run(osArgs)
		stdout := outputBuffer.String()
		expect(t, err, nil)
		expect(t, stdout, c.expectedOut)
	}

}

func TestCommand_NoVersionFlagOnCommands(t *testing.T) {
	app := &App{
		Version: "some version",
		Commands: []*Command{
			{
				Name:        "bar",
				Usage:       "this is for testing",
				Subcommands: []*Command{{}}, // some subcommand
				HideHelp:    true,
				Action: func(c *Context) error {
					if len(c.Command.VisibleFlags()) != 0 {
						t.Fatal("unexpected flag on command")
					}
					return nil
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar"})
	expect(t, err, nil)
}

func TestCommand_CanAddVFlagOnCommands(t *testing.T) {
	app := &App{
		Version: "some version",
		Writer:  io.Discard,
		Commands: []*Command{
			{
				Name:        "bar",
				Usage:       "this is for testing",
				Subcommands: []*Command{{}}, // some subcommand
				Flags: []Flag{
					&BoolFlag{
						Name: "v",
					},
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar"})
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
	c := &Command{
		Name:  "bar",
		Usage: "this is for testing",
		Subcommands: []*Command{
			subc1,
			{
				Name:   "subc2",
				Usage:  "subc2 command2",
				Hidden: true,
			},
			subc3,
		},
	}

	expect(t, c.VisibleCommands(), []*Command{subc1, subc3})
}

func TestCommand_VisibleFlagCategories(t *testing.T) {

	c := &Command{
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
			&StringFlag{
				Name:     "sfd",
				Category: "cat2",
				Hidden:   true,
			},
		},
	}

	vfc := c.VisibleFlagCategories()
	if len(vfc) != 2 {
		t.Fatalf("unexpected visible flag categories %+v", vfc)
	}
	if vfc[1].Name() != "cat1" {
		t.Errorf("expected category name cat1 got %s", vfc[0].Name())
	}
	if len(vfc[1].Flags()) != 1 {
		t.Fatalf("expected flag category to have just one flag got %+v", vfc[0].Flags())
	}

	fl := vfc[1].Flags()[0]
	if !reflect.DeepEqual(fl.Names(), []string{"intd", "altd1", "altd2"}) {
		t.Errorf("unexpected flag %+v", fl.Names())
	}
}

func TestCommand_RunSubcommandWithDefault(t *testing.T) {
	app := &App{
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
				Name:        "bar",
				Usage:       "this is for testing",
				Subcommands: []*Command{{}}, // some subcommand
				Action: func(*Context) error {
					return nil
				},
			},
		},
	}

	err := app.Run([]string{"app", "bar"})
	expect(t, err, nil)

	err = app.Run([]string{"app"})
	expect(t, err, errors.New("should not run this subcommand"))
}

func TestCommand_PreservesSeparatorOnSubcommands(t *testing.T) {
	var values []string
	subcommand := &Command{
		Name: "bar",
		Flags: []Flag{
			&StringSliceFlag{Name: "my-flag"},
		},
		Action: func(c *Context) error {
			values = c.StringSlice("my-flag")
			return nil
		},
	}
	app := &App{
		Commands: []*Command{
			{
				Name:        "foo",
				Subcommands: []*Command{subcommand},
			},
		},
		SliceFlagSeparator: ";",
	}

	err := app.Run([]string{"app", "foo", "bar", "--my-flag", "1;2;3"})
	expect(t, err, nil)

	expect(t, values, []string{"1", "2", "3"})
}
