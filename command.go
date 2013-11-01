package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Command is a subcommand for a cli.App.
type Command struct {
	// The name of the command
	Name string
	// short name of the command. Typically one character
	ShortName string
	// A short description of the usage of this command
	Usage string
	// A longer explaination of how the command works
	Description string
	// The function to call when this command is invoked
	Action func(context *Context)
	// List of flags to parse
	Flags []Flag
}

// Invokes the command given the context, parses ctx.Args() to generate command-specific flags
func (c Command) Run(ctx *Context) {
	// append help to flags
	c.Flags = append(
		c.Flags,
		helpFlag{"show help"},
	)

	set := flagSet(c.Name, c.Flags)
	set.SetOutput(ioutil.Discard)

	firstFlagIndex := -1
	for index, arg := range ctx.Args() {
		if strings.HasPrefix(arg, "-") {
			firstFlagIndex = index
			break
		}
	}

	var err error
	if firstFlagIndex > -1 {
		args := ctx.Args()[1:firstFlagIndex]
		flags := ctx.Args()[firstFlagIndex:]
		err = set.Parse(append(flags, args...))
	} else {
		err = set.Parse(ctx.Args()[1:])
	}

	if err != nil {
		fmt.Println("Incorrect Usage.\n")
		ShowCommandHelp(ctx, c.Name)
		fmt.Println("")
		os.Exit(1)
	}

	context := NewContext(ctx.App, set, ctx.globalSet)
	checkCommandHelp(context, c.Name)
	c.Action(context)
}

// Returns true if Command.Name or Command.ShortName matches given name
func (c Command) HasName(name string) bool {
	return c.Name == name || c.ShortName == name
}
