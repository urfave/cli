//go:build go1.17
// +build go1.17

package cli

import (
	"bytes"
	"flag"
	"testing"
)

func TestApp_RunAsSubCommandIncorrectUsage(t *testing.T) {
	a := App{
		Flags: []Flag{
			StringFlag{Name: "--foo"},
		},
		Writer: bytes.NewBufferString(""),
	}

	set := flag.NewFlagSet("", flag.ContinueOnError)
	_ = set.Parse([]string{"", "---foo"})
	c := &Context{flagSet: set}

	// Go 1.17+ panics when invalid flag is given.
	// Catch it here and consider the test passed.
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("expected error, got nothing")
		}
	}()

	_ = a.RunAsSubcommand(c)
}
