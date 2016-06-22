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
		testArgs        []string
		skipFlagParsing bool
		expectedErr     error
	}{
		{[]string{"test-cmd", "-break", "blah", "blah"}, false, errors.New("flag provided but not defined: -break")}, // Test normal "not ignoring flags" flow
		{[]string{"test-cmd", "blah", "blah"}, true, nil},                                                            // Test SkipFlagParsing without any args that look like flags
		{[]string{"test-cmd", "-break", "blah"}, true, nil},                                                          // Test SkipFlagParsing with random flag arg
		{[]string{"test-cmd", "-help", "blah"}, true, nil},                                                           // Test SkipFlagParsing with "special" help flag arg
	}

	for _, c := range cases {
		app := &App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		set.Parse(c.testArgs)

		context := NewContext(app, set, nil)

		command := Command{
			Name:        "test-cmd",
			Aliases:     []string{"tc"},
			Usage:       "this is for testing",
			Description: "testing",
			Action:      func(_ *Context) error { return nil },
		}

		command.SkipFlagParsing = c.skipFlagParsing

		err := command.Run(context)

		expect(t, err, c.expectedErr)
		expect(t, context.Args().Slice(), c.testArgs)
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
