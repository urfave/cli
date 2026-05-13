package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
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
		Flags: []Flag{
			&StringFlag{
				Name:     "goo",
				Required: true,
			},
		},
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

func TestCompletionBashNoShebang(t *testing.T) {
	// Regression test for https://github.com/urfave/cli/issues/2259
	// Bash completion scripts are sourced, not executed, so they must not
	// start with a `#!` shebang (flagged by Debian lintian as
	// `bash-completion-with-hashbang`).

	cmd := &Command{
		EnableShellCompletion: true,
	}

	r := require.New(t)

	bashRender := shellCompletions["bash"]
	r.NotNil(bashRender, "bash completion renderer should exist")

	output, err := bashRender(cmd, "myapp")
	r.NoError(err)
	r.NotEmpty(output, "bash completion output should not be empty")
	r.False(strings.HasPrefix(output, "#!"), "bash completion should not start with a shebang")
}

func TestCompletionFishFormat(t *testing.T) {
	// Regression test for https://github.com/urfave/cli/issues/2285
	// Fish completion was broken due to incorrect format specifiers

	cmd := &Command{
		Name:                  "myapp",
		EnableShellCompletion: true,
	}

	r := require.New(t)

	// Test the fish shell completion renderer directly
	fishRender := shellCompletions["fish"]
	r.NotNil(fishRender, "fish completion renderer should exist")

	output, err := fishRender(cmd, "myapp")
	r.NoError(err)

	// Verify the function name is correctly formatted
	r.Contains(output, "function __myapp_perform_completion", "function name should contain app name")

	// Verify no format errors (like %! or (string=) which indicate broken fmt.Sprintf)
	r.NotContains(output, "%!", "output should not contain format errors")
	r.NotContains(output, "(string=", "output should not contain invalid fish syntax")

	// Verify the complete commands reference the app correctly
	r.Contains(output, "complete -c myapp", "complete command should reference app name")
	r.Contains(output, "(__myapp_perform_completion)", "completion function should be registered")
}

func TestCompletionSubcommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		contains    string
		msg         string
		msgArgs     []any
		notContains bool
	}{
		{
			name:     "subcommand general completion",
			args:     []string{"foo", "bar", completionFlag},
			contains: "xyz",
			msg:      "Expected output to contain shell name %[1]q",
			msgArgs: []any{
				"xyz",
			},
		},
		{
			name:     "subcommand flag completion",
			args:     []string{"foo", "bar", "-", completionFlag},
			contains: "l1",
			msg:      "Expected output to contain shell name %[1]q",
			msgArgs: []any{
				"l1",
			},
		},
		{
			name:     "subcommand double dash shows long flags",
			args:     []string{"foo", "bar", "--", completionFlag},
			contains: "--l1",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"--l1",
			},
		},
		{
			name:     "sub sub command general completion",
			args:     []string{"foo", "bar", "xyz", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"-g",
			},
			notContains: true,
		},
		{
			name:     "sub sub command flag completion",
			args:     []string{"foo", "bar", "xyz", "-", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"-g",
			},
		},
		{
			name:     "sub sub command double dash shows flags",
			args:     []string{"foo", "bar", "xyz", "--", completionFlag},
			contains: "--help",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"--help",
			},
		},
		{
			name:     "sub sub command no completion extra args",
			args:     []string{"foo", "bar", "xyz", "--", "sargs", completionFlag},
			contains: "-g",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"-g",
			},
			notContains: true,
		},
		{
			name:     "subcommand partial double dash flag completion",
			args:     []string{"foo", "bar", "--l", completionFlag},
			contains: "--l1",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"--l1",
			},
		},
		{
			name:     "sub sub command partial double dash flag completion",
			args:     []string{"foo", "bar", "xyz", "--he", completionFlag},
			contains: "--help",
			msg:      "Expected output to contain flag %[1]q",
			msgArgs: []any{
				"--help",
			},
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
