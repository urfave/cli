package cli

import (
	"bytes"
	"testing"

	itesting "github.com/urfave/cli/v3/internal/testing"
)

func TestCompletionDisable(t *testing.T) {
	cmd := &Command{}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName})
	itesting.Error(t, err, "Expected error for no help topic for completion")
}

func TestCompletionEnable(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName})
	itesting.ErrorContains(t, err, "no shell provided")
}

func TestCompletionEnableDiffCommandName(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "junky",
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "junky"})
	itesting.ErrorContains(t, err, "no shell provided")
}

func TestCompletionShell(t *testing.T) {
	for k := range shellCompletions {
		out := &bytes.Buffer{}

		t.Run(k, func(t *testing.T) {
			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
			}

			itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, k}))
			itesting.RequireContainsf(t,
				k, out.String(),
				"Expected output to contain shell name %[1]q", k,
			)
		})
	}
}

func TestCompletionSubcommand(t *testing.T) {
	out := &bytes.Buffer{}

	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Commands: []*Command{
			{
				Name: "bar",
				Commands: []*Command{
					{
						Name: "xyz",
						Flags: []Flag{
							&StringFlag{
								Name: "g",
								Aliases: []string{
									"t",
								},
							},
						},
					},
				},
			},
		},
	}

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar", "--generate-shell-completion"}))
	itesting.RequireContainsf(t,
		out.String(), "xyz",
		"Expected output to contain shell name %[1]q", "xyz",
	)

	out.Reset()

	itesting.RequireNoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar", "xyz", "-", "--generate-shell-completion"}))
	itesting.RequireContainsf(t,
		out.String(), "-g",
		"Expected output to contain flag %[1]q", "-g",
	)
}

func TestCompletionInvalidShell(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, "junky-sheell"})
	itesting.ErrorContains(t, err, "unknown shell junky-sheell")
}
