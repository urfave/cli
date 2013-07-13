package cli

type App struct {
	Name        string
	Description string
	Commands    []Command
}

type Command struct {
	Name        string
	Description string
	Action      Action
}

type Action func(name string)

func (a App) Run(command string) {
	for _, c := range a.Commands {
		if c.Name == command {
			c.Action(command)
		}
	}
}
