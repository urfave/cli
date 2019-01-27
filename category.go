package cli

// CommandCategories is a slice of *CommandCategory.
type CommandCategories []*CommandCategory

// CommandCategory is a category containing commands.
type CommandCategory struct {
	Name     string
	Commands Commands
}

func (c CommandCategories) Less(i, j int) bool {
	return lexicographicLess(c[i].Name, c[j].Name)
}

func (c CommandCategories) Len() int {
	return len(c)
}

func (c CommandCategories) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// AddCommand adds a command to a category.
func (c CommandCategories) AddCommand(category string, command Command) CommandCategories {
	for _, commandCategory := range c {
		if commandCategory.Name == category {
			commandCategory.Commands = append(commandCategory.Commands, command)
			return c
		}
	}
	return append(c, &CommandCategory{Name: category, Commands: []Command{command}})
}

// VisibleCommands returns a slice of the Commands with Hidden=false
func (c *CommandCategory) VisibleCommands() []Command {
	ret := []Command{}
	for _, command := range c.Commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}

// FlagCategories is a slice of *FlagCategory.
type FlagCategories []*FlagCategory

// FlagCategory is a category containing commands.
type FlagCategory struct {
	Name  string
	Flags Flags
}

func (f FlagCategories) Less(i, j int) bool {
	return lexicographicLess(f[i].Name, f[j].Name)
}

func (f FlagCategories) Len() int {
	return len(f)
}

func (f FlagCategories) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// AddFlags adds a command to a category.
func (f FlagCategories) AddFlag(category string, flag Flag) FlagCategories {
	for _, flagCategory := range f {
		if flagCategory.Name == category {
			flagCategory.Flags = append(flagCategory.Flags, flag)
			return f
		}
	}
	return append(f, &FlagCategory{Name: category, Flags: []Flag{flag}})
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (c *FlagCategory) VisibleFlags() []Flag {
	ret := []Flag{}
	for _, flag := range c.Flags {
		if !flag.GetHidden() {
			ret = append(ret, flag)
		}
	}
	return ret
}
