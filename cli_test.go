package cli_test

import (
	"os"

	"github.com/codegangsta/cli"
)

func Example() {
	app := cli.NewApp()
	app.Name = "todo"
	app.Usage = "task list on the command line"
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a task to the list",
			Action: func(c *cli.Context) int {
				println("added task: ", c.Args().First())
				return 0
			},
		},
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task on the list",
			Action: func(c *cli.Context) int {
				println("completed task: ", c.Args().First())
				return 0
			},
		},
	}

	app.Run(os.Args)
}

func ExampleSubcommand() {
	app := cli.NewApp()
	app.Name = "say"
	app.Commands = []cli.Command{
		{
			Name:        "hello",
			Aliases:     []string{"hi"},
			Usage:       "use it to see a description",
			Description: "This is how we describe hello the function",
			Subcommands: []cli.Command{
				{
					Name:        "english",
					Aliases:     []string{"en"},
					Usage:       "sends a greeting in english",
					Description: "greets someone in english",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "Bob",
							Usage: "Name of the person to greet",
						},
					},
					Action: func(c *cli.Context) int {
						println("Hello, ", c.String("name"))
						return 0
					},
				}, {
					Name:    "spanish",
					Aliases: []string{"sp"},
					Usage:   "sends a greeting in spanish",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "surname",
							Value: "Jones",
							Usage: "Surname of the person to greet",
						},
					},
					Action: func(c *cli.Context) int {
						println("Hola, ", c.String("surname"))
						return 0
					},
				}, {
					Name:    "french",
					Aliases: []string{"fr"},
					Usage:   "sends a greeting in french",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "nickname",
							Value: "Stevie",
							Usage: "Nickname of the person to greet",
						},
					},
					Action: func(c *cli.Context) int {
						println("Bonjour, ", c.String("nickname"))
						return 0
					},
				},
			},
		}, {
			Name:  "bye",
			Usage: "says goodbye",
			Action: func(c *cli.Context) int {
				println("bye")
				return 0
			},
		},
	}

	app.Run(os.Args)
}
