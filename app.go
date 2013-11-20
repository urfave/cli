package cli

import (
	"fmt"
	"io/ioutil"
	"os"
)

// App is the main structure of a cli application. It is recomended that
// and app be created with the cli.NewApp() function
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

// Creates a new cli Application with some reasonable defaults for Name, Usage, Version and Action.
func NewApp() *App {
	return &App{
		Name:    os.Args[0],
		Usage:   "A new cli application",
		Version: "0.0.0",
		Action:  helpCommand.Action,
	}
}

// Entry point to the cli app. Parses the arguments slice and routes to the proper flag/args combination
func (a *App) Run(arguments []string) error {
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
	nerr := normalizeFlags(a.Flags, set)
	if nerr != nil {
		fmt.Println(nerr)
		context := NewContext(a, set, set)
		ShowAppHelp(context)
		fmt.Println("")
		return nerr
	}
	context := NewContext(a, set, set)

	if err != nil {
		fmt.Println("Incorrect Usage.\n")
		fmt.Println("")
		ShowAppHelp(context)
		fmt.Println("")
		return err
	}

	if checkHelp(context) {
		return nil
	}

	if checkVersion(context) {
		return nil
	}

	args := context.Args()
	if len(args) > 0 {
		name := args[0]
		c := a.Command(name)
		if c != nil {
			return c.Run(context)
		}
	}

	// Run default Action
	a.Action(context)
	return nil
}

// Returns the named command on App. Returns nil if the command does not exist
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
