package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Name = "flaggen"
	app.Usage = "Generate flag type code!"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "package, p",
			Value: "cli",
			Usage: "`Name of the package` for which the flag types will be generated",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
