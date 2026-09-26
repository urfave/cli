//go:build !urfave_cli_no_template

package cli

import (
	"io"
	"strings"
	"text/tabwriter"
	"text/template"
)

// templatesSupported is true unless built with the urfave_cli_no_template tag.
const templatesSupported = true

// RootCommandHelpTemplate is the text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var RootCommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[global options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

AUTHOR{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

GLOBAL OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{template "copyrightTemplate" .}}{{end}}
`

// CommandHelpTemplate is the text template for the command help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}

GLOBAL OPTIONS:{{template "visiblePersistentFlagTemplate" .}}{{end}}
`

// SubcommandHelpTemplate is the text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}

GLOBAL OPTIONS:{{template "visiblePersistentFlagTemplate" .}}{{end}}
`

var FishCompletionTemplate = `# {{ .Command.Name }} fish shell completion

function __fish_{{ .Command.Name }}_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i{{ range $v := .AllCommands }} {{ $v }}{{ end }}
            return 1
        end
    end
    return 0
end

{{ range $v := .Completions }}{{ $v }}
{{ end }}`

// DefaultPrintHelpCustom is the default implementation of HelpPrinterCustom.
//
// The customFuncs map will be combined with a default template.FuncMap to
// allow using arbitrary functions in template rendering.
func DefaultPrintHelpCustom(out io.Writer, templ string, data any, customFuncs map[string]any) {
	const maxLineLength = 10000

	tracef("building default funcMap")
	funcMap := template.FuncMap{
		"join":           strings.Join,
		"subtract":       subtract,
		"indent":         indent,
		"nindent":        nindent,
		"trim":           strings.TrimSpace,
		"wrap":           func(input string, offset int) string { return wrap(input, offset, maxLineLength) },
		"offset":         offset,
		"offsetCommands": offsetCommands,
	}

	if wa, ok := customFuncs["wrapAt"]; ok {
		if wrapAtFunc, ok := wa.(func() int); ok {
			wrapAt := wrapAtFunc()
			customFuncs["wrap"] = func(input string, offset int) string {
				return wrap(input, offset, wrapAt)
			}
		}
	}

	for key, value := range customFuncs {
		funcMap[key] = value
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))

	if _, err := t.New("helpNameTemplate").Parse(helpNameTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("argsTemplate").Parse(argsTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("usageTemplate").Parse(usageTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("descriptionTemplate").Parse(descriptionTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleCommandTemplate").Parse(visibleCommandTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("copyrightTemplate").Parse(copyrightTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("versionTemplate").Parse(versionTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleFlagCategoryTemplate").Parse(visibleFlagCategoryTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleFlagTemplate").Parse(visibleFlagTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visiblePersistentFlagTemplate").Parse(visiblePersistentFlagTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleGlobalFlagCategoryTemplate").Parse(strings.ReplaceAll(visibleFlagCategoryTemplate, "OPTIONS", "GLOBAL OPTIONS")); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("authorsTemplate").Parse(authorsTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleCommandCategoryTemplate").Parse(visibleCommandCategoryTemplate); err != nil {
		handleTemplateError(err)
	}

	tracef("executing template")
	handleTemplateError(t.Execute(w, data))

	_ = w.Flush()
}

func renderFishCompletion(w io.Writer, data *fishCommandCompletionTemplate) error {
	const name = "cli"
	t, err := template.New(name).Parse(FishCompletionTemplate)
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, name, data)
}
