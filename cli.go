package cli

type Application struct {
	Name     string
	Usage    string
	Action   Action
	Commands []Command
}

func (a *Application) Run(args []string) {
  help := Command{
    Name: "help",
    Usage: "View help topics",
    Action: func(name string) {
      println("usage: " + a.Name + " [global-options] COMMAND [command-options]")
    },
  }

	command := args[1]
	for _, c := range a.Commands {
		if c.Name == command {
			c.Action(command)
      return
		}
	}

  // Run default action
  help.Action("foo")
}

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Action
}

type Action func(name string)
