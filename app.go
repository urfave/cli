package cli

import (
	"fmt"
	"io/ioutil"
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
	if a.Command(helpCommand.Name) == nil {
		a.Commands = append(a.Commands, helpCommand)
	}

	//append version/help flags
	a.appendFlag(BoolFlag{"version", "print the version"})
	a.appendFlag(helpFlag{"show help"})

	// parse flags
	set := flagSet(a.Name, a.Flags)
	set.SetOutput(ioutil.Discard)
	err := set.Parse(arguments[1:])
	context := NewContext(a, set, set)

	if err != nil {
		fmt.Println("Incorrect Usage.\n")
		ShowAppHelp(context)
		fmt.Println("")
		os.Exit(1)
	}

	checkHelp(context)
	checkVersion(context)

	args := context.Args()
	if len(args) > 0 {
		name := args[0]
		c := a.Command(name)
		if c != nil {
			c.Run(context)
			return
		}
	}

	// Run default Action
	a.Action(context)
}

func (a *App) Command(name string) *Command {
	for _, c := range a.Commands {
		if c.HasName(name) {
			return &c
		}
	}

	return nil
}

func (a *App) hasFlag(flag Flag) bool {
	for _, f := range a.Flags {
		if flag == f {
			return true
		}
	}

	return false
}

func (a *App) appendFlag(flag Flag) {
	if !a.hasFlag(flag) {
		a.Flags = append(a.Flags, flag)
	}
}
