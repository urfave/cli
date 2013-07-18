package cli

import "os"
import "text/tabwriter"
import "text/template"

type HelpData struct {
	Name     string
	Usage    string
	Version  string
	Commands []Command
	Flags    []Flag
}

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
	data := HelpData{
		Name,
		Usage,
		Version,
		append(Commands, HelpCommand),
		Flags,
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	t.Execute(w, data)
	w.Flush()
}
