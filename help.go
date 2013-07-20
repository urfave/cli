package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"
)

// The text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [global options] command [command options] [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

// The text template for the command help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CommandHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   command {{.Name}} [command options] [arguments...]

DESCRIPTION:
   {{.Description}}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

var helpCommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "Shows a list of commands or help for one command",
	Action: func(c *Context) {
		args := c.Args()
		if len(args) > 0 {
			showCommandHelp(c)
		} else {
			showAppHelp(c)
		}
	},
}

func showAppHelp(c *Context) {
	printHelp(AppHelpTemplate, c.App)
}

func showCommandHelp(c *Context) {
	name := c.Args()[0]
	for _, c := range c.App.Commands {
		if c.HasName(name) {
			printHelp(CommandHelpTemplate, c)
			return
		}
	}

	fmt.Printf("No help topic for '%v'\n", name)
	os.Exit(1)
}

func printHelp(templ string, data interface{}) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(templ))
	t.Execute(w, data)
	w.Flush()
}

func showVersion(c *Context) {
	fmt.Printf("%v version %v\n", c.App.Name, c.App.Version)
}
