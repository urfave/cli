package cli

import (
	"os"
)

type App struct {
	// The name of the program. Defaults to os.Args[0]
	Name string
	// Description of the program.
	Usage string
	// Version of the program
	Version string
	// List of commands to execute
	Commands []Command
	Flags    []Flag
	// The action to execute when no subcommands are specified
	Action Handler
}

func NewApp() *App {
	return &App{
		Name:    os.Args[0],
		Usage:   "A new cli application",
		Version: "0.0.0",
		Action:  helpCommand.Action,
	}
}

func (a *App) Run(arguments []string) {
	// parse flags
	set := flagSet(a.Flags)
	set.Parse(arguments[1:])

	// append help to commands
	a.Commands = append(a.Commands, helpCommand)

	context := NewContext(a, set, set)
	args := context.Args()
	if len(args) > 0 {
		name := args[0]
		for _, c := range a.Commands {
			if c.HasName(name) {
				c.Run(context)
				return
			}
		}
	}

	// Run default Action
	a.Action(context)
}
