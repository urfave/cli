package cli_test

import (
	"github.com/codegangsta/cli"
	"os"
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
			Action: func(c *cli.Context) {
				println("added task: ", c.Args().First())
			},
		},
		{
			Name:      "complete",
			ShortName: "c",
			Usage:     "complete a task on the list",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
	}

	app.Run(os.Args)
}
