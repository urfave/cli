package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "short help flag",
			args: []string{"foo", completionCommandName, "-h"},
		},
		{
			name: "long help flag",
			args: []string{"foo", completionCommandName, "--help"},
		},
		{
			name: "completion bash short help flag",
			args: []string{"foo", completionCommandName, "bash", "-h"},
		},
		{
			name: "completion bash long help flag",
			args: []string{"foo", completionCommandName, "bash", "--help"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := &bytes.Buffer{}

			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
				Flags: []Flag{
					&StringFlag{
						Name:     "required-flag",
						Required: true,
					},
				},
			}

			r := require.New(t)

			r.NoError(cmd.Run(buildTestContext(t), test.args))
			r.Contains(out.String(), "USAGE")
			r.NotContains(out.String(), "GLOBAL OPTIONS")
		})
	}
}

func TestCompletionDisable(t *testing.T) {
	cmd := &Command{}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName})
	assert.Error(t, err, "Expected error for no help topic for completion")
}

func TestCompletionEnable(t *testing.T) {
	out := &bytes.Buffer{}

	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Flags: []Flag{
			&StringFlag{
				Name:     "goo",
				Required: true,
			},
		},
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandName}))
	r.Contains(out.String(), "USAGE")
}

func TestCompletionEnableDiffCommandName(t *testing.T) {
	out := &bytes.Buffer{}

	cmd := &Command{
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "junky",
		Writer:                     out,
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "junky"}))
	r.Contains(out.String(), "USAGE")
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
			r.NotEmpty(out.String(), "Expected non-empty completion output for shell %q", k)
		})
	}
}

