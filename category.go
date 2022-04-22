package cli

// CommandCategories interface allows for category manipulation
type CommandCategories interface {
	// AddCommand adds a command to a category, creating a new category if necessary.
	AddCommand(category string, command *Command)
	// categories returns a copy of the category slice
	Categories() []CommandCategory
}

type commandCategories []*commandCategory

func newCommandCategories() CommandCategories {
	ret := commandCategories([]*commandCategory{})
	return &ret
}

func (c *commandCategories) Less(i, j int) bool {
	return lexicographicLess((*c)[i].Name(), (*c)[j].Name())
}

func (c *commandCategories) Len() int {
	return len(*c)
}

func (c *commandCategories) Swap(i, j int) {
	(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
}

func (c *commandCategories) AddCommand(category string, command *Command) {
	for _, commandCategory := range []*commandCategory(*c) {
		if commandCategory.name == category {
			commandCategory.commands = append(commandCategory.commands, command)
			return
		}
	}
	newVal := append(*c,
		&commandCategory{name: category, commands: []*Command{command}})
	*c = newVal
}

func (c *commandCategories) Categories() []CommandCategory {
	ret := make([]CommandCategory, len(*c))
	for i, cat := range *c {
		ret[i] = cat
	}
	return ret
}

// CommandCategory is a category containing commands.
type CommandCategory interface {
	// Name returns the category name string
	Name() string
	// VisibleCommands returns a slice of the Commands with Hidden=false
	VisibleCommands() []*Command
}

type commandCategory struct {
	name     string
	commands []*Command
}

func (c *commandCategory) Name() string {
	return c.name
}

func (c *commandCategory) VisibleCommands() []*Command {
	if c.commands == nil {
		c.commands = []*Command{}
	}

	var ret []*Command
	for _, command := range c.commands {
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
	Flags []Flag
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
func (c *FlagCategory) VisibleFlags() []VisibleFlag {
	ret := []VisibleFlag{}
	for _, fl := range c.Flags {
		if vf, ok := fl.(VisibleFlag); ok {
			if vf.IsVisible() {
				ret = append(ret, vf)
			}
		}
	}
	return ret
}
