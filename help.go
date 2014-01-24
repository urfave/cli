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
   {{range .Commands}}{{if not .Hidden}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
AUTHOR:
    Written by {{.Author}}.
{{if .Reporting}}
REPORTING BUGS:
    {{.Reporting}}
    {{end}}
COPYRIGHT:
    Copyright © {{.Copyright}} {{if .CopyrightHolder}}{{.CopyrightHolder}}{{else}}{{.Author}}{{end}}
    {{if .License}}Licensed under the {{.License}}
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
		if args.Present() {
			ShowCommandHelp(c, args.First())
		} else {
			ShowAppHelp(c)
		}
	},
}

// Prints help for the App
func ShowAppHelp(c *Context) {
	printHelp(AppHelpTemplate, c.App)
}

// Prints help for the given command
func ShowCommandHelp(c *Context, command string) {
	for _, c := range c.App.Commands {
		if c.HasName(command) {
			printHelp(CommandHelpTemplate, c)
			return
		}
	}

	fmt.Printf("No help topic for '%v'\n", command)
}

// Prints the available metadata about the App
func ShowVersion(c *Context) {
	fmt.Printf("%v version %v\n", c.App.Name, c.App.Version)
	if c.App.CopyrightHolder != "" {
		fmt.Printf("Copyright © %v %v\n", c.App.Copyright,
			c.App.CopyrightHolder)
	} else {
		fmt.Printf("Copyright © %v %v\n", c.App.Copyright,
			c.App.Author)
	}
	if c.App.License != "" {
		fmt.Printf("Licensed under the %v.\n", c.App.License)
	}
	// we assume it would be != from Author so we don't test it
	if c.App.CopyrightHolder != "" {
		fmt.Printf("Written by %v.\n", c.App.Author)
	}
}

func printHelp(templ string, data interface{}) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(templ))
	err := t.Execute(w, data)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

func checkVersion(c *Context) bool {
	if c.GlobalBool("version") {
		ShowVersion(c)
		return true
	}

	return false
}

func checkHelp(c *Context) bool {
	if c.GlobalBool("h") || c.GlobalBool("help") {
		ShowAppHelp(c)
		return true
	}

	return false
}

func checkCommandHelp(c *Context, name string) bool {
	if c.Bool("h") || c.Bool("help") {
		ShowCommandHelp(c, name)
		return true
	}

	return false
}