func TestCompletionSubcommandOrder(t *testing.T) {
	// The completion subcommands must appear in a deterministic order so that
	// help output (and docs generated from it) does not change between runs.
	// Previously they were built by iterating a map, whose order Go randomizes.
	want := []string{"bash", "zsh", "fish", "pwsh"}

	// Build several times to guard against intra-process variation.
	for range 10 {
		cmd := buildCompletionCommand("foo")

		got := make([]string, 0, len(cmd.Commands))
		for _, sub := range cmd.Commands {
			got = append(got, sub.Name)
		}

		assert.Equal(t, want, got)
	}

	// Every shell in shellCompletions must be represented in the ordered list.
	assert.Len(t, completionShells, len(shellCompletions))
	for shell := range shellCompletions {
		assert.Contains(t, completionShells, shell)
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

func TestCompletionBashAppendsSpace(t *testing.T) {
	// Regression test for https://github.com/urfave/cli/issues/2332
	// Do not register bash completions with `-o nospace`: after a command or
	// subcommand completion, Bash should append a space so the next word can be
	// completed without manually typing one.

	cmd := &Command{
		EnableShellCompletion: true,
	}

	r := require.New(t)

	bashRender := shellCompletions["bash"]
	r.NotNil(bashRender, "bash completion renderer should exist")

	output, err := bashRender(cmd, "myapp")
	r.NoError(err)
	r.NotContains(output, "-o nospace", "bash completion should append spaces after completed words")
	r.Contains(output, "complete -o bashdefault -o default -F __myapp_bash_autocomplete myapp")
}

func TestCompletionBashGreedyColonParsing(t *testing.T) {
	// Regression test for https://github.com/urfave/cli/issues/2335
	// The bash completion template uses fmt.Sprintf, so
	// literal "%" in the template must be escaped as "%%". The token
	// extraction must use the greedy ${line%%:*} (double %%) to split on
	// the *first* colon. A single % would use ${line%:*} which splits on
	// the *last* colon, breaking descriptions that contain colons
	// (e.g. "export:Export configs such as: compose-config").

	cmd := &Command{
		EnableShellCompletion: true,
	}

	r := require.New(t)

	bashRender := shellCompletions["bash"]
	r.NotNil(bashRender, "bash completion renderer should exist")

	output, err := bashRender(cmd, "myapp")
	r.NoError(err)

	// After fmt.Sprintf, the rendered script must contain ${line%%:*}
	// (greedy match) not ${line%:*} (non-greedy match).
	r.Contains(output, `${line%%:*}`, "token extraction should use greedy %% to match first colon")
	r.NotContains(output, `${line%:*}`, "token extraction must not use non-greedy single % (splits on last colon)")
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

func TestCompletionFishSendsTokenBeingCompleted(t *testing.T) {
	// The word under the cursor is part of the request, quoted so that an empty one
	// is still an argument: without it, "cmd --<TAB>" and "cmd -- <TAB>" would reach
	// the command as the same request.
	cmd := &Command{
		Name:                  "myapp",
		EnableShellCompletion: true,
	}

	r := require.New(t)

	fishRender := shellCompletions["fish"]
	r.NotNil(fishRender, "fish completion renderer should exist")

	output, err := fishRender(cmd, "myapp")
	r.NoError(err)

	r.Contains(output, `set -l lastArg (string unescape -- (commandline -ct))`)
	r.Contains(output, `set results ($cmd __complete $args[2..-1] "$lastArg" 2> /dev/null)`)
	r.NotContains(output, completionFlag, "the deprecated request form must not be generated")
}

func TestCompletionBashSendsTokenBeingCompleted(t *testing.T) {
	// The word under the cursor is part of the request, empty or not: without it,
	// "cmd --<TAB>" and "cmd -- <TAB>" would reach the command as the same request.
	// The request is an array rather than a string to eval, so a word holding a space
	// or a quote reaches the command as the single word it is.
	cmd := &Command{
		Name:                  "myapp",
		EnableShellCompletion: true,
	}

	r := require.New(t)

	bashRender := shellCompletions["bash"]
	r.NotNil(bashRender, "bash completion renderer should exist")

	output, err := bashRender(cmd, "myapp")
	r.NoError(err)

	r.Contains(output, `__myapp_dequote "${words[0]}"`)
	r.Contains(output, `__myapp_completion_request=("${cmd}" "__complete")`)
	r.Contains(output, `__myapp_dequote "${words[cword]-}"`)
	r.Contains(output, `opts=$("${__myapp_completion_request[@]}" 2>/dev/null)`)
	r.Contains(output, `for (( i = 1; i < cword; i++ )); do`,
		"the request must come from the words __myapp_init_completion reassembled, not from COMP_WORDS")
	r.NotContains(output, `eval "`, "the request must not go through eval")
	r.NotContains(output, completionFlag, "the deprecated request form must not be generated")
}

func TestCompletionZshSendsTokenBeingCompleted(t *testing.T) {
	cmd := &Command{
		Name:                  "myapp",
		EnableShellCompletion: true,
	}

	r := require.New(t)

	zshRender := shellCompletions["zsh"]
	r.NotNil(zshRender, "zsh completion renderer should exist")

	output, err := zshRender(cmd, "myapp")
	r.NoError(err)

	r.Contains(output, `request=("$cmd" "__complete" "${(@Q)words[2,CURRENT-1]}" "$current")`)
	r.Contains(output, `opts=("${(@f)$("${request[@]}" 2>/dev/null)}")`,
		"a command writing to stderr must not break the prompt")
	r.NotContains(output, completionFlag, "the deprecated request form must not be generated")
}

func TestCompletionPowershellSendsTokenBeingCompleted(t *testing.T) {
	cmd := &Command{
		Name:                  "myapp",
		EnableShellCompletion: true,
	}

	r := require.New(t)

	pwshRender := shellCompletions["pwsh"]
	r.NotNil(pwshRender, "pwsh completion renderer should exist")

	output, err := pwshRender(cmd, "myapp")
	r.NoError(err)

	r.Contains(output, `& $command __complete @words $word 2>$null`)
	r.Contains(output, `if ($cursorPosition -gt $extent.StartOffset -and $cursorPosition -le $extent.EndOffset) {`,
		"the word being completed must come from the cursor, not from a comparison with $wordToComplete")
	r.NotContains(output, completionFlag, "the deprecated request form must not be generated")
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

func TestCompletionSubcommandCustomShellComplete(t *testing.T) {
	out := &bytes.Buffer{}

	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Commands: []*Command{
			{
				Name: "index",
				Commands: []*Command{
					{
						Name: "show",
						ShellComplete: func(ctx context.Context, cmd *Command) {
							fmt.Fprintln(cmd.Root().Writer, "custom-index")
						},
						Action: func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
		},
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "index", "show", completionFlag}))
	r.Equal("custom-index\n", out.String())
}

func TestCompletionRunsBeforeChain(t *testing.T) {
	type contextKey struct{}

	out := &bytes.Buffer{}
	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
			return context.WithValue(ctx, contextKey{}, "ready"), nil
		},
		Commands: []*Command{
			{
				Name: "index",
				Commands: []*Command{
					{
						Name: "show",
						ShellComplete: func(ctx context.Context, cmd *Command) {
							fmt.Fprintln(cmd.Root().Writer, ctx.Value(contextKey{}))
						},
						Action: func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
		},
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "index", "show", completionFlag}))
	r.Equal("ready\n", out.String())
}

func TestCompletionReturnsBeforeError(t *testing.T) {
	beforeErr := errors.New("load config")
	completed := false

	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                io.Discard,
		Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
			return nil, beforeErr
		},
		ShellComplete: func(ctx context.Context, cmd *Command) {
			completed = true
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionFlag})

	require.ErrorIs(t, err, beforeErr)
	assert.False(t, completed)
}

func TestCompletionInvalidShell(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	unknownShellName := "junky-sheell"
	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, unknownShellName})
	assert.ErrorContains(t, err, fmt.Sprintf("No help topic for '%s'", unknownShellName))
}

