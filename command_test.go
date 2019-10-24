package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCommandFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		skipArgReorder         bool
		expectedErr            error
		UseShortOptionHandling bool
	}{
		// Test normal "not ignoring flags" flow
		{[]string{"test-cmd", "blah", "blah", "-break"}, false, false, nil, false},

		// Test no arg reorder
		{[]string{"test-cmd", "blah", "blah", "-break"}, false, true, nil, false},
		{[]string{"test-cmd", "blah", "blah", "-break", "ls", "-l"}, false, true, nil, true},

		{[]string{"test-cmd", "blah", "blah"}, true, false, nil, false},   // Test SkipFlagParsing without any args that look like flags
		{[]string{"test-cmd", "blah", "-break"}, true, false, nil, false}, // Test SkipFlagParsing with random flag arg
		{[]string{"test-cmd", "blah", "-help"}, true, false, nil, false},  // Test SkipFlagParsing with "special" help flag arg
		{[]string{"test-cmd", "blah"}, false, false, nil, true},           // Test UseShortOptionHandling

	}

	for _, c := range cases {
		app := NewApp()
		app.Writer = ioutil.Discard
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := NewContext(app, set, nil)

		command := Command{
			Name:                   "test-cmd",
			Aliases:                []string{"tc"},
			Usage:                  "this is for testing",
			Description:            "testing",
			Action:                 func(_ *Context) error { return nil },
			SkipFlagParsing:        c.skipFlagParsing,
			SkipArgReorder:         c.skipArgReorder,
			UseShortOptionHandling: c.UseShortOptionHandling,
		}

		err := command.Run(context)

		expect(t, err, c.expectedErr)
		expect(t, []string(context.Args()), c.testArgs)
	}
}

func TestParseAndRunShortOpts(t *testing.T) {
	cases := []struct {
		testArgs     []string
		expectedErr  error
		expectedArgs []string
	}{
		{[]string{"foo", "test", "-a"}, nil, []string{}},
		{[]string{"foo", "test", "-c", "arg1", "arg2"}, nil, []string{"arg1", "arg2"}},
		{[]string{"foo", "test", "-f"}, nil, []string{}},
		{[]string{"foo", "test", "-ac", "--fgh"}, nil, []string{}},
		{[]string{"foo", "test", "-af"}, nil, []string{}},
		{[]string{"foo", "test", "-cf"}, nil, []string{}},
		{[]string{"foo", "test", "-acf"}, nil, []string{}},
		{[]string{"foo", "test", "--acf"}, errors.New("flag provided but not defined: -acf"), nil},
		{[]string{"foo", "test", "-invalid"}, errors.New("flag provided but not defined: -invalid"), nil},
		{[]string{"foo", "test", "-acf", "-invalid"}, errors.New("flag provided but not defined: -invalid"), nil},
		{[]string{"foo", "test", "--invalid"}, errors.New("flag provided but not defined: -invalid"), nil},
		{[]string{"foo", "test", "-acf", "--invalid"}, errors.New("flag provided but not defined: -invalid"), nil},
		{[]string{"foo", "test", "-acf", "arg1", "-invalid"}, nil, []string{"arg1", "-invalid"}},
		{[]string{"foo", "test", "-acf", "arg1", "--invalid"}, nil, []string{"arg1", "--invalid"}},
		{[]string{"foo", "test", "-acfi", "not-arg", "arg1", "-invalid"}, nil, []string{"arg1", "-invalid"}},
		{[]string{"foo", "test", "-i", "ivalue"}, nil, []string{}},
		{[]string{"foo", "test", "-i", "ivalue", "arg1"}, nil, []string{"arg1"}},
		{[]string{"foo", "test", "-i"}, errors.New("flag needs an argument: -i"), nil},
	}

	for _, c := range cases {
		var args []string
		cmd := Command{
			Name:        "test",
			Usage:       "this is for testing",
			Description: "testing",
			Action: func(c *Context) error {
				args = c.Args()
				return nil
			},
			SkipArgReorder:         true,
			UseShortOptionHandling: true,
			Flags: []Flag{
				BoolFlag{Name: "abc, a"},
				BoolFlag{Name: "cde, c"},
				BoolFlag{Name: "fgh, f"},
				StringFlag{Name: "ijk, i"},
			},
		}

		app := NewApp()
		app.Commands = []Command{cmd}

		err := app.Run(c.testArgs)

		expect(t, err, c.expectedErr)
		expect(t, args, c.expectedArgs)
	}
}

