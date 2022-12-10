package cli

import (
	_ "embed"
	"fmt"
)

const (
	completionCommandName = "generate-completion"
)

var (
	//go:embed autocomplete
	autoCompleteFS embed.FS

	shellCompletions = map[string]renderCompletion{
		"bash": getCompletion("autocomplete/bash_autocomplete"),
		"ps":   getCompletion("autocomplete/powershell_autocomplete.ps1"),
		"zsh":  getCompletion("autocomplete/zsh_autocomplete"),
		"fish": func(a *App) (string, error) {
			return a.ToFishCompletion()
		},
	}
)

type renderCompletion func(a *App) (string, error)

func getCompletion(s string) renderCompletion {
	return func(a *App) (string, error) {
		b, err := autoCompleteFS.ReadFile(s)
		return string(b), err
	}
}

var completionCommand = &Command{
	Name:   completionCommandName,
	Hidden: true,
	Action: func(ctx *Context) error {
		var shells []string
		for k := range shellCompletions {
			shells = append(shells, k)
		}
		
		sort.Strings(shells)

		if ctx.Args().Len() == 0 {
			return Exit(fmt.Sprintf("no shell provided for completion command. available shells are %+v", shells), 1)
		}
		s := ctx.Args().First()

		if rc, ok := shellCompletions[s]; !ok {
			return Exit(fmt.Sprintf("unknown shell %s, available shells are %+v", s, shells), 1)
		} else if c, err := rc(ctx.App); err != nil {
			return Exit(err, 1)
		} else {
			ctx.App.Writer.Write([]byte(c))
		}
		return nil
	},
}