func TestCompletionShellRenderError(t *testing.T) {
	unknownShellName := "junky-sheell"

	enableError := true
	shellCompletions[unknownShellName] = func(c *Command, appName string) (string, error) {
		if enableError {
			return "", fmt.Errorf("can't do completion")
		}
		return "something", nil
	}
	// buildCompletionCommand only turns shells listed in completionShells into
	// subcommands, so register the injected shell there too (restoring the
	// original slice afterward) for it to be reachable.
	defer func(orig []string) { completionShells = orig }(completionShells)
	completionShells = append(completionShells, unknownShellName)
	defer func() {
		delete(shellCompletions, unknownShellName)
	}()

	cmd := &Command{
		EnableShellCompletion: true,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, unknownShellName})
	assert.ErrorContains(t, err, "can't do completion")
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

func TestCompletionShellWriteError(t *testing.T) {
	shellName := "mock-shell"
	shellCompletions[shellName] = func(c *Command, appName string) (string, error) {
		return "something", nil
	}
	// buildCompletionCommand only turns shells listed in completionShells into
	// subcommands, so register the injected shell there too (restoring the
	// original slice afterward) for it to be reachable.
	defer func(orig []string) { completionShells = orig }(completionShells)
	completionShells = append(completionShells, shellName)
	defer func() {
		delete(shellCompletions, shellName)
	}()

	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                &mockWriter{err: fmt.Errorf("writer error")},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", completionCommandName, shellName})
	assert.ErrorContains(t, err, "writer error")
}

// TestCompletionRequestNeverRunsAction is the regression test for
// https://github.com/urfave/cli/issues/1993: a shell asking for completions must
// never run the command, whatever the command line holds. A request appended to the
// end of the command line cannot promise that, because "--" turns it into a
// positional argument that a wrapper command is entitled to pass on, which is what
// https://github.com/urfave/cli/issues/1932 asked for.
func TestCompletionRequestNeverRunsAction(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{
			name: "plain",
			args: []string{"foo", completionCommandRequest, "exec", ""},
		},
		{
			name: "after a double dash",
			args: []string{"foo", completionCommandRequest, "exec", "--", "rm", "-rf"},
		},
		{
			name: "completing the double dash",
			args: []string{"foo", completionCommandRequest, "exec", "--"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ran := false
			out := &bytes.Buffer{}
			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
				Commands: []*Command{
					{
						Name:            "exec",
						SkipFlagParsing: true,
						Action: func(context.Context, *Command) error {
							ran = true
							return nil
						},
					},
				},
			}

			r := require.New(t)
			r.NoError(cmd.Run(buildTestContext(t), tc.args))
			r.False(ran, "the action must not run for a completion request")
		})
	}
}

