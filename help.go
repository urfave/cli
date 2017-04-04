package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
)

// AppHelpTemplate is the text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{.Copyright}}{{end}}
`

// CommandHelpTemplate is the text template for the command help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .VisibleFlags}}

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

// SubcommandHelpTemplate is the text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

var helpCommand = &Command{
	Name:      "help",
	Aliases:   []string{"h"},
	Usage:     "Shows a list of commands or help for one command",
	ArgsUsage: "[command]",
	Action: func(c *Context) error {
		args := c.Args()
		if args.Present() {
			return ShowCommandHelp(c, args.First())
		}

		ShowAppHelp(c)
		return nil
	},
}

var helpSubcommand = &Command{
	Name:      "help",
	Aliases:   []string{"h"},
	Usage:     "Shows a list of commands or help for one command",
	ArgsUsage: "[command]",
	Action: func(c *Context) error {
		args := c.Args()
		if args.Present() {
			return ShowCommandHelp(c, args.First())
		}

		return ShowSubcommandHelp(c)
	},
}

// Prints help for the App or Command
type helpPrinter func(w io.Writer, templ string, data interface{})

// HelpPrinter is a function that writes the help output. If not set a default
// is used. The function signature is:
// func(w io.Writer, templ string, data interface{})
var HelpPrinter helpPrinter = printHelp

// VersionPrinter prints the version for the App
var VersionPrinter = printVersion

// ShowAppHelp is an action that displays the help.
func ShowAppHelp(c *Context) {
	HelpPrinter(c.App.Writer, AppHelpTemplate, c.App)
}

// DefaultAppComplete prints the list of subcommands as the default app completion method
func DefaultAppComplete(c *Context) {
	for _, command := range c.App.Commands {
		if command.Hidden {
			continue
		}
		for _, name := range command.Names() {
			fmt.Fprintln(c.App.Writer, name)
		}
	}
}

// ShowCommandHelp prints help for the given command
func ShowCommandHelp(ctx *Context, command string) error {
	// show the subcommand help for a command with subcommands
	if command == "" {
		HelpPrinter(ctx.App.Writer, SubcommandHelpTemplate, ctx.App)
		return nil
	}

	for _, c := range ctx.App.Commands {
		if c.HasName(command) {
			HelpPrinter(ctx.App.Writer, CommandHelpTemplate, c)
			return nil
		}
	}

	if ctx.App.CommandNotFound == nil {
		return Exit(fmt.Sprintf("No help topic for '%v'", command), 3)
	}

	ctx.App.CommandNotFound(ctx, command)
	return nil
}

// ShowSubcommandHelp prints help for the given subcommand
func ShowSubcommandHelp(c *Context) error {
	if c == nil {
		return nil
	}

	if c.Command != nil {
		return ShowCommandHelp(c, c.Command.Name)
	}

	return ShowCommandHelp(c, "")
}

// ShowVersion prints the version number of the App
func ShowVersion(c *Context) {
	VersionPrinter(c)
}

func printVersion(c *Context) {
	fmt.Fprintf(c.App.Writer, "%v version %v\n", c.App.Name, c.App.Version)
}

// ShowCompletions prints the lists of commands within a given context
func ShowCompletions(c *Context) {
	a := c.App
	if a != nil && a.ShellComplete != nil {
		a.ShellComplete(c)
	}
}

// ShowCommandCompletions prints the custom completions for a given command
func ShowCommandCompletions(ctx *Context, command string) {
	c := ctx.App.Command(command)
	if c != nil && c.ShellComplete != nil {
		c.ShellComplete(ctx)
	}
}

func printHelp(out io.Writer, templ string, data interface{}) {
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))

	errDebug := os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != ""

	err := t.Execute(w, data)
	if err != nil {
		if errDebug {
			fmt.Fprintf(ErrWriter, "CLI TEMPLATE ERROR: %#v\n", err)
		}
		return
	}

	w.Flush()
}

func checkVersion(c *Context) bool {
	found := false
	if VersionFlag.Name != "" {
		for _, name := range VersionFlag.Names() {
			if c.Bool(name) {
				found = true
			}
		}
	}
	return found
}

func checkHelp(c *Context) bool {
	found := false
	if HelpFlag.Name != "" {
		for _, name := range HelpFlag.Names() {
			if c.Bool(name) {
				found = true
			}
		}
	}
	return found
}

func checkCommandHelp(c *Context, name string) bool {
	if c.Bool("h") || c.Bool("help") {
		ShowCommandHelp(c, name)
		return true
	}

	return false
}

func checkSubcommandHelp(c *Context) bool {
	if c.Bool("h") || c.Bool("help") {
		ShowSubcommandHelp(c)
		return true
	}

	return false
}

func checkCompletions(c *Context) bool {
	if c.Bool(GenerateCompletionFlag.Name) && c.App.EnableShellCompletion {
		ShowCompletions(c)
		return true
	}

	return false
}

func checkCommandCompletions(c *Context, name string) bool {
	if c.Bool(GenerateCompletionFlag.Name) && c.App.EnableShellCompletion {
		ShowCommandCompletions(c, name)
		return true
	}

	return false
}

func checkInitCompletion(c *Context) (bool, error) {
	if c.IsSet(InitCompletionFlag.Name) {
		shell := c.String(InitCompletionFlag.Name)
		progName := os.Args[0]
		switch shell {
		case "bash":
			fmt.Print(bashCompletionCode(progName))
			return true, nil
		case "zsh":
			fmt.Print(zshCompletionCode(progName))
			return true, nil
		default:
			return false, fmt.Errorf("--init-completion value cannot be '%s'", shell)
		}
	}
	return false, nil
}

func bashCompletionCode(progName string) string {
	var template = `_cli_bash_autocomplete() {
     local cur opts base;
     COMPREPLY=();
     cur="${COMP_WORDS[COMP_CWORD]}";
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-completion );
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) );
     return 0;
};
complete -F _cli_bash_autocomplete %s`
	return fmt.Sprintf(template, progName)
}

func zshCompletionCode(progName string) string {
	var template = `autoload -U compinit && compinit;
autoload -U bashcompinit && bashcompinit;`

	return template + "\n" + bashCompletionCode(progName)
}
