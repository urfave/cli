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
	// List of flags to parse
	Flags []Flag
	// The action to execute when no subcommands are specified
	Action func(context *Context)
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
	// append help to commands
	a.Commands = append(a.Commands, helpCommand)
	// append version to flags
	a.Flags = append(
		a.Flags,
		BoolFlag{"version", "print the version"},
		helpFlag{"show help"},
	)

	// parse flags
	set := flagSet(a.Name, a.Flags)
	err := set.Parse(arguments[1:])
	if err != nil {
		os.Exit(1)
	}

	context := NewContext(a, set, set)
	checkHelp(context)
	checkVersion(context)

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