// TestCompletionRequestAfterDoubleDash checks that the words after a "--" get no
// suggestion: they are positional arguments of whatever the command runs, so this
// command's flags and subcommands are no answer to them. The "--" being completed is
// not one of them.
func TestCompletionRequestAfterDoubleDash(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		expected string
	}{
		{
			// The completion is for the word after "exec", which has no subcommand of
			// its own to offer beyond the built-in help.
			name:     "before the double dash",
			args:     []string{"foo", completionCommandRequest, "exec", ""},
			expected: "help:Shows a list of commands or help for one command\n",
		},
		{
			name:     "the double dash itself",
			args:     []string{"foo", completionCommandRequest, "exec", "--"},
			expected: "--excitement\n--help:show help\n",
		},
		{
			name:     "after the double dash",
			args:     []string{"foo", completionCommandRequest, "exec", "--", "git", "pu"},
			expected: "",
		},
		{
			name:     "a flag after the double dash",
			args:     []string{"foo", completionCommandRequest, "exec", "--", "git", "--ver"},
			expected: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
				Commands: []*Command{
					{
						Name:   "exec",
						Flags:  []Flag{&BoolFlag{Name: "excitement"}},
						Action: func(context.Context, *Command) error { return nil },
					},
				},
			}

			r := require.New(t)
			r.NoError(cmd.Run(buildTestContext(t), tc.args))
			r.Equal(tc.expected, out.String())
		})
	}
}

// TestCompletionDeprecatedRequestPassedOnAfterDoubleDash is the regression test for
// https://github.com/urfave/cli/issues/1932: after a "--", the deprecated request
// form is a positional argument, so a wrapper command passes it on to whatever it
// runs instead of answering it. That command, run by the wrapper, is the one the
// shell was asking about.
func TestCompletionDeprecatedRequestPassedOnAfterDoubleDash(t *testing.T) {
	var got []string
	out := &bytes.Buffer{}
	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Commands: []*Command{
			{
				Name:            "exec",
				SkipFlagParsing: true,
				Action: func(_ context.Context, cmd *Command) error {
					got = cmd.Args().Slice()
					return nil
				},
			},
		},
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "exec", "--", "child", completionFlag}))
	r.Equal([]string{"--", "child", completionFlag}, got)
	r.Empty(out.String(), "the wrapper must not answer a request meant for what it runs")
}

// TestCompletionRequestKeepsArgsShape checks that a ShellComplete function sees the
// same cmd.Args() under both request forms: the word being completed is part of them
// when it starts with "-", and left out otherwise.
func TestCompletionRequestKeepsArgsShape(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "deprecated form completing a flag",
			args:     []string{"foo", "sub", "arg", "-", completionFlag},
			expected: "[arg -]\n",
		},
		{
			name:     "request completing a flag",
			args:     []string{"foo", completionCommandRequest, "sub", "arg", "-"},
			expected: "[arg -]\n",
		},
		{
			name:     "deprecated form completing a word",
			args:     []string{"foo", "sub", "arg", completionFlag},
			expected: "[arg]\n",
		},
		{
			name:     "request completing a word",
			args:     []string{"foo", completionCommandRequest, "sub", "arg", "wor"},
			expected: "[arg]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
				Commands: []*Command{
					{
						Name: "sub",
						ShellComplete: func(_ context.Context, cmd *Command) {
							fmt.Fprintf(cmd.Root().Writer, "%v\n", cmd.Args().Slice())
						},
						Action: func(context.Context, *Command) error { return nil },
					},
				},
			}

			r := require.New(t)
			r.NoError(cmd.Run(buildTestContext(t), tc.args))
			r.Equal(tc.expected, out.String())
		})
	}
}

