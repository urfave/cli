package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"
)

var helpCommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "Shows a list of commands or help for one command",
	Action: func(c *Context) {
		helpTemplate := `NAME:
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

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		t := template.Must(template.New("help").Parse(helpTemplate))
		t.Execute(w, c.App)
		w.Flush()
	},
}

func showVersion(c *Context) {
	fmt.Printf("%v version %v\n", c.App.Name, c.App.Version)
}
