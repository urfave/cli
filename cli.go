package cli

import "fmt"
import "os"

// The name of the program. Defaults to os.Args[0]
var Name = os.Args[0]

// Description of the program.
var Usage = "<No Description>"

// Version of the program
var Version = "0.0.0"

// List of commands to execute
var Commands []Command

var Flags []Flag

// The action to execute when no subcommands are specified
var Action = ShowHelp

func Run(args []string) {

  set := flagSet(Flags)
  set.Parse(args[1:])

	context := Context{}
	if len(args) > 1 {
		name := args[1]
		for _, c := range append(Commands, HelpCommand) {
			if c.Name == name || c.ShortName == name {
				c.Action(context)
				return
			}
		}
	}

	// Run default Action
	Action(context)
}

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Handler
	Flags       []Flag
}

type Context struct {
}

type Handler func(context Context)
