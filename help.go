package cli

import "os"
import "log"

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
	Usage:     "View help topics",
	Action:    ShowHelp,
}

var helpTemplate = `NAME:
  {{.Name}} - {{.Usage}}

USAGE:
  {{.Name}} [global-options] COMMAND [command-options]

VERSION:
  {{.Version}}

COMMANDS:
  {{range .Commands}}{{.Name}}{{ "\t" }}{{.Usage}}
  {{end}}
  
`

var ShowHelp = func(name string) {

	data := HelpData{
		Name,
		Usage,
		Commands,
		Version,
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	err := t.Execute(w, data)
	w.Flush()
	if err != nil {
		log.Println("executing template:", err)
	}
	// fmt.Printf("Usage: %v [global-options] COMMAND [command-options]\n\n", Name)
	// if Commands != nil {
	// 	fmt.Printf("The most commonly used %v commands are:\n", Name)
	// 	for _, c := range Commands {
	// 		fmt.Fprintln(w, "   "+c.Name+"\t"+c.Usage)
	// 	}
	// 	w.Flush()
	// }
}
