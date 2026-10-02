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

// Modifiers change how a personality reports errors but keep its exit
// codes, so a script, a person and an agent all see the same code for the
// same failure.

// Agent returns a copy of base that writes one JSON object to stderr, so a
// program such as a coding agent can act on the failure without parsing
// prose:
//
//	{"version":1,"exit_code":2,"errors":[{"kind":"too_many_args","message":"...","command":"hello","value":"extra","count":1}]}
//
// Errors the library did not create have kind "error". The format is part of
// the public API; see [AgentFormatVersion].
func Agent(base *cli.Personality) *cli.Personality {
	p := *base
	p.Name = base.Name + "+agent"
	p.Report = reportJSON
	return &p
}

// Quiet returns a copy of base that prints one line per error and no hints,
// for CLIs that mostly run inside scripts.
func Quiet(base *cli.Personality) *cli.Personality {
	p := *base
	p.Name = base.Name + "+quiet"
	p.Report = func(w io.Writer, cmd *cli.Command, err error, _ int) {
		for _, e := range cli.Split(err) {
			fmt.Fprintf(w, "%s: %s\n", cmd.Name, cmd.Localize(e))
		}
	}
	return &p
}

// AgentFormatVersion is the version of the JSON that [Agent] writes. It
// changes only when a field is removed or changes meaning, so a program can
// reject a format it doesn't understand. New fields don't change it.
const AgentFormatVersion = 1

// AgentEnv is the environment variable that asks for [Agent] output. Agents,
// or the tools that run them, can set it once for every command they run.
const AgentEnv = "URFAVE_CLI_AGENT"

// Auto returns Agent(base) when AgentEnv is set to a true value such as
// "1", and base otherwise. lookup is usually [os.LookupEnv].
//
// Whether stderr is a terminal does not matter here: `prog 2>err.log` is
// still a person.
func Auto(lookup func(string) (string, bool), base *cli.Personality) *cli.Personality {
	v, _ := lookup(AgentEnv)
	switch v {
	case "1", "true", "TRUE", "True", "yes":
		return Agent(base)
	}
	return base
}

func reportJSON(w io.Writer, cmd *cli.Command, err error, code int) {
	out := envelope{Version: AgentFormatVersion, ExitCode: code}
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
}

type envelope struct {
	Version  int     `json:"version"`
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
