//go:build !go1.17
// +build !go1.17

package cli

import (
	"bytes"
	"errors"
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

	err := a.RunAsSubcommand(c)

	expect(t, err, errors.New("bad flag syntax: ---foo"))
}
