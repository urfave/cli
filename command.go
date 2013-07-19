package cli

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Handler
	Flags       []Flag
}

func (command Command) Run(c *Context) {
	set := flagSet(command.Flags)
	set.Parse(c.Args()[1:])
	command.Action(NewContext(set, c.globalSet))
}
