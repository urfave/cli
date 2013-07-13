package main

import "os"

func main() {
	app := App{
		Name:        "math",
		Description: "a simple command line math utility",
		Commands: []Command{{
			Name:        "add",
			Description: "Add 2 and 2",
			Action:      DoAdd,
		}, {
			Name:        "subtract",
			Description: "Subtract 2 and 2",
			Action:      DoSubtract,
		}, {
			Name:        "multiply",
			Description: "Multiply 2 and 2",
			Action:      DoMultiply,
		}, {
			Name:        "divide",
			Description: "Divide 2 and 2",
			Action:      DoDivide,
		}},
	}

	app.Run(os.Args[1])
}

func DoAdd(name string) {
	println("2+2=", 2+2)
}

func DoSubtract(name string) {
	println("2-2=", 2-2)
}

func DoMultiply(name string) {
	println("2*2=", 2*2)
}

func DoDivide(name string) {
	println("2/2=", 2/2)
}

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
