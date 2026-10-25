package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestCommandLine gives the test its own flag.CommandLine, where
// AllowExtFlags looks for Go flags.
func newTestCommandLine(t *testing.T) {
	orig := flag.CommandLine
	t.Cleanup(func() { flag.CommandLine = orig })
	flag.CommandLine = flag.NewFlagSet("foo", flag.ContinueOnError)
}

// traceLocation does not accept its own default text, like klog's
// -log_backtrace_at.
type traceLocation struct{}

func (*traceLocation) String() string { return ":0" }

func (*traceLocation) Set(string) error {
	return errors.New("syntax error: expect file.go:234")
}

func TestExtFlagDefaultNotAccepted(t *testing.T) {
	newTestCommandLine(t)
	flag.Var(&traceLocation{}, "log_backtrace_at", "")

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
}

func TestExtFlagResetOnRerun(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Int("ext", 2, "")

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext", "5"}))
	require.Equal(t, 5, *ext)

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
	require.Equal(t, 2, *ext)
}

func TestExtFlagBoolWithoutValue(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Bool("ext", false, "")

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext"}))
	require.True(t, *ext)
}

func TestExtFlagFromSubcommand(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Int("ext", 2, "")

	cmd := &Command{
		AllowExtFlags: true,
		Commands: []*Command{{
			Name:   "sub",
			Action: func(context.Context, *Command) error { return nil },
		}},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "sub", "--ext", "5"}))
	require.Equal(t, 5, *ext)
}

func TestExtFlagHelp(t *testing.T) {
	newTestCommandLine(t)
	flag.Int("ext", 2, "ext usage")
	flag.Bool("ext-bool", false, "ext bool usage")
	flag.String("ext-string", "", "ext string usage")
	flag.Bool("v", false, "ext v usage")

	var out bytes.Buffer
	cmd := &Command{
		AllowExtFlags: true,
		Writer:        &out,
		Flags: []Flag{
			&BoolFlag{Name: "verbose", Aliases: []string{"v"}, Category: "Main"},
		},
		Action: func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext-string", "bob", "--help"}))
	assert.Contains(t, out.String(), "--ext int            ext usage (default: 2)")
	assert.Contains(t, out.String(), "--ext-bool           ext bool usage (default: false)")
	assert.Contains(t, out.String(), "--ext-string string  ext string usage\n")
	assert.NotContains(t, out.String(), "ext v usage")
}

// An external flag set on the command line must report IsSet, NumFlags and
// LocalFlagNames like any other flag: Command.IsSet documents that it
// determines if the flag was actually set, and v2 reported external flags
// through the command's flag.FlagSet.Visit.
func TestExtFlagIsSet(t *testing.T) {
	newTestCommandLine(t)
	flag.Int("ext", 2, "")

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext", "5"}))
	require.True(t, cmd.IsSet("ext"))
	require.Equal(t, 1, cmd.NumFlags())
	require.Equal(t, []string{"ext"}, cmd.LocalFlagNames())
}
