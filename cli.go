package cli

import "os"

// The name of the program. Defaults to os.Args[0]
var Name = os.Args[0]

// Description of the program.
var Usage = ""

// Version of the program
var Version = "0.0.0"

// List of commands to execute
var Commands []Command = nil

var DefaultAction = ShowHelp

func Run(args []string) {
  if len(args) > 1 {
    command := args[1]
    commands := CommandsWithDefaults()
    for _, c := range commands {
      if c.Name == command {
        c.Action(command)
        return
      }
    }
  }

	// Run default Action
  DefaultAction("")
}

func CommandsWithDefaults() []Command {
	return append(append([]Command(nil), HelpCommand), Commands...)
}

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Action
}

type Action func(name string)