// TestCompletionRequestStateIsPerRun checks that a Command answering several requests
// carries no state from one into the next. A shell runs one request per process, but
// a test, a REPL or an embedded use answers several through the same Command.
func TestCompletionRequestStateIsPerRun(t *testing.T) {
	out := &bytes.Buffer{}
	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Commands: []*Command{
			{
				Name:   "exec",
				Flags:  []Flag{&BoolFlag{Name: "excitement"}},
				Action: func(context.Context, *Command) error { return nil },
			},
		},
	}

	r := require.New(t)

	// A request past a "--" gets no suggestion, and records that.
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "--", "git", "pu"}))
	r.Empty(out.String())

	// The next request is a different one, and is answered on its own terms.
	out.Reset()
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "-"}))
	r.Equal("--excitement\n--help:show help\n", out.String())

	// The same holds for a request in the deprecated form, which says nothing about
	// the word being completed and so must not read what an earlier one said.
	out.Reset()
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "--", "git", "pu"}))
	r.Empty(out.String())
	out.Reset()
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", "exec", "-", completionFlag}))
	r.Equal("--excitement\n--help:show help\n", out.String())
}

// TestCompletionCustomShellCompleteNotRunPastDoubleDash checks that a command
// carrying a ShellComplete of its own suggests nothing past a "--" without having to
// know about "--": the words there are positional arguments of whatever it runs.
func TestCompletionCustomShellCompleteNotRunPastDoubleDash(t *testing.T) {
	ran := false
	out := &bytes.Buffer{}
	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Commands: []*Command{
			{
				Name: "exec",
				ShellComplete: func(_ context.Context, cmd *Command) {
					ran = true
					fmt.Fprintln(cmd.Root().Writer, "custom")
				},
				Action: func(context.Context, *Command) error { return nil },
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "--", "git", "pu"}))
	r.False(ran, "the completion func must not run past a double dash")
	r.Empty(out.String())

	// It is the "--" that stops it, not the command.
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "pu"}))
	r.True(ran)
	r.Equal("custom\n", out.String())
}

// TestCompletionRequestIgnoredWhenDisabled checks that the request form means nothing
// to an app that has not enabled shell completion: the first argument reaches it as
// the positional argument it wrote.
func TestCompletionRequestIgnoredWhenDisabled(t *testing.T) {
	var got []string
	out := &bytes.Buffer{}
	cmd := &Command{
		Writer: out,
		Action: func(_ context.Context, cmd *Command) error {
			got = cmd.Args().Slice()
			return nil
		},
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "bar"}))
	r.Equal([]string{completionCommandRequest, "bar"}, got)
	r.Empty(out.String())
}

// TestCompletionRequestNestedSubcommand checks that a request is answered by the
// command it names however deep that is, rather than by the one above it.
func TestCompletionRequestNestedSubcommand(t *testing.T) {
	out := &bytes.Buffer{}
	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Commands: []*Command{
			{
				Name: "one",
				Commands: []*Command{
					{
						Name:   "two",
						Flags:  []Flag{&BoolFlag{Name: "deep"}},
						Action: func(context.Context, *Command) error { return nil },
					},
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "one", ""}))
	r.Equal("two\nhelp:Shows a list of commands or help for one command\n", out.String())

	out.Reset()
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "one", "two", "-"}))
	r.Equal("--deep\n--help:show help\n", out.String())

	// The word being completed is a flag of the command it follows, even with a
	// positional argument in between, which the arguments alone could not say.
	out.Reset()
	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "one", "two", "arg", "--de"}))
	r.Equal("--deep\n", out.String())
}

// TestCompletionBeforeNotRunPastDoubleDash checks that a Before is not run for a
// request past a "--": there is no completion to prepare for, since the words there
// belong to whatever the command runs, and a Before with a side effect would fire on
// every tab key for an answer that is always empty.
func TestCompletionBeforeNotRunPastDoubleDash(t *testing.T) {
	ran := 0
	out := &bytes.Buffer{}
	cmd := &Command{
		EnableShellCompletion: true,
		Writer:                out,
		Before: func(ctx context.Context, _ *Command) (context.Context, error) {
			ran++
			return ctx, nil
		},
		Commands: []*Command{
			{
				Name:   "exec",
				Action: func(context.Context, *Command) error { return nil },
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "--", "git", "pu"}))
	r.Zero(ran, "Before must not run for a request past a double dash")
	r.Empty(out.String())

	r.NoError(cmd.Run(buildTestContext(t), []string{"foo", completionCommandRequest, "exec", "pu"}))
	r.Equal(1, ran, "Before still runs for a request the command answers")
}
