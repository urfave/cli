package cli

import (
	"context"
	"embed"
	"fmt"
	"strings"
)

const (
	completionCommandName = "completion"

	// This flag is supposed to only be used by the completion script itself to generate completions on the fly.
	completionFlag = "--generate-shell-completion"
)

type renderCompletion func(cmd *Command, appName string) (string, error)

var (
	//go:embed autocomplete
	autoCompleteFS embed.FS

	shellCompletions = map[string]renderCompletion{
		"bash": func(c *Command, appName string) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/bash_autocomplete")
			return fmt.Sprintf(string(b), appName), err
		},
		"zsh": func(c *Command, appName string) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/zsh_autocomplete")
			return fmt.Sprintf(string(b), appName), err
		},
		"fish": func(c *Command, appName string) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/fish_autocomplete")
			return fmt.Sprintf(string(b), appName), err
		},
		"pwsh": func(c *Command, appName string) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/powershell_autocomplete.ps1")
			return string(b), err
		},
	}
)

const completionDescription = `Output shell completion script for bash, zsh, fish, or Powershell.
Source the output to enable completion.

# .bashrc
source <($COMMAND completion bash)

# .zshrc
source <($COMMAND completion zsh)

# fish
$COMMAND completion fish > ~/.config/fish/completions/$COMMAND.fish

# Powershell
Output the script to path/to/autocomplete/$COMMAND.ps1 an run it.
`

func buildCompletionCommand(appName string) *Command {
	cmd := &Command{
		Name:                completionCommandName,
		Hidden:              true,
		Usage:               "Output shell completion script for bash, zsh, fish, or Powershell",
		Description:         strings.ReplaceAll(completionDescription, "$COMMAND", appName),
		isCompletionCommand: true,
	}

	for shell, render := range shellCompletions {
		cmd.Commands = append(cmd.Commands, buildShellCompletionSubcommand(shell, render, appName))
	}

	return cmd
}

func buildShellCompletionSubcommand(shell string, render renderCompletion, appName string) *Command {
	return &Command{
		Name:                shell,
		Usage:               fmt.Sprintf("Output %s completion script", shell),
		isCompletionCommand: true,
		Action: func(ctx context.Context, cmd *Command) error {
			completionScript, err := render(cmd, appName)
			if err != nil {
				return Exit(err, 1)
			}
			_, err = cmd.Writer.Write([]byte(completionScript))
			if err != nil {
				return Exit(err, 1)
			}
			return nil
		},
	}
}
