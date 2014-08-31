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
			Name:      "add",
			ShortName: "a",
			Usage:     "add a task to the list",
			Action: func(c *cli.Context) error {
				println("added task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:      "complete",
			ShortName: "c",
			Usage:     "complete a task on the list",
			Action: func(c *cli.Context) error {
				println("completed task: ", c.Args().First())
				return nil
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
			ShortName:   "hi",
			Usage:       "use it to see a description",
			Description: "This is how we describe hello the function",
			Subcommands: []cli.Command{
				{
					Name:        "english",
					ShortName:   "en",
					Usage:       "sends a greeting in english",
					Description: "greets someone in english",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "Bob",
							Usage: "Name of the person to greet",
						},
					},
					Action: func(c *cli.Context) error {
						println("Hello, ", c.String("name"))
						return nil
					},
				}, {
					Name:      "spanish",
					ShortName: "sp",
					Usage:     "sends a greeting in spanish",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "surname",
							Value: "Jones",
							Usage: "Surname of the person to greet",
						},
					},
					Action: func(c *cli.Context) error {
						println("Hola, ", c.String("surname"))
						return nil
					},
				}, {
					Name:      "french",
					ShortName: "fr",
					Usage:     "sends a greeting in french",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "nickname",
							Value: "Stevie",
							Usage: "Nickname of the person to greet",
						},
					},
					Action: func(c *cli.Context) error {
						println("Bonjour, ", c.String("nickname"))
						return nil
					},
				},
			},
		}, {
			Name:  "bye",
			Usage: "says goodbye",
			Action: func(c *cli.Context) error {
				println("bye")
				return nil
			},
		},
	}

	app.Run(os.Args)
}
