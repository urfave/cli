package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
)

func Test_ShowAppHelp_NoAuthor(t *testing.T) {
	output := new(bytes.Buffer)
	app := &App{Writer: output}

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("AUTHOR(S):")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "AUTHOR(S):")
	}
}

func Test_ShowAppHelp_NoVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := &App{Writer: output}

	app.Version = ""

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("VERSION:")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "VERSION:")
	}
}

func Test_ShowAppHelp_HideVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := &App{Writer: output}

	app.HideVersion = true

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("VERSION:")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "VERSION:")
	}
}

func Test_ShowAppHelp_MultiLineDescription(t *testing.T) {
	output := new(bytes.Buffer)
	app := &App{Writer: output}

	app.HideVersion = true
	app.Description = "multi\n  line"

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if !bytes.Contains(output.Bytes(), []byte("DESCRIPTION:\n   multi\n     line")) {
		t.Errorf("expected\n%s\nto include\n%s", output.String(), "DESCRIPTION:\n   multi\n     line")
	}
}

func Test_Help_Custom_Flags(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = &BoolFlag{
		Name:    "help",
		Aliases: []string{"x"},
		Usage:   "show help",
	}

	app := App{
		Flags: []Flag{
			&BoolFlag{Name: "foo", Aliases: []string{"h"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("h") != true {
				t.Errorf("custom help flag not set")
			}
			return nil
		},
	}
	output := new(bytes.Buffer)
	app.Writer = output
	_ = app.Run([]string{"test", "-h"})
	if output.Len() > 0 {
		t.Errorf("unexpected output: %s", output.String())
	}
}

func Test_Version_Custom_Flags(t *testing.T) {
	oldFlag := VersionFlag
	defer func() {
		VersionFlag = oldFlag
	}()

	VersionFlag = &BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "show version",
	}

	app := App{
		Flags: []Flag{
			&BoolFlag{Name: "foo", Aliases: []string{"v"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("v") != true {
				t.Errorf("custom version flag not set")
			}
			return nil
		},
	}
	output := new(bytes.Buffer)
	app.Writer = output
	_ = app.Run([]string{"test", "-v"})
	if output.Len() > 0 {
		t.Errorf("unexpected output: %s", output.String())
	}
}

func Test_helpCommand_Action_ErrorIfNoTopic(t *testing.T) {
	app := &App{}

	set := flag.NewFlagSet("test", 0)
	_ = set.Parse([]string{"foo"})

	c := NewContext(app, set, nil)

	err := helpCommand.Action(c)

	if err == nil {
		t.Fatalf("expected error from helpCommand.Action(), but got nil")
	}

	exitErr, ok := err.(*exitError)
	if !ok {
		t.Fatalf("expected *exitError from helpCommand.Action(), but instead got: %v", err.Error())
	}

	if !strings.HasPrefix(exitErr.Error(), "No help topic for") {
		t.Fatalf("expected an unknown help topic error, but got: %v", exitErr.Error())
	}

	if exitErr.exitCode != 3 {
		t.Fatalf("expected exit value = 3, got %d instead", exitErr.exitCode)
	}
}

func Test_helpCommand_InHelpOutput(t *testing.T) {
	app := &App{}
	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"test", "--help"})

	s := output.String()

	if strings.Contains(s, "\nCOMMANDS:\nGLOBAL OPTIONS:\n") {
		t.Fatalf("empty COMMANDS section detected: %q", s)
	}

	if !strings.Contains(s, "help, h") {
		t.Fatalf("missing \"help, h\": %q", s)
	}
}

func Test_helpSubcommand_Action_ErrorIfNoTopic(t *testing.T) {
	app := &App{}

	set := flag.NewFlagSet("test", 0)
	_ = set.Parse([]string{"foo"})

	c := NewContext(app, set, nil)

	err := helpSubcommand.Action(c)

	if err == nil {
		t.Fatalf("expected error from helpCommand.Action(), but got nil")
	}

	exitErr, ok := err.(*exitError)
	if !ok {
		t.Fatalf("expected *exitError from helpCommand.Action(), but instead got: %v", err.Error())
	}

	if !strings.HasPrefix(exitErr.Error(), "No help topic for") {
		t.Fatalf("expected an unknown help topic error, but got: %v", exitErr.Error())
	}

	if exitErr.exitCode != 3 {
		t.Fatalf("expected exit value = 3, got %d instead", exitErr.exitCode)
	}
}

func TestShowAppHelp_CommandAliases(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name:    "frobbly",
				Aliases: []string{"fr", "frob"},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"foo", "--help"})

	if !strings.Contains(output.String(), "frobbly, fr, frob") {
		t.Errorf("expected output to include all command aliases; got: %q", output.String())
	}
}

