package cli

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Action
	Flags       []Flag
}

func (c Command) Run(ctx *Context) {
	set := flagSet(c.Name, c.Flags)
	set.Parse(ctx.Args()[1:])
	c.Action.Execute(NewContext(ctx.App, set, ctx.globalSet))
}

func (c Command) SetAction(function interface{}) {
  c.Action = ParseAction(function)
}

func (c Command) HasName(name string) bool {
	return c.Name == name || c.ShortName == name
}
