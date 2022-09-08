---
tags:
  - v2
search:
  boost: 2
---

You can wrap command usage, e.g.:

<!-- {
  "output": "NAME:
   long - Long command description cli

USAGE:
   long  command [command options] [arguments...]

COMMANDS:
   help, h     
   testing, t  aaaaaaaaa aaaaaaaaa aaaaa aaaaaaaaa aaaaaaaaaaaaaaa aaaaa
               aaaaaaaaaaaa aaa aa aaaaaa aa aaaa"
} -->
```go
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "long",
		Usage: "Long command description cli",


		Commands: []*cli.Command{
			{
				Name:    "help",
				Aliases: []string{"h"},
				Action: func(cCtx *cli.Context) error {
					_ = ShowAppHelp(cCtx)
					return nil
				},
			},
			{
				Name:    "testing",
				Aliases: []string{"t"},
				Usage:   "aaaaaaaaa aaaaaaaaa aaaaa aaaaaaaaa aaaaaaaaaaaaaaa aaaaa aaaaaaaaaaaa aaa aa aaaaaa aa aaaa",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func ShowAppHelp(c *cli.Context) error {
	template := c.App.CustomAppHelpTemplate
	if template == "" {
		template = cli.AppHelpTemplate
	}

	customAppData := func() map[string]interface{} {
		return map[string]interface{}{
			"wrapAt": func() int {
				return 80
			},
		}
	}
	cli.HelpPrinterCustom(c.App.Writer, template, c.App, customAppData())
	return nil
}
```

In this example, we fix sentence length to 80 characters.

