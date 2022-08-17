---
tags:
  - v2
search:
  boost: 2
---

The default help flag (`-h/--help`) is defined as `cli.HelpFlag` and is checked
by the cli internals in order to print generated help text for the app, command,
or subcommand, and break execution.

#### Customization

All of the help text generation may be customized, and at multiple levels.  The
templates are exposed as variables `AppHelpTemplate`, `CommandHelpTemplate`, and
`SubcommandHelpTemplate` which may be reassigned or augmented, and full override
is possible by assigning a compatible func to the `cli.HelpPrinter` variable,
e.g.:

<!-- {
  "output": "Ha HA.  I pwnd the help!!1"
} -->
```go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	// EXAMPLE: Append to an existing template
	cli.AppHelpTemplate = fmt.Sprintf(`%s

WEBSITE: http://awesometown.example.com

SUPPORT: support@awesometown.example.com

`, cli.AppHelpTemplate)

	// EXAMPLE: Override a template
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`

	// EXAMPLE: Replace the `HelpPrinter` func
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		fmt.Println("Ha HA.  I pwnd the help!!1")
	}

	(&cli.App{}).Run(os.Args)
}
```

The default flag may be customized to something other than `-h/--help` by
setting `cli.HelpFlag`, e.g.:

<!-- {
  "args": ["&#45;&#45halp"],
  "output": "haaaaalp.*HALP"
} -->
```go
package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "haaaaalp",
		Aliases: []string{"halp"},
		Usage:   "HALP",
		EnvVars: []string{"SHOW_HALP", "HALPPLZ"},
	}

	(&cli.App{}).Run(os.Args)
}
```
