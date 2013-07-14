package cli

type App struct {
	Name     string
	Usage    string
	Action   Action
	Commands []Command
}

type Command struct {
	Name        string
	ShortName   string
	Usage       string
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