func TestShowCommandHelp_HelpPrinter(t *testing.T) {
	doublecho := func(text string) string {
		return text + " " + text
	}

	tests := []struct {
		name         string
		template     string
		printer      helpPrinter
		command      string
		wantTemplate string
		wantOutput   string
	}{
		{
			name:     "no-command",
			template: "",
			printer: func(w io.Writer, templ string, data interface{}) {
				fmt.Fprint(w, "yo")
			},
			command:      "",
			wantTemplate: SubcommandHelpTemplate,
			wantOutput:   "yo",
		},
		{
			name:     "standard-command",
			template: "",
			printer: func(w io.Writer, templ string, data interface{}) {
				fmt.Fprint(w, "yo")
			},
			command:      "my-command",
			wantTemplate: CommandHelpTemplate,
			wantOutput:   "yo",
		},
		{
			name:     "custom-template-command",
			template: "{{doublecho .Name}}",
			printer: func(w io.Writer, templ string, data interface{}) {
				// Pass a custom function to ensure it gets used
				fm := map[string]interface{}{"doublecho": doublecho}
				HelpPrinterCustom(w, templ, data, fm)
			},
			command:      "my-command",
			wantTemplate: "{{doublecho .Name}}",
			wantOutput:   "my-command my-command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(old helpPrinter) {
				HelpPrinter = old
			}(HelpPrinter)
			HelpPrinter = func(w io.Writer, templ string, data interface{}) {
				if templ != tt.wantTemplate {
					t.Errorf("want template:\n%s\ngot template:\n%s", tt.wantTemplate, templ)
				}

				tt.printer(w, templ, data)
			}

			var buf bytes.Buffer
			app := &App{
				Name:   "my-app",
				Writer: &buf,
				Commands: []*Command{
					{
						Name:               "my-command",
						CustomHelpTemplate: tt.template,
					},
				},
			}

			err := app.Run([]string{"my-app", "help", tt.command})
			if err != nil {
				t.Fatal(err)
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("want output %q, got %q", tt.wantOutput, got)
			}
		})
	}
}

func TestShowCommandHelp_HelpPrinterCustom(t *testing.T) {
	doublecho := func(text string) string {
		return text + " " + text
	}

	tests := []struct {
		name         string
		template     string
		printer      helpPrinterCustom
		command      string
		wantTemplate string
		wantOutput   string
	}{
		{
			name:     "no-command",
			template: "",
			printer: func(w io.Writer, templ string, data interface{}, fm map[string]interface{}) {
				fmt.Fprint(w, "yo")
			},
			command:      "",
			wantTemplate: SubcommandHelpTemplate,
			wantOutput:   "yo",
		},
		{
			name:     "standard-command",
			template: "",
			printer: func(w io.Writer, templ string, data interface{}, fm map[string]interface{}) {
				fmt.Fprint(w, "yo")
			},
			command:      "my-command",
			wantTemplate: CommandHelpTemplate,
			wantOutput:   "yo",
		},
		{
			name:     "custom-template-command",
			template: "{{doublecho .Name}}",
			printer: func(w io.Writer, templ string, data interface{}, _ map[string]interface{}) {
				// Pass a custom function to ensure it gets used
				fm := map[string]interface{}{"doublecho": doublecho}
				printHelpCustom(w, templ, data, fm)
			},
			command:      "my-command",
			wantTemplate: "{{doublecho .Name}}",
			wantOutput:   "my-command my-command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(old helpPrinterCustom) {
				HelpPrinterCustom = old
			}(HelpPrinterCustom)
			HelpPrinterCustom = func(w io.Writer, templ string, data interface{}, fm map[string]interface{}) {
				if fm != nil {
					t.Error("unexpected function map passed")
				}

				if templ != tt.wantTemplate {
					t.Errorf("want template:\n%s\ngot template:\n%s", tt.wantTemplate, templ)
				}

				tt.printer(w, templ, data, fm)
			}

			var buf bytes.Buffer
			app := &App{
				Name:   "my-app",
				Writer: &buf,
				Commands: []*Command{
					{
						Name:               "my-command",
						CustomHelpTemplate: tt.template,
					},
				},
			}

			err := app.Run([]string{"my-app", "help", tt.command})
			if err != nil {
				t.Fatal(err)
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("want output %q, got %q", tt.wantOutput, got)
			}
		})
	}
}

