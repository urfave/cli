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
		{[]string{"test-cmd", "blah", "blah", "-break"}, false, false, errors.New("flag provided but not defined: -break"), false},

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
		set.Parse(c.testArgs)

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
		args := []string{}
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
		value := ""
		args := []string{}
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
						value = c.String("flag")
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
