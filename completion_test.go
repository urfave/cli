package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionDisable(t *testing.T) {
	cmd := &Command{}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName})
	assert.Error(t, err, "Expected error for no help topic for completion")
}

func TestCompletionEnable(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName})
	assert.ErrorContains(t, err, "no shell provided")
}

func TestCompletionEnableDiffCommandName(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "junky",
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "junky"})
	assert.ErrorContains(t, err, "no shell provided")
}

func TestCompletionShell(t *testing.T) {
	for k := range shellCompletions {
		out := &bytes.Buffer{}

		t.Run(k, func(t *testing.T) {
			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
			}

			r := require.New(t)

			r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, k}))
			r.Containsf(
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

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "bar", "--generate-shell-completion"}))
	r.Containsf(
		out.String(), "xyz",
		"Expected output to contain shell name %[1]q", "xyz",
	)

	out.Reset()

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "bar", "xyz", "-", "--generate-shell-completion"}))
	r.Containsf(
		out.String(), "g",
		"Expected output to contain shell name %[1]q", "g",
	)
}

func TestCompletionInvalidShell(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, "junky-sheell"})
	assert.ErrorContains(t, err, "unknown shell junky-sheell")
}
