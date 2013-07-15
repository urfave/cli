package cli

import "os"
import "text/tabwriter"
import "text/template"

type HelpData struct {
	Name     string
	Usage    string
	Commands []Command
	Version  string
}

var HelpCommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "Shows a list of commands or help for one command",
}

func init() {
	HelpCommand.Action = ShowHelp
}

func ShowHelp(name string) {
	helpTemplate := `NAME:
    {{.Name}} - {{.Usage}}

USAGE:
    {{.Name}} [global options] command [command options] [arguments...]

VERSION:
    {{.Version}}

COMMANDS:
    {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
    {{end}}
`
	data := HelpData{
		Name,
		Usage,
		append(Commands, HelpCommand),
		Version,
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	t.Execute(w, data)
	w.Flush()
}
