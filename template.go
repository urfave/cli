package cli

var (
	helpNameTemplate    = `{{$v := offset .FullName 6}}{{wrap .FullName 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}`
	argsTemplate        = `{{if .Arguments}}{{range .Arguments}}{{.Usage}} {{end}}{{end}}`
	usageTemplate       = `{{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleFlags}} [options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} {{template "argsTemplate" .}}{{end}}{{end}}{{end}}`
	descriptionTemplate = `{{wrap .Description 3}}`
	authorsTemplate     = `{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}`
)

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

var visiblePersistentFlagTemplate = `{{range $i, $e := .VisiblePersistentFlags}}
   {{wrap $e.String 6}}{{end}}`

var versionTemplate = `{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}`

var copyrightTemplate = `{{wrap .Copyright 3}}`
