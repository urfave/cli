package cli

// CommandCategories is a slice of *CommandCategory.
type CommandCategories struct {
	Categories []*CommandCategory
}

func NewCommandCategories() *CommandCategories {
	return &CommandCategories{Categories: []*CommandCategory{}}
}

// CommandCategory is a category containing commands.
type CommandCategory struct {
	Name     string
	Commands []*Command
}

func (c *CommandCategories) Less(i, j int) bool {
	return c.Categories[i].Name < c.Categories[j].Name
}

func (c *CommandCategories) Len() int {
	return len(c.Categories)
}

func (c *CommandCategories) Swap(i, j int) {
	c.Categories[i], c.Categories[j] = c.Categories[j], c.Categories[i]
}

// AddCommand adds a command to a category.
func (c *CommandCategories) AddCommand(category string, command *Command) *CommandCategories {
	for _, commandCategory := range c.Categories {
		if commandCategory.Name == category {
			commandCategory.Commands = append(commandCategory.Commands, command)
			return c
		}
	}
	c.Categories = append(c.Categories,
		&CommandCategory{Name: category, Commands: []*Command{command}})
	return c
}

// VisibleCommands returns a slice of the Commands with Hidden=false
func (c *CommandCategory) VisibleCommands() []*Command {
	ret := []*Command{}
	for _, command := range c.Commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}
