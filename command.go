package cli

import (
	"io/ioutil"
	"os"
	"fmt"
)

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      func(context *Context)
	Flags       []Flag
}

func (c Command) Run(ctx *Context) {
	// append help to flags
	c.Flags = append(
		c.Flags,
		helpFlag{"show help"},
	)

	set := flagSet(c.Name, c.Flags)
	set.SetOutput(ioutil.Discard)
	err := set.Parse(ctx.Args()[1:])

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

func (c Command) HasName(name string) bool {
	return c.Name == name || c.ShortName == name
}
