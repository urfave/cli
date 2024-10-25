package cli

import (
	"context"
	"embed"
	"fmt"
	"sort"
)

const (
	completionCommandName = "generate-completion"
	completionFlagName    = "generate-shell-completion"
	completionFlag        = "--" + completionFlagName
)

var (
	//go:embed autocomplete
	autoCompleteFS embed.FS

	shellCompletions = map[string]renderCompletion{
		"bash": getCompletion("autocomplete/bash_autocomplete"),
		"ps":   getCompletion("autocomplete/powershell_autocomplete.ps1"),
		"zsh":  getCompletion("autocomplete/zsh_autocomplete"),
		"fish": func(c *Command) (string, error) {
			return c.ToFishCompletion()
		},
	}
)

type renderCompletion func(*Command) (string, error)

func getCompletion(s string) renderCompletion {
	return func(c *Command) (string, error) {
		b, err := autoCompleteFS.ReadFile(s)
		return string(b), err
	}
}

func buildCompletionCommand() *Command {
	return &Command{
		Name:   completionCommandName,
		Hidden: true,
		Action: completionCommandAction,
	}
}

func completionCommandAction(ctx context.Context, cmd *Command) error {
	var shells []string
	for k := range shellCompletions {
		shells = append(shells, k)
	}

	sort.Strings(shells)

	if cmd.Args().Len() == 0 {
		return Exit(fmt.Sprintf("no shell provided for completion command. available shells are %+v", shells), 1)
	}
	s := cmd.Args().First()

	if rc, ok := shellCompletions[s]; !ok {
		return Exit(fmt.Sprintf("unknown shell %s, available shells are %+v", s, shells), 1)
	} else if c, err := rc(cmd); err != nil {
		return Exit(err, 1)
	} else {
		if _, err = cmd.Writer.Write([]byte(c)); err != nil {
			return Exit(err, 1)
		}
	}
	return nil
}
