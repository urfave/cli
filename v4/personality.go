package cli

import (
	"errors"
	"fmt"
	"io"
)

// Personality decides how a program behaves when something goes wrong: what
// it prints and which exit code it returns. The core ships POSIX; package
// personality has others.
type Personality struct {
	Name string

	// UsageCode is the exit code for usage errors.
	UsageCode int

	// Codes overrides the exit code for specific kinds.
	Codes map[Kind]int

	// Report writes err to w. code is the exit code the program will use.
	Report func(w io.Writer, cmd *Command, err error, code int)
}

// ExitCode returns the exit code for err. An [ExitCoder] anywhere in the
// chain chooses its own code. Usage errors exit UsageCode, and anything else
// exits 1.
func (p *Personality) ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ec ExitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	var e *Error
	if errors.As(err, &e) {
		if code, ok := p.Codes[e.Kind]; ok {
			return code
		}
		if e.Kind.Class() == Usage {
			return p.UsageCode
		}
	}
	return 1
}

// POSIX behaves like GNU and BSD tools such as ls and grep:
//
//	ls: unknown flag: z
//	Try 'ls --help' for more information.
//
// Usage errors exit 2.
var POSIX = &Personality{
	Name:      "posix",
	UsageCode: 2,
	Report: func(w io.Writer, cmd *Command, err error, _ int) {
		for _, e := range Split(err) {
			fmt.Fprintf(w, "%s: %s\n", cmd.Name, cmd.Localize(e))
		}
		if IsUsage(err) {
			fmt.Fprintln(w, cmd.Text("hint.try_help", Args{Command: cmd.Name}))
		}
	},
}
