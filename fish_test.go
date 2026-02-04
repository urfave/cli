package cli

import (
	"context"
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

	oldTemplate := FishCompletionTemplate
	defer func() { FishCompletionTemplate = oldTemplate }()
	FishCompletionTemplate = "{{something"

	// test error case
	_, err1 := cmd.ToFishCompletion()
	assert.Error(t, err1)

	// reset the template
	FishCompletionTemplate = oldTemplate
	// When
	res, err := cmd.ToFishCompletion()

	// Then
	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
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

	assert.Contains(t, res, "complete -c greet -n '__fish_greet_no_subcommand' -xa '(greet --generate-shell-completion 2>/dev/null)'")
	assert.Contains(t, res, "complete -c greet -n '__fish_seen_subcommand_from config c' -xa '(greet config --generate-shell-completion 2>/dev/null)'")
	assert.Contains(t, res, "complete -c greet -n '__fish_seen_subcommand_from config c; and __fish_seen_subcommand_from sub-config s ss' -xa '(greet config sub-config --generate-shell-completion 2>/dev/null)'")
}
