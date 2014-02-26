package cli_test

import (
	"os"

	"github.com/AntonioMeireles/cli"
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
			Action: func(c *cli.Context) (err error) {
				println("added task: ", c.Args().First())
				return err
			},
		},
		{
			Name:      "complete",
			ShortName: "c",
			Usage:     "complete a task on the list",
			Action: func(c *cli.Context) (err error) {
				println("completed task: ", c.Args().First())
				return err
			},
		},
	}

	app.Run(os.Args)
}
