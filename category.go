package cli

// CommandCategories wraps a slice of *CommandCategory.
type CommandCategories struct {
	slice []*CommandCategory
}

func NewCommandCategories() *CommandCategories {
	return &CommandCategories{slice: []*CommandCategory{}}
}

func (c *CommandCategories) Less(i, j int) bool {
	return c.slice[i].Name < c.slice[j].Name
}

func (c *CommandCategories) Len() int {
	return len(c.slice)
}

func (c *CommandCategories) Swap(i, j int) {
	c.slice[i], c.slice[j] = c.slice[j], c.slice[i]
}

// AddCommand adds a command to a category, creating a new category if necessary.
func (c *CommandCategories) AddCommand(category string, command *Command) {
	for _, commandCategory := range c.slice {
		if commandCategory.Name == category {
			commandCategory.commands = append(commandCategory.commands, command)
			return
		}
	}
	c.slice = append(c.slice,
		&CommandCategory{Name: category, commands: []*Command{command}})
}

// Categories returns a copy of the category slice
func (c *CommandCategories) Categories() []*CommandCategory {
	ret := make([]*CommandCategory, len(c.slice))
	copy(ret, c.slice)
	return ret
}

// CommandCategory is a category containing commands.
type CommandCategory struct {
	Name string

	commands []*Command
}

// VisibleCommands returns a slice of the Commands with Hidden=false
func (c *CommandCategory) VisibleCommands() []*Command {
	ret := []*Command{}
	for _, command := range c.commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}
