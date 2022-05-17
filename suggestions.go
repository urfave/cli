package cli

import (
	"fmt"

	"github.com/antzucaro/matchr"
)

const didYouMeanTemplate = "Did you mean '%s'?"

func (a *App) suggestFlagFromError(err error, command string) (string, error) {
	flag, parseErr := flagFromError(err)
	if parseErr != nil {
		return "", err
	}

	flags := a.Flags
	if command != "" {
		cmd := a.Command(command)
		if cmd == nil {
			return "", err
		}
		flags = cmd.Flags
	}

	suggestion := a.suggestFlag(flags, flag)
	if len(suggestion) == 0 {
		return "", err
	}

	return fmt.Sprintf(didYouMeanTemplate+"\n\n", suggestion), nil
}

func (a *App) suggestFlag(flags []Flag, provided string) (suggestion string) {
	distance := 0.0

	for _, flag := range flags {
		flagNames := flag.Names()
		if !a.HideHelp {
			flagNames = append(flagNames, HelpFlag.Names()...)
		}
		for _, name := range flagNames {
			newDistance := matchr.JaroWinkler(name, provided, true)
			if newDistance > distance {
				distance = newDistance
				suggestion = name
			}
		}
	}

	if len(suggestion) == 1 {
		suggestion = "-" + suggestion
	} else if len(suggestion) > 1 {
		suggestion = "--" + suggestion
	}

	return suggestion
}

// suggestCommand takes a list of commands and a provided string to suggest a
// command name
func suggestCommand(commands []*Command, provided string) (suggestion string) {
	distance := 0.0
	for _, command := range commands {
		for _, name := range append(command.Names(), helpName, helpAlias) {
			newDistance := matchr.JaroWinkler(name, provided, true)
			if newDistance > distance {
				distance = newDistance
				suggestion = name
			}
		}
	}

	return fmt.Sprintf(didYouMeanTemplate, suggestion)
}
