package cli

import "os"
import "text/tabwriter"
import "text/template"

var HelpCommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "Shows a list of commands or help for one command",
}

func init() {
	HelpCommand.Action = ShowHelp
}

func ShowHelp(c *Context) {
	helpTemplate := `NAME:
    {{.Name}} - {{.Usage}}

USAGE:
    {{.Name}} [global options] command [command options] [arguments...]

VERSION:
    {{.Version}}

COMMANDS:
    {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
    {{end}}
GLOBAL OPTIONS
    {{range .Flags}}{{.}}
    {{end}}
`

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	t.Execute(w, c.App)
	w.Flush()
}
