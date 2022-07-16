package cli

// AppHelpTemplate is the text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var AppHelpTemplate = `NAME:
   {{$v := offset .Name 6}}{{wrap .Name 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{wrap .Description 3}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlagCategories}}

GLOBAL OPTIONS:{{range .VisibleFlagCategories}}
   {{if .Name}}{{.Name}}
   {{end}}{{range .Flags}}{{.}}
   {{end}}{{end}}{{else}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{wrap $option.String 6}}{{end}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{wrap .Copyright 3}}{{end}}
`

// CommandHelpTemplate is the text template for the command help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CommandHelpTemplate = `NAME:
   {{$v := offset .HelpName 6}}{{wrap .HelpName 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{wrap .Description 3}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{range .VisibleFlagCategories}}
   {{if .Name}}{{.Name}}
   {{end}}{{range .Flags}}{{.}}
   {{end}}{{end}}{{else}}{{if .VisibleFlags}}

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{end}}
`

// SubcommandHelpTemplate is the text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{wrap .Description 3}}{{end}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

var MarkdownDocTemplate = `{{if gt .SectionNum 0}}% {{ .App.Name }} {{ .SectionNum }}

{{end}}# NAME

{{ .App.Name }}{{ if .App.Usage }} - {{ .App.Usage }}{{ end }}

# SYNOPSIS

{{ .App.Name }}
{{ if .SynopsisArgs }}
` + "```" + `
{{ range $v := .SynopsisArgs }}{{ $v }}{{ end }}` + "```" + `
{{ end }}{{ if .App.Description }}
# DESCRIPTION

{{ .App.Description }}
{{ end }}
**Usage**:

` + "```" + `{{ if .App.UsageText }}
{{ .App.UsageText }}
{{ else }}
{{ .App.Name }} [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
{{ end }}` + "```" + `
{{ if .GlobalArgs }}
# GLOBAL OPTIONS
{{ range $v := .GlobalArgs }}
{{ $v }}{{ end }}
{{ end }}{{ if .Commands }}
# COMMANDS
{{ range $v := .Commands }}
{{ $v }}{{ end }}{{ end }}`

var FishCompletionTemplate = `# {{ .App.Name }} fish shell completion

function __fish_{{ .App.Name }}_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i{{ range $v := .AllCommands }} {{ $v }}{{ end }}
            return 1
        end
    end
    return 0
end

{{ range $v := .Completions }}{{ $v }}
{{ end }}`
