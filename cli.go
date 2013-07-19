package cli

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

func Run(arguments []string) {

	set := flagSet(Flags)
	set.Parse(arguments[1:])

	context := NewContext(set, set)
	args := context.Args()
	if len(args) > 0 {
		name := args[0]
		for _, c := range append(Commands, HelpCommand) {
			if c.Name == name || c.ShortName == name {
				c.Run(context)
				return
			}
		}
	}

	// Run default Action
	Action(context)
}

type Handler func(context *Context)
