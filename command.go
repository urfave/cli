package cli

import (
	"fmt"
	"io/ioutil"
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
func (c Command) Run(ctx *Context) error {
	// append help to flags
	c.Flags = append(
		c.Flags,
		BoolFlag{"help, h", "show help"},
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
		args := ctx.Args()
		regularArgs := args[1:firstFlagIndex]
		flagArgs := args[firstFlagIndex:]
		err = set.Parse(append(flagArgs, regularArgs...))
	} else {
		err = set.Parse(ctx.Args().Tail())
	}

	if err != nil {
		fmt.Printf("Incorrect Usage.\n\n")
		ShowCommandHelp(ctx, c.Name)
		fmt.Println("")
		return err
	}

	nerr := normalizeFlags(c.Flags, set)
	if nerr != nil {
		fmt.Println(nerr)
		fmt.Println("")
		ShowCommandHelp(ctx, c.Name)
		fmt.Println("")
		return nerr
	}
	context := NewContext(ctx.App, set, ctx.globalSet)
	if checkCommandHelp(context, c.Name) {
		return nil
	}
	c.Action(context)
	return nil
}

// Returns true if Command.Name or Command.ShortName matches given name
func (c Command) HasName(name string) bool {
	return c.Name == name || c.ShortName == name
}