func TestCommand_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := NewApp()
	app.Commands = []Command{
		{
			Name: "bar",
			Before: func(c *Context) error {
				return fmt.Errorf("before error")
			},
			After: func(c *Context) error {
				return fmt.Errorf("after error")
			},
		},
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

	app := NewApp()
	app.Commands = []Command{
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
	app := NewApp()
	app.Commands = []Command{
		{
			Name: "bar",
			Flags: []Flag{
				IntFlag{Name: "flag"},
			},
			OnUsageError: func(c *Context, err error, _ bool) error {
				return fmt.Errorf("intercepted in %s: %s", c.Command.Name, err.Error())
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
	app := NewApp()
	app.Commands = []Command{
		{
			Name: "bar",
			Flags: []Flag{
				IntFlag{Name: "flag"},
			},
			OnUsageError: func(c *Context, err error, _ bool) error {
				if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
					t.Errorf("Expect an invalid value error, but got \"%v\"", err)
				}
				return errors.New("intercepted: " + err.Error())
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
	app := NewApp()
	app.Commands = []Command{
		{
			Name: "bar",
			Subcommands: []Command{
				{
					Name: "baz",
				},
			},
			Flags: []Flag{
				IntFlag{Name: "flag"},
			},
			OnUsageError: func(c *Context, err error, _ bool) error {
				if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
					t.Errorf("Expect an invalid value error, but got \"%v\"", err)
				}
				return errors.New("intercepted: " + err.Error())
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
	app := NewApp()
	app.ErrWriter = ioutil.Discard
	app.Commands = []Command{
		{
			Name:  "bar",
			Usage: "this is for testing",
			Subcommands: []Command{
				{
					Name:  "baz",
					Usage: "this is for testing",
					Action: func(c *Context) error {
						if c.App.ErrWriter != ioutil.Discard {
							return fmt.Errorf("ErrWriter not passed")
						}

						return nil
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

func TestCommandFlagReordering(t *testing.T) {
	cases := []struct {
		testArgs      []string
		expectedValue string
		expectedArgs  []string
		expectedErr   error
	}{
		{[]string{"some-exec", "some-command", "some-arg", "--flag", "foo"}, "foo", []string{"some-arg"}, nil},
		{[]string{"some-exec", "some-command", "some-arg", "--flag=foo"}, "foo", []string{"some-arg"}, nil},
		{[]string{"some-exec", "some-command", "--flag=foo", "some-arg"}, "foo", []string{"some-arg"}, nil},
	}

	for _, c := range cases {
		value := ""
		var args []string
		app := &App{
			Commands: []Command{
				{
					Name: "some-command",
					Flags: []Flag{
						StringFlag{Name: "flag"},
					},
					Action: func(c *Context) {
						fmt.Printf("%+v\n", c.String("flag"))
						value = c.String("flag")
						args = c.Args()
					},
				},
			},
		}

		err := app.Run(c.testArgs)
		expect(t, err, c.expectedErr)
		expect(t, value, c.expectedValue)
		expect(t, args, c.expectedArgs)
	}
}

func TestCommandSkipFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs     []string
		expectedArgs []string
		expectedErr  error
	}{
		{[]string{"some-exec", "some-command", "some-arg", "--flag", "foo"}, []string{"some-arg", "--flag", "foo"}, nil},
		{[]string{"some-exec", "some-command", "some-arg", "--flag=foo"}, []string{"some-arg", "--flag=foo"}, nil},
	}

	for _, c := range cases {
		var args []string
		app := &App{
			Commands: []Command{
				{
					SkipFlagParsing: true,
					Name:            "some-command",
					Flags: []Flag{
						StringFlag{Name: "flag"},
					},
					Action: func(c *Context) {
						fmt.Printf("%+v\n", c.String("flag"))
						args = c.Args()
					},
				},
			},
		}

		err := app.Run(c.testArgs)
		expect(t, err, c.expectedErr)
		expect(t, args, c.expectedArgs)
	}
}
