package cli

import (
	"errors"
	"flag"
	"io/ioutil"
	"testing"
)

func TestCommandFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs        []string
		skipFlagParsing bool
		expectedErr     error
	}{
		{[]string{"blah", "blah", "-break"}, false, errors.New("flag provided but not defined: -break")}, // Test normal "not ignoring flags" flow
		{[]string{"blah", "blah"}, true, nil},                                                            // Test SkipFlagParsing without any args that look like flags
		{[]string{"blah", "-break"}, true, nil},                                                          // Test SkipFlagParsing with random flag arg
		{[]string{"blah", "-help"}, true, nil},                                                           // Test SkipFlagParsing with "special" help flag arg
	}

	for _, c := range cases {
		app := NewApp()
		app.Writer = ioutil.Discard
		set := flag.NewFlagSet("test", 0)
		set.Parse(c.testArgs)

		context := NewContext(app, set, nil)

		command := Command{
			Name:        "test-cmd",
			Aliases:     []string{"tc"},
			Usage:       "this is for testing",
			Description: "testing",
			Action:      func(_ *Context) {},
		}

		command.SkipFlagParsing = c.skipFlagParsing

		err := command.Run(context)

		expect(t, err, c.expectedErr)
		expect(t, []string(context.Args()), c.testArgs)
	}
}
