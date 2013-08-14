package cli

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
	set.Parse(ctx.Args()[1:])

	context := NewContext(ctx.App, set, ctx.globalSet)
	checkCommandHelp(context, c.Name)

	c.Action(context)
}

func (c Command) HasName(name string) bool {
	return c.Name == name || c.ShortName == name
}
