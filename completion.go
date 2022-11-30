package cli

import (
	_ "embed"
	"fmt"
)

const (
	completionCommandName = "generate-completion"
)

//go:embed autocomplete/bash_autocomplete
var bash_autocomplete string

//go:embed autocomplete/powershell_autocomplete.ps1
var powershell_autocomplete string

//go:embed autocomplete/zsh_autocomplete
var zsh_autocomplete string

type renderCompletion func(a *App) (string, error)

func getCompletion(s string) renderCompletion {
	return func(a *App) (string, error) {
		return s, nil
	}
}

var shellCompletions = map[string]renderCompletion{
	"bash": getCompletion(bash_autocomplete),
	"ps":   getCompletion(powershell_autocomplete),
	"zsh":  getCompletion(zsh_autocomplete),
	"fish": func(a *App) (string, error) {
		return a.ToFishCompletion()
	},
}

var completionCommand = &Command{
	Name:   completionCommandName,
	Hidden: true,
	Action: func(ctx *Context) error {
		var shells []string
		for k := range shellCompletions {
			shells = append(shells, k)
		}

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
