package cli

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFishCompletion(t *testing.T) {
	// Given
	cmd := buildExtendedTestCommand()
	cmd.Flags = append(cmd.Flags,
		&StringFlag{
			Name:      "logfile",
			TakesFile: true,
		},
		&StringSliceFlag{
			Name:      "foofile",
			TakesFile: true,
		})
	cmd.setupCommandGraph()

	// When
	res, err := cmd.ToFishCompletion()

	// Then
	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}

func TestFishCompletionShortAndLongNames(t *testing.T) {
	// fish reads the argument of -s as a group of one-character options, so a
	// name that is longer than one character must be given with -l, and a
	// one-character name must be given with -s, wherever it appears in the
	// flag's list of names.
	cmd := &Command{
		Name: "greet",
		Flags: []Flag{
			&BoolFlag{Name: "verbose", Aliases: []string{"loud", "v"}},
			&StringFlag{Name: "o", Aliases: []string{"output", "out"}},
		},
	}
	cmd.setupCommandGraph()

	res, err := cmd.ToFishCompletion()
	require.NoError(t, err)

	assert.Contains(t, res, "complete -c greet -n '__fish_greet_no_subcommand' -f -l verbose -l loud -s v")
	assert.Contains(t, res, "complete -c greet -n '__fish_greet_no_subcommand' -f -s o -l output -l out -r")
	assert.NotContains(t, res, " -s loud")
	assert.NotContains(t, res, " -s output")
	assert.NotContains(t, res, " -l o ")
}

func TestFishCompletionBackslashEscaping(t *testing.T) {
	// Inside fish single-quoted strings the only escape sequences are \\ and
	// \', so a backslash in a description must be emitted as \\. An unescaped
	// backslash silently corrupts the description, and a trailing one turns the
	// closing quote into an escaped quote, leaving the string unterminated.
	// Ref: https://fishshell.com/docs/current/language.html
	cmd := &Command{
		Name: "greet",
		Flags: []Flag{
			&StringFlag{
				Name:  "path",
				Usage: `match \d+ then C:\tmp\`,
			},
		},
		Commands: []*Command{
			{
				Name:  "win",
				Usage: `run under C:\sys\`,
			},
		},
	}
	cmd.setupCommandGraph()

	res, err := cmd.ToFishCompletion()
	require.NoError(t, err)

	// Both the flag and the subcommand descriptions must have their backslashes
	// doubled in the generated `-d '...'` tokens.
	assert.Contains(t, res, `-d 'match \\d+ then C:\\tmp\\'`)
	assert.Contains(t, res, `-d 'run under C:\\sys\\'`)
}

func TestFishCompletionShellComplete(t *testing.T) {
	cmd := buildExtendedTestCommand()
	cmd.ShellComplete = func(context.Context, *Command) {}

	configCmd := cmd.Command("config")
	configCmd.ShellComplete = func(context.Context, *Command) {}

	subConfigCmd := configCmd.Command("sub-config")
	subConfigCmd.ShellComplete = func(context.Context, *Command) {}

	cmd.setupCommandGraph()

	res, err := cmd.ToFishCompletion()
	require.NoError(t, err)

	assert.Contains(t, res, fmt.Sprintf("complete -c greet -n '__fish_greet_no_subcommand' -xa '(greet %s 2>/dev/null)'", completionFlag))
	assert.Contains(t, res, fmt.Sprintf("complete -c greet -n '__fish_seen_subcommand_from config c' -xa '(greet config %s 2>/dev/null)'", completionFlag))
	assert.Contains(t, res, fmt.Sprintf("complete -c greet -n '__fish_seen_subcommand_from config c; and __fish_seen_subcommand_from sub-config s ss' -xa '(greet config sub-config %s 2>/dev/null)'", completionFlag))
}
