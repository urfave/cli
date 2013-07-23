package cli_test

import (
	"fmt"
  "github.com/codegangsta/cli"
	"os"
)

func ExampleContextApp() {
	// set args for examples sake
	os.Args = []string{"greet", "--name", "Jeremy"}

	app := cli.NewApp()
	app.Name = "greet"
	app.Flags = []cli.Flag{
		cli.StringFlag{"name", "bob", "a name to say"},
	}
	app.SetAction(func(c *cli.Context) {
		fmt.Printf("Hello %v\n", c.String("name"))
	})
	app.Run(os.Args)
	// Output:
	// Hello Jeremy
}

func ExamplePlainApp() {
	// set args for examples sake
	app := cli.NewApp()
	app.Name = "greet"
	app.Flags = []cli.Flag{
		cli.StringFlag{"name", "bob", "a name to say"},
	}
	app.SetAction(func() {
		fmt.Printf("Hello Jeremy")
	})
	app.Run(os.Args)
	// Output:
	// Hello Jeremy
}
