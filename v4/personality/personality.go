// Package personality holds presets for [cli.Personality] beyond the core's
// POSIX default. Import only the ones you use.
package personality

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v4"
)

// Git behaves like git and other command dispatchers:
//
//	git: unknown command "foo"
//	See 'git --help'.
//
// An unknown command exits 1 and other usage errors exit 129, as git does.
var Git = &cli.Personality{
	Name:      "git",
	UsageCode: 129,
	Codes:     map[cli.Kind]int{cli.UnknownCommand: 1},
	Report: func(w io.Writer, cmd *cli.Command, err error, _ int) {
		for _, e := range cli.Split(err) {
			fmt.Fprintf(w, "%s: %s\n", cmd.Name, cmd.Localize(e))
			var ce *cli.Error
			if errors.As(e, &ce) && ce.Kind == cli.UnknownCommand && len(ce.Choices) > 0 {
				fmt.Fprintln(w, cmd.Text("hint.similar", ce.Args()))
			}
		}
		if cli.IsUsage(err) {
			fmt.Fprintln(w, cmd.Text("hint.see_help", cli.Args{Command: cmd.Name}))
		}
	},
}

// Agent writes one JSON object to stderr so that a program, such as a coding
// agent, can act on the failure without parsing prose:
//
//	{"exit_code":2,"errors":[{"kind":"too_many_args","message":"...","command":"hello","value":"extra","count":1}]}
//
// Exit codes match POSIX. Errors the library did not create have kind "error".
var Agent = &cli.Personality{
	Name:      "agent",
	UsageCode: 2,
	Report: func(w io.Writer, cmd *cli.Command, err error, code int) {
		out := envelope{ExitCode: code}
		for _, e := range cli.Split(err) {
			item := entry{Kind: "error", Message: cmd.Localize(e)}
			var ce *cli.Error
			if errors.As(e, &ce) {
				item.Kind = ce.Kind.String()
				item.Command = ce.Command
				item.Flag = ce.Flag
				item.Arg = ce.Arg
				item.Value = ce.Value
				item.Choices = ce.Choices
				item.Count = ce.Count
			}
			out.Errors = append(out.Errors, item)
		}
		b, _ := json.Marshal(out)
		fmt.Fprintf(w, "%s\n", b)
	},
}

type envelope struct {
	ExitCode int     `json:"exit_code"`
	Errors   []entry `json:"errors"`
}

type entry struct {
	Kind    string   `json:"kind"`
	Message string   `json:"message"`
	Command string   `json:"command,omitempty"`
	Flag    string   `json:"flag,omitempty"`
	Arg     string   `json:"arg,omitempty"`
	Value   string   `json:"value,omitempty"`
	Choices []string `json:"choices,omitempty"`
	Count   int      `json:"count,omitempty"`
}
