package cli

import (
	"bytes"
	"context"
	"fmt"
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
	tests := []struct {
		name        string
		args        []string
		contains    string
		msg         string
		msgArgs     []interface{}
		notContains bool
	}{
		{
			name:     "subcommand general completion",
			args:     []string{"foo", "bar", completionFlag},
			contains: "xyz",
			msg:      "Expected output to contain shell name %[1]q",
			msgArgs: []interface{}{
				"xyz",
			},
		},
		{
			name:     "subcommand flag completion",
			args:     []string{"foo", "bar", "-", completionFlag},
			contains: "l1",
			msg:      "Expected output to contain shell name %[1]q",
			msgArgs: []interface{}{
				"l1",
			},
		},
		{
			name:     "subcommand flag no completion",
			args:     []string{"foo", "bar", "--", completionFlag},
			contains: "l1",
			msg:      "Expected output to contain shell name %[1]q",
			msgArgs: []interface{}{
				"l1",
			},
			notContains: true,
		},
		{
			name:     "sub sub command general completion",
			args:     []string{"foo", "bar", "xyz", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []interface{}{
				"-g",
			},
			notContains: true,
		},
		{
			name:     "sub sub command flag completion",
			args:     []string{"foo", "bar", "xyz", "-", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []interface{}{
				"-g",
			},
		},
		{
			name:     "sub sub command no completion",
			args:     []string{"foo", "bar", "xyz", "--", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []interface{}{
				"-g",
			},
			notContains: true,
		},
		{
			name:     "sub sub command no completion extra args",
			args:     []string{"foo", "bar", "xyz", "--", "sargs", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []interface{}{
				"-g",
			},
			notContains: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := &bytes.Buffer{}

			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
				Commands: []*Command{
					{
						Name: "bar",
						Flags: []Flag{
							&StringFlag{
								Name: "l1",
							},
						},
						Action: func(ctx context.Context, c *Command) error { return nil },
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
								Action: func(ctx context.Context, c *Command) error { return nil },
							},
						},
					},
				},
			}

			r := require.New(t)

			r.NoError(cmd.Run(buildTestContext(t), test.args))
			if test.notContains {
				r.NotContainsf(out.String(), test.contains, test.msg, test.msgArgs...)
			} else {
				r.Containsf(out.String(), test.contains, test.msg, test.msgArgs...)
			}
		})
	}
}

type mockWriter struct {
	err error
}

func (mw *mockWriter) Write(p []byte) (int, error) {
	if mw.err != nil {
		return 0, mw.err
	}
	return len(p), nil
}

func TestCompletionInvalidShell(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	unknownShellName := "junky-sheell"
	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, unknownShellName})
	assert.ErrorContains(t, err, "unknown shell junky-sheell")

	enableError := true
	shellCompletions[unknownShellName] = func(c *Command, appName string) (string, error) {
		if enableError {
			return "", fmt.Errorf("cant do completion")
		}
		return "something", nil
	}
	defer func() {
		delete(shellCompletions, unknownShellName)
	}()

	err = cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, unknownShellName})
	assert.ErrorContains(t, err, "cant do completion")

	// now disable shell completion error
	enableError = false
	c := cmd.Command(completionCommandName)
	assert.NotNil(t, c)
	c.Writer = &mockWriter{
		err: fmt.Errorf("writer error"),
	}
	err = cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, unknownShellName})
	assert.ErrorContains(t, err, "writer error")
}
