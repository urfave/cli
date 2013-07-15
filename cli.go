package cli

import "os"

// The name of the program. Defaults to os.Args[0]
var Name = os.Args[0]

// Description of the program.
var Usage = "<No Description>"

// Version of the program
var Version = "0.0.0"

// List of commands to execute
var Commands []Command = nil

var DefaultAction = ShowHelp

func Run(args []string) {
	if len(args) > 1 {
		name := args[1]
		for _, c := range append(Commands, HelpCommand) {
			if c.Name == name || c.ShortName == name {
				c.Action(name)
				return
			}
		}
	}

	// Run default Action
	DefaultAction("")
}

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Action
}

type Action func(name string)
