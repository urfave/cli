package cli

var helpNameTemplate = `{{$v := offset .FullName 6}}{{wrap .FullName 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}`
var usageTemplate = `{{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleFlags}} [command [command options]]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}`
var descriptionTemplate = `{{wrap .Description 3}}`
var authorsTemplate = `{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}`
var visibleCommandTemplate = `{{ $cv := offsetCommands .VisibleCommands 5}}{{range .VisibleCommands}}
   {{$s := join .Names ", "}}{{$s}}{{ $sp := subtract $cv (offset $s 3) }}{{ indent $sp ""}}{{wrap .Usage $cv}}{{end}}`
var visibleCommandCategoryTemplate = `{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{template "visibleCommandTemplate" .}}{{end}}{{end}}`
var visibleFlagCategoryTemplate = `{{range .VisibleFlagCategories}}
   {{if .Name}}{{.Name}}

   {{end}}{{$flglen := len .Flags}}{{range $i, $e := .Flags}}{{if eq (subtract $flglen $i) 1}}{{$e}}
{{else}}{{$e}}
   {{end}}{{end}}{{end}}`

var visibleFlagTemplate = `{{range $i, $e := .VisibleFlags}}
   {{wrap $e.String 6}}{{end}}`

var versionTemplate = `{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}`

var copyrightTemplate = `{{wrap .Copyright 3}}`

// RootCommandHelpTemplate is the text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var RootCommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[global options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

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

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

// SubcommandHelpTemplate is the text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleCommands}}[command [command options]] {{end}}{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

var MarkdownDocTemplate = `{{if gt .SectionNum 0}}% {{ .Command.Name }} {{ .SectionNum }}

{{end}}# NAME

{{ .Command.Name }}{{ if .Command.Usage }} - {{ .Command.Usage }}{{ end }}

# SYNOPSIS

{{ .Command.Name }}
{{ if .SynopsisArgs }}
` + "```" + `
{{ range $v := .SynopsisArgs }}{{ $v }}{{ end }}` + "```" + `
{{ end }}{{ if .Command.Description }}
# DESCRIPTION

{{ .Command.Description }}
{{ end }}
**Usage**:

` + "```" + `{{ if .Command.UsageText }}
{{ .Command.UsageText }}
{{ else }}
{{ .Command.Name }} [GLOBAL OPTIONS] [command [COMMAND OPTIONS]] [ARGUMENTS...]
{{ end }}` + "```" + `
{{ if .GlobalArgs }}
# GLOBAL OPTIONS
{{ range $v := .GlobalArgs }}
{{ $v }}{{ end }}
{{ end }}{{ if .Commands }}
# COMMANDS
{{ range $v := .Commands }}
{{ $v }}{{ end }}{{ end }}`

var MarkdownTabularDocTemplate = `{{ define "flags" }}
| Name | Description | Default value | Environment variables |
|------|-------------|:-------------:|:---------------------:|
{{   range $flag := . -}}
{{- /**/ -}} | ` + "`" + `{{ $flag.Name }}{{ if $flag.TakesValue }}="…"{{ end }}` + "`" + ` {{ if $flag.Aliases }}(` + "`" + `{{ join $flag.Aliases "` + "`, `" + `" }}` + "`" + `) {{ end }}
{{- /**/ -}} | {{ $flag.Usage }}
{{- /**/ -}} | {{ if $flag.Default }}` + "`" + `{{ $flag.Default }}` + "`" + `{{ end }}
{{- /**/ -}} | {{ if $flag.EnvVars }}` + "`" + `{{ join $flag.EnvVars "` + "`, `" + `" }}` + "`" + `{{ else }}*none*{{ end }}
{{- /**/ -}} |
{{   end }}
{{ end }}

{{ define "command" }}
### ` + "`" + `{{ .Name }}` + "`" + ` {{ if gt .Level 0 }}sub{{ end }}command{{ if .Aliases }} (aliases: ` + "`" + `{{ join .Aliases "` + "`, `" + `" }}` + "`" + `){{ end }}
{{   if .Usage }}
{{     .Usage }}.
{{   end }}
{{   if .UsageText }}
{{     range $line := .UsageText -}}
> {{ $line }}
{{     end -}}
{{   end }}
{{   if .Description }}
{{     .Description }}.
{{   end }}
Usage:

` + "```" + `bash
$ {{ .AppPath }} [GLOBAL FLAGS] {{ .Name }}{{ if .Flags }} [COMMAND FLAGS]{{ end }} {{ if .ArgsUsage }}{{ .ArgsUsage }}{{ else }}[ARGUMENTS...]{{ end }}
` + "```" + `

{{   if .Flags -}}
The following flags are supported:
{{     template "flags" .Flags }}
{{   end -}}

{{   if .SubCommands -}}
{{     range $subCmd := .SubCommands -}}
{{       template "command" $subCmd }}
{{     end -}}
{{   end -}}
{{ end }}

## CLI interface{{ if .Name }} - {{ .Name }}{{ end }}

{{ if .Description }}{{ .Description }}.
{{ end }}
{{ if .Usage }}{{ .Usage }}.
{{ end }}
{{ if .UsageText }}
{{   range $line := .UsageText -}}
> {{ $line }}
{{   end -}}
{{ end }}
Usage:

` + "```" + `bash
$ {{ .AppPath }}{{ if .GlobalFlags }} [GLOBAL FLAGS]{{ end }} [COMMAND] [COMMAND FLAGS] {{ if .ArgsUsage }}{{ .ArgsUsage }}{{ else }}[ARGUMENTS...]{{ end }}
` + "```" + `

{{ if .GlobalFlags }}
Global flags:

{{ template "flags" .GlobalFlags }}

{{ end -}}
{{ if .Commands -}}
{{   range $cmd := .Commands -}}
{{     template "command" $cmd }}
{{   end }}
{{- end }}`

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
