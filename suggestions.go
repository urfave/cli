package cli

import (
	"fmt"

	"github.com/antzucaro/matchr"
)

func suggestFlag(flags []Flag, provided string, hideHelp bool) string {
	distance := 0.0
	suggestion := ""

	for _, flag := range flags {
		flagNames := flag.Names()
		if !hideHelp {
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

	return fmt.Sprintf(SuggestDidYouMeanTemplate, suggestion)
}
