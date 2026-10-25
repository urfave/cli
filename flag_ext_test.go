package cli

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"flag"
	"strings"
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
type traceLocation struct{ loc string }

func (l *traceLocation) String() string { return cmp.Or(l.loc, ":0") }

func (l *traceLocation) Set(s string) error {
	if !strings.Contains(s, ".go:") {
		return errors.New("syntax error: expect file.go:234")
	}
	l.loc = s
	return nil
}

func TestExtFlagDefaultNotAccepted(t *testing.T) {
	newTestCommandLine(t)
	flag.Var(&traceLocation{}, "log_backtrace_at", "")

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))

	// a rerun can't restore ":0" either, and must not fail
	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--log_backtrace_at", "main.go:1"}))
	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
}

func TestExtFlagResetOnRerun(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Int("ext", 2, "")
	extString := flag.String("ext-string", "", "")

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext", "5", "--ext-string", "bob"}))
	require.Equal(t, 5, *ext)
	require.Equal(t, "bob", *extString)

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
	require.Equal(t, 2, *ext)
	require.Empty(t, *extString)
}

func TestExtFlagKeepsProgramValue(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Int("ext", 2, "")
	require.NoError(t, flag.Set("ext", "7"))

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
	require.Equal(t, 7, *ext)

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext", "5"}))
	require.Equal(t, 5, *ext)

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
	require.Equal(t, 7, *ext)
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

func TestExtFlagRootAndSubcommand(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Int("ext", 2, "")

	var got int
	var isSet bool
	cmd := &Command{
		AllowExtFlags: true,
		Commands: []*Command{{
			Name:          "sub",
			AllowExtFlags: true,
			Action: func(_ context.Context, cmd *Command) error {
				got = *ext
				isSet = cmd.IsSet("ext")
				return nil
			},
		}},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext", "5", "sub"}))
	require.Equal(t, 5, got)
	require.True(t, isSet)
}

func TestExtFlagParentFlagWins(t *testing.T) {
	for _, arg := range []string{"--verbose", "-v"} {
		t.Run(arg, func(t *testing.T) {
			newTestCommandLine(t)
			ext := flag.Bool("v", false, "")

			var verbose bool
			cmd := &Command{
				Flags: []Flag{
					&BoolFlag{Name: "verbose", Aliases: []string{"v"}},
				},
				Commands: []*Command{{
					Name:          "sub",
					AllowExtFlags: true,
					Action: func(_ context.Context, cmd *Command) error {
						verbose = cmd.Bool("verbose")
						return nil
					},
				}},
			}

			require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "sub", arg}))
			require.True(t, verbose)
			require.False(t, *ext)
		})
	}
}

func TestExtFlagMutuallyExclusiveFlagWins(t *testing.T) {
	newTestCommandLine(t)
	ext := flag.Bool("v", false, "")

	cmd := &Command{
		AllowExtFlags: true,
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{{
			Flags: [][]Flag{
				{&BoolFlag{Name: "verbose", Aliases: []string{"v"}}},
				{&BoolFlag{Name: "quiet"}},
			},
		}},
		Action: func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "-v"}))
	require.True(t, cmd.Bool("verbose"))
	require.False(t, *ext)
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

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo"}))
	require.False(t, cmd.IsSet("ext"))
}

func TestExtFlagWithoutGet(t *testing.T) {
	newTestCommandLine(t)
	var got string
	flag.Func("ext", "", func(s string) error {
		got = s
		return nil
	})

	cmd := &Command{
		AllowExtFlags: true,
		Action:        func(context.Context, *Command) error { return nil },
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--ext", "bob"}))
	require.Equal(t, "bob", got)
	require.Implements(t, (*flag.Value)(nil), cmd.Value("ext"))
	require.Empty(t, cmd.String("ext"))
}
