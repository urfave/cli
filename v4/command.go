package cli

import (
	"context"
	"errors"
	"io"
	"os"
)

// ActionFunc runs a command.
type ActionFunc func(ctx context.Context, cmd *Command) error

// Command is a program or one of its subcommands. Every setting that v3 kept
// in package variables lives here, so two commands never share state.
type Command struct {
	Name  string
	Usage string

	// MaxArgs is the most positional arguments the command accepts.
	// A negative value means no limit.
	MaxArgs int

	Action ActionFunc

	Writer    io.Writer // defaults to os.Stdout
	ErrWriter io.Writer // defaults to os.Stderr

	// Localizer renders messages. It defaults to English. Use
	// SelectLocalizer to pick one from the user's environment.
	Localizer Localizer

	// Personality decides how errors are reported and which exit codes
	// they map to. It defaults to POSIX.
	Personality *Personality

	args []string
}

// Run parses args, whose first element is the program name as in os.Args,
// and runs the action. It never exits the process; see [Main] and [Handle].
func (c *Command) Run(ctx context.Context, args []string) error {
	if len(args) > 0 {
		args = args[1:]
	}
	c.args = args

	if c.MaxArgs >= 0 && len(args) > c.MaxArgs {
		extra := args[c.MaxArgs:]
		return &Error{Kind: TooManyArgs, Command: c.Name, Value: extra[0], Count: len(extra)}
	}
	if c.Action == nil {
		return nil
	}
	return c.Action(ctx, c)
}

// Args returns the positional arguments from the last Run.
func (c *Command) Args() []string { return c.args }

// Out returns the writer for normal output.
func (c *Command) Out() io.Writer {
	if c.Writer != nil {
		return c.Writer
	}
	return os.Stdout
}

// Err returns the writer for diagnostics.
func (c *Command) Err() io.Writer {
	if c.ErrWriter != nil {
		return c.ErrWriter
	}
	return os.Stderr
}

func (c *Command) localizer() Localizer {
	if c.Localizer != nil {
		return c.Localizer
	}
	return English
}

// Text renders a catalogue message in the command's language.
func (c *Command) Text(key string, a Args) string {
	return c.localizer().Message(key, a)
}

// Localize renders err in the command's language. Errors the library did not
// create are returned as they are.
func (c *Command) Localize(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return c.Text(e.Kind.Key(), e.Args())
	}
	return err.Error()
}

func (c *Command) personality() *Personality {
	if c.Personality != nil {
		return c.Personality
	}
	return POSIX
}

// Handle reports err through the command's personality and returns the exit
// code for it. It returns 0 for a nil error. Help and version requests are
// not reported, since the help renderer has already written its output.
func Handle(cmd *Command, err error) int {
	p := cmd.personality()
	code := p.ExitCode(err)
	if err != nil && p.Report != nil && !errors.Is(err, ErrHelp) && !errors.Is(err, ErrVersion) {
		p.Report(cmd.Err(), cmd, err, code)
	}
	return code
}