func TestShowCommandHelp_CommandAliases(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name:    "frobbly",
				Aliases: []string{"fr", "frob", "bork"},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"foo", "help", "fr"})

	if !strings.Contains(output.String(), "frobbly") {
		t.Errorf("expected output to include command name; got: %q", output.String())
	}

	if strings.Contains(output.String(), "bork") {
		t.Errorf("expected output to exclude command aliases; got: %q", output.String())
	}
}

func TestShowSubcommandHelp_CommandAliases(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name:    "frobbly",
				Aliases: []string{"fr", "frob", "bork"},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"foo", "help"})

	if !strings.Contains(output.String(), "frobbly, fr, frob, bork") {
		t.Errorf("expected output to include all command aliases; got: %q", output.String())
	}
}

func TestShowCommandHelp_Customtemplate(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "frobbly",
				Action: func(ctx *Context) error {
					return nil
				},
				HelpName: "foo frobbly",
				CustomHelpTemplate: `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}} [FLAGS] TARGET [TARGET ...]

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
   1. Frobbly runs with this param locally.
      $ {{.HelpName}} wobbly
`,
			},
		},
	}
	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"foo", "help", "frobbly"})

	if strings.Contains(output.String(), "2. Frobbly runs without this param locally.") {
		t.Errorf("expected output to exclude \"2. Frobbly runs without this param locally.\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "1. Frobbly runs with this param locally.") {
		t.Errorf("expected output to include \"1. Frobbly runs with this param locally.\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "$ foo frobbly wobbly") {
		t.Errorf("expected output to include \"$ foo frobbly wobbly\"; got: %q", output.String())
	}
}

func TestShowSubcommandHelp_CommandUsageText(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name:      "frobbly",
				UsageText: "this is usage text",
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output

	_ = app.Run([]string{"foo", "frobbly", "--help"})

	if !strings.Contains(output.String(), "this is usage text") {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestShowSubcommandHelp_MultiLine_CommandUsageText(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "frobbly",
				UsageText: `This is a
multi
line
UsageText`,
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output

	_ = app.Run([]string{"foo", "frobbly", "--help"})

	expected := `USAGE:
   This is a
   multi
   line
   UsageText
`

	if !strings.Contains(output.String(), expected) {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestShowSubcommandHelp_SubcommandUsageText(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "frobbly",
				Subcommands: []*Command{
					{
						Name:      "bobbly",
						UsageText: "this is usage text",
					},
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"foo", "frobbly", "bobbly", "--help"})

	if !strings.Contains(output.String(), "this is usage text") {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestShowSubcommandHelp_MultiLine_SubcommandUsageText(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "frobbly",
				Subcommands: []*Command{
					{
						Name: "bobbly",
						UsageText: `This is a
multi
line
UsageText`,
					},
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"foo", "frobbly", "bobbly", "--help"})

	expected := `USAGE:
   This is a
   multi
   line
   UsageText
`

	if !strings.Contains(output.String(), expected) {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestShowAppHelp_HiddenCommand(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "frobbly",
				Action: func(ctx *Context) error {
					return nil
				},
			},
			{
				Name:   "secretfrob",
				Hidden: true,
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"app", "--help"})

	if strings.Contains(output.String(), "secretfrob") {
		t.Errorf("expected output to exclude \"secretfrob\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "frobbly") {
		t.Errorf("expected output to include \"frobbly\"; got: %q", output.String())
	}
}

func TestShowAppHelp_HelpPrinter(t *testing.T) {
	doublecho := func(text string) string {
		return text + " " + text
	}

	tests := []struct {
		name         string
		template     string
		printer      helpPrinter
		wantTemplate string
		wantOutput   string
	}{
		{
			name:     "standard-command",
			template: "",
			printer: func(w io.Writer, templ string, data interface{}) {
				fmt.Fprint(w, "yo")
			},
			wantTemplate: AppHelpTemplate,
			wantOutput:   "yo",
		},
		{
			name:     "custom-template-command",
			template: "{{doublecho .Name}}",
			printer: func(w io.Writer, templ string, data interface{}) {
				// Pass a custom function to ensure it gets used
				fm := map[string]interface{}{"doublecho": doublecho}
				printHelpCustom(w, templ, data, fm)
			},
			wantTemplate: "{{doublecho .Name}}",
			wantOutput:   "my-app my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(old helpPrinter) {
				HelpPrinter = old
			}(HelpPrinter)
			HelpPrinter = func(w io.Writer, templ string, data interface{}) {
				if templ != tt.wantTemplate {
					t.Errorf("want template:\n%s\ngot template:\n%s", tt.wantTemplate, templ)
				}

				tt.printer(w, templ, data)
			}

			var buf bytes.Buffer
			app := &App{
				Name:                  "my-app",
				Writer:                &buf,
				CustomAppHelpTemplate: tt.template,
			}

			err := app.Run([]string{"my-app", "help"})
			if err != nil {
				t.Fatal(err)
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("want output %q, got %q", tt.wantOutput, got)
			}
		})
	}
}

func TestShowAppHelp_HelpPrinterCustom(t *testing.T) {
	doublecho := func(text string) string {
		return text + " " + text
	}

	tests := []struct {
		name         string
		template     string
		printer      helpPrinterCustom
		wantTemplate string
		wantOutput   string
	}{
		{
			name:     "standard-command",
			template: "",
			printer: func(w io.Writer, templ string, data interface{}, fm map[string]interface{}) {
				fmt.Fprint(w, "yo")
			},
			wantTemplate: AppHelpTemplate,
			wantOutput:   "yo",
		},
		{
			name:     "custom-template-command",
			template: "{{doublecho .Name}}",
			printer: func(w io.Writer, templ string, data interface{}, _ map[string]interface{}) {
				// Pass a custom function to ensure it gets used
				fm := map[string]interface{}{"doublecho": doublecho}
				printHelpCustom(w, templ, data, fm)
			},
			wantTemplate: "{{doublecho .Name}}",
			wantOutput:   "my-app my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(old helpPrinterCustom) {
				HelpPrinterCustom = old
			}(HelpPrinterCustom)
			HelpPrinterCustom = func(w io.Writer, templ string, data interface{}, fm map[string]interface{}) {
				if fm != nil {
					t.Error("unexpected function map passed")
				}

				if templ != tt.wantTemplate {
					t.Errorf("want template:\n%s\ngot template:\n%s", tt.wantTemplate, templ)
				}

				tt.printer(w, templ, data, fm)
			}

			var buf bytes.Buffer
			app := &App{
				Name:                  "my-app",
				Writer:                &buf,
				CustomAppHelpTemplate: tt.template,
			}

			err := app.Run([]string{"my-app", "help"})
			if err != nil {
				t.Fatal(err)
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("want output %q, got %q", tt.wantOutput, got)
			}
		})
	}
}

func TestShowAppHelp_CustomAppTemplate(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name: "frobbly",
				Action: func(ctx *Context) error {
					return nil
				},
			},
			{
				Name:   "secretfrob",
				Hidden: true,
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
		ExtraInfo: func() map[string]string {
			platform := fmt.Sprintf("OS: %s | Arch: %s", runtime.GOOS, runtime.GOARCH)
			goruntime := fmt.Sprintf("Version: %s | CPUs: %d", runtime.Version(), runtime.NumCPU())
			return map[string]string{
				"PLATFORM": platform,
				"RUNTIME":  goruntime,
			}
		},
		CustomAppHelpTemplate: `NAME:
  {{.Name}} - {{.Usage}}

USAGE:
  {{.Name}} {{if .VisibleFlags}}[FLAGS] {{end}}COMMAND{{if .VisibleFlags}} [COMMAND FLAGS | -h]{{end}} [ARGUMENTS...]

COMMANDS:
  {{range .VisibleCommands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
  {{end}}{{if .VisibleFlags}}
GLOBAL FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
VERSION:
  2.0.0
{{"\n"}}{{range $key, $value := ExtraInfo}}
{{$key}}:
  {{$value}}
{{end}}`,
	}

	output := &bytes.Buffer{}
	app.Writer = output
	_ = app.Run([]string{"app", "--help"})

	if strings.Contains(output.String(), "secretfrob") {
		t.Errorf("expected output to exclude \"secretfrob\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "frobbly") {
		t.Errorf("expected output to include \"frobbly\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "PLATFORM:") ||
		!strings.Contains(output.String(), "OS:") ||
		!strings.Contains(output.String(), "Arch:") {
		t.Errorf("expected output to include \"PLATFORM:, OS: and Arch:\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "RUNTIME:") ||
		!strings.Contains(output.String(), "Version:") ||
		!strings.Contains(output.String(), "CPUs:") {
		t.Errorf("expected output to include \"RUNTIME:, Version: and CPUs:\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "VERSION:") ||
		!strings.Contains(output.String(), "2.0.0") {
		t.Errorf("expected output to include \"VERSION:, 2.0.0\"; got: %q", output.String())
	}
}

func TestShowAppHelp_UsageText(t *testing.T) {
	app := &App{
		UsageText: "This is a sinlge line of UsageText",
		Commands: []*Command{
			{
				Name: "frobbly",
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output

	_ = app.Run([]string{"foo"})

	if !strings.Contains(output.String(), "This is a sinlge line of UsageText") {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestShowAppHelp_MultiLine_UsageText(t *testing.T) {
	app := &App{
		UsageText: `This is a
multi
line
App UsageText`,
		Commands: []*Command{
			{
				Name: "frobbly",
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output

	_ = app.Run([]string{"foo"})

	expected := `USAGE:
   This is a
   multi
   line
   App UsageText
`

	if !strings.Contains(output.String(), expected) {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestShowAppHelp_CommandMultiLine_UsageText(t *testing.T) {
	app := &App{
		UsageText: `This is a
multi
line
App UsageText`,
		Commands: []*Command{
			{
				Name:    "frobbly",
				Aliases: []string{"frb1", "frbb2", "frl2"},
				Usage:   "this is a long help output for the run command, long usage \noutput, long usage output, long usage output, long usage output\noutput, long usage output, long usage output",
			},
			{
				Name:    "grobbly",
				Aliases: []string{"grb1", "grbb2"},
				Usage:   "this is another long help output for the run command, long usage \noutput, long usage output",
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output

	_ = app.Run([]string{"foo"})

	expected := "COMMANDS:\n" +
		"   frobbly, frb1, frbb2, frl2  this is a long help output for the run command, long usage \n" +
		"                               output, long usage output, long usage output, long usage output\n" +
		"                               output, long usage output, long usage output\n" +
		"   grobbly, grb1, grbb2        this is another long help output for the run command, long usage \n" +
		"                               output, long usage output"
	if !strings.Contains(output.String(), expected) {
		t.Errorf("expected output to include usage text; got: %q", output.String())
	}
}

func TestHideHelpCommand(t *testing.T) {
	app := &App{
		HideHelpCommand: true,
		Writer:          ioutil.Discard,
	}

	err := app.Run([]string{"foo", "help"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	if !strings.Contains(err.Error(), "No help topic for 'help'") {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	err = app.Run([]string{"foo", "--help"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHideHelpCommand_False(t *testing.T) {
	app := &App{
		HideHelpCommand: false,
		Writer:          ioutil.Discard,
	}

	err := app.Run([]string{"foo", "help"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	err = app.Run([]string{"foo", "--help"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHideHelpCommand_WithHideHelp(t *testing.T) {
	app := &App{
		HideHelp:        true, // effective (hides both command and flag)
		HideHelpCommand: true, // ignored
		Writer:          ioutil.Discard,
	}

	err := app.Run([]string{"foo", "help"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	if !strings.Contains(err.Error(), "No help topic for 'help'") {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	err = app.Run([]string{"foo", "--help"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	if !strings.Contains(err.Error(), "flag: help requested") {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func newContextFromStringSlice(ss []string) *Context {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	_ = set.Parse(ss)
	return &Context{flagSet: set}
}

func TestHideHelpCommand_RunAsSubcommand(t *testing.T) {
	app := &App{
		HideHelpCommand: true,
		Writer:          ioutil.Discard,
		Commands: []*Command{
			{
				Name: "dummy",
			},
		},
	}

	err := app.RunAsSubcommand(newContextFromStringSlice([]string{"", "help"}))
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	if !strings.Contains(err.Error(), "No help topic for 'help'") {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	err = app.RunAsSubcommand(newContextFromStringSlice([]string{"", "--help"}))
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHideHelpCommand_RunAsSubcommand_False(t *testing.T) {
	app := &App{
		HideHelpCommand: false,
		Writer:          ioutil.Discard,
		Commands: []*Command{
			{
				Name: "dummy",
			},
		},
	}

	err := app.RunAsSubcommand(newContextFromStringSlice([]string{"", "help"}))
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	err = app.RunAsSubcommand(newContextFromStringSlice([]string{"", "--help"}))
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHideHelpCommand_WithSubcommands(t *testing.T) {
	app := &App{
		Writer: ioutil.Discard,
		Commands: []*Command{
			{
				Name: "dummy",
				Subcommands: []*Command{
					{
						Name: "dummy2",
					},
				},
				HideHelpCommand: true,
			},
		},
	}

	err := app.Run([]string{"foo", "dummy", "help"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	if !strings.Contains(err.Error(), "No help topic for 'help'") {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	err = app.Run([]string{"foo", "dummy", "--help"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestDefaultCompleteWithFlags(t *testing.T) {
	origEnv := os.Environ()
	origArgv := os.Args

	t.Cleanup(func() {
		os.Args = origArgv
		resetEnv(origEnv)
	})

	os.Setenv("SHELL", "bash")

	for _, tc := range []struct {
		name     string
		c        *Context
		cmd      *Command
		argv     []string
		expected string
	}{
		{
			name:     "empty",
			c:        &Context{App: &App{}},
			cmd:      &Command{},
			argv:     []string{"prog", "cmd"},
			expected: "",
		},
		{
			name: "typical-flag-suggestion",
			c: &Context{App: &App{
				Name: "cmd",
				Flags: []Flag{
					&BoolFlag{Name: "happiness"},
					&Int64Flag{Name: "everybody-jump-on"},
				},
				Commands: []*Command{
					{Name: "putz"},
				},
			}},
			cmd: &Command{
				Flags: []Flag{
					&BoolFlag{Name: "excitement"},
					&StringFlag{Name: "hat-shape"},
				},
			},
			argv:     []string{"cmd", "--e", "--generate-bash-completion"},
			expected: "--excitement\n",
		},
		{
			name: "typical-command-suggestion",
			c: &Context{App: &App{
				Name: "cmd",
				Flags: []Flag{
					&BoolFlag{Name: "happiness"},
					&Int64Flag{Name: "everybody-jump-on"},
				},
			}},
			cmd: &Command{
				Name: "putz",
				Subcommands: []*Command{
					{Name: "futz"},
				},
				Flags: []Flag{
					&BoolFlag{Name: "excitement"},
					&StringFlag{Name: "hat-shape"},
				},
			},
			argv:     []string{"cmd", "--generate-bash-completion"},
			expected: "futz\n",
		},
	} {
		t.Run(tc.name, func(ct *testing.T) {
			writer := &bytes.Buffer{}
			tc.c.App.Writer = writer

			os.Args = tc.argv
			f := DefaultCompleteWithFlags(tc.cmd)
			f(tc.c)

			written := writer.String()

			if written != tc.expected {
				ct.Errorf("written help does not match expected %q != %q", written, tc.expected)
			}
		})
	}
}

func TestWrappedHelp(t *testing.T) {

	// Reset HelpPrinter after this test.
	defer func(old helpPrinter) {
		HelpPrinter = old
	}(HelpPrinter)

	output := new(bytes.Buffer)
	app := &App{
		Writer: output,
		Flags: []Flag{
			&BoolFlag{Name: "foo",
				Aliases: []string{"h"},
				Usage:   "here's a really long help text line, let's see where it wraps. blah blah blah and so on.",
			},
		},
		Usage:     "here's a sample App.Usage string long enough that it should be wrapped in this test",
		UsageText: "i'm not sure how App.UsageText differs from App.Usage, but this should also be wrapped in this test",
		// TODO: figure out how to make ArgsUsage appear in the help text, and test that
		Description: `here's a sample App.Description string long enough that it should be wrapped in this test

with a newline
   and an indented line`,
		Copyright: `Here's a sample copyright text string long enough that it should be wrapped.
Including newlines.
   And also indented lines.


And then another long line. Blah blah blah does anybody ever read these things?`,
	}

	c := NewContext(app, nil, nil)

	HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		funcMap := map[string]interface{}{
			"wrapAt": func() int {
				return 30
			},
		}

		HelpPrinterCustom(w, templ, data, funcMap)
	}

	_ = ShowAppHelp(c)

	expected := `NAME:
    - here's a sample
      App.Usage string long
      enough that it should be
      wrapped in this test

USAGE:
   i'm not sure how
   App.UsageText differs from
   App.Usage, but this should
   also be wrapped in this
   test

DESCRIPTION:
   here's a sample
   App.Description string long
   enough that it should be
   wrapped in this test
   
   with a newline
      and an indented line

GLOBAL OPTIONS:
   --foo, -h here's a
      really long help text
      line, let's see where it
      wraps. blah blah blah
      and so on. (default:
      false)

COPYRIGHT:
   Here's a sample copyright
   text string long enough
   that it should be wrapped.
   Including newlines.
      And also indented lines.
   
   
   And then another long line.
   Blah blah blah does anybody
   ever read these things?
`

	if output.String() != expected {
		t.Errorf("Unexpected wrapping, got:\n%s\nexpected: %s",
			output.String(), expected)
	}
}

func TestWrappedCommandHelp(t *testing.T) {

	// Reset HelpPrinter after this test.
	defer func(old helpPrinter) {
		HelpPrinter = old
	}(HelpPrinter)

	output := new(bytes.Buffer)
	app := &App{
		Writer: output,
		Commands: []*Command{
			{
				Name:        "add",
				Aliases:     []string{"a"},
				Usage:       "add a task to the list",
				UsageText:   "this is an even longer way of describing adding a task to the list",
				Description: "and a description long enough to wrap in this test case",
				Action: func(c *Context) error {
					return nil
				},
			},
		},
	}

	c := NewContext(app, nil, nil)

	HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		funcMap := map[string]interface{}{
			"wrapAt": func() int {
				return 30
			},
		}

		HelpPrinterCustom(w, templ, data, funcMap)
	}

	_ = ShowCommandHelp(c, "add")

	expected := `NAME:
    - add a task to the list

USAGE:
   this is an even longer way
   of describing adding a task
   to the list

DESCRIPTION:
   and a description long
   enough to wrap in this test
   case
`

	if output.String() != expected {
		t.Errorf("Unexpected wrapping, got:\n%s\nexpected: %s",
			output.String(), expected)
	}
}

func TestWrappedSubcommandHelp(t *testing.T) {

	// Reset HelpPrinter after this test.
	defer func(old helpPrinter) {
		HelpPrinter = old
	}(HelpPrinter)

	output := new(bytes.Buffer)
	app := &App{
		Name:   "cli.test",
		Writer: output,
		Commands: []*Command{
			{
				Name:        "bar",
				Aliases:     []string{"a"},
				Usage:       "add a task to the list",
				UsageText:   "this is an even longer way of describing adding a task to the list",
				Description: "and a description long enough to wrap in this test case",
				Action: func(c *Context) error {
					return nil
				},
				Subcommands: []*Command{
					{
						Name:      "grok",
						Usage:     "remove an existing template",
						UsageText: "longer usage text goes here, la la la, hopefully this is long enough to wrap even more",
						Action: func(c *Context) error {
							return nil
						},
					},
				},
			},
		},
	}

	HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		funcMap := map[string]interface{}{
			"wrapAt": func() int {
				return 30
			},
		}

		HelpPrinterCustom(w, templ, data, funcMap)
	}

	_ = app.Run([]string{"foo", "bar", "grok", "--help"})

	expected := `NAME:
   cli.test bar grok - remove
                       an
                       existing
                       template

USAGE:
   longer usage text goes
   here, la la la, hopefully
   this is long enough to wrap
   even more

OPTIONS:
   --help, -h  show help (default: false)
   
`

	if output.String() != expected {
		t.Errorf("Unexpected wrapping, got:\n%s\nexpected: %s",
			output.String(), expected)
	}
}
