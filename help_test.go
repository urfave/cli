package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
)

func Test_ShowAppHelp_NoAuthor(t *testing.T) {
	output := new(bytes.Buffer)
	app := NewApp()
	app.Writer = output

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("AUTHOR(S):")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "AUTHOR(S):")
	}
}

func Test_ShowAppHelp_NoVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := NewApp()
	app.Writer = output

	app.Version = ""

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("VERSION:")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "VERSION:")
	}
}

func Test_ShowAppHelp_HideVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := NewApp()
	app.Writer = output

	app.HideVersion = true

	c := NewContext(app, nil, nil)

	_ = ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("VERSION:")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "VERSION:")
	}
}

func Test_Help_Custom_Flags(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = BoolFlag{
		Name:  "help, x",
		Usage: "show help",
	}

	app := App{
		Flags: []Flag{
			BoolFlag{Name: "foo, h"},
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

	VersionFlag = BoolFlag{
		Name:  "version, V",
		Usage: "show version",
	}

	app := App{
		Flags: []Flag{
			BoolFlag{Name: "foo, v"},
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
	app := NewApp()

	set := flag.NewFlagSet("test", 0)
	_ = set.Parse([]string{"foo"})

	c := NewContext(app, set, nil)

	err := helpCommand.Action.(func(*Context) error)(c)

	if err == nil {
		t.Fatalf("expected error from helpCommand.Action(), but got nil")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected ExitError from helpCommand.Action(), but instead got: %v", err.Error())
	}

	if !strings.HasPrefix(exitErr.Error(), "No help topic for") {
		t.Fatalf("expected an unknown help topic error, but got: %v", exitErr.Error())
	}

	if exitErr.exitCode != 3 {
		t.Fatalf("expected exit value = 3, got %d instead", exitErr.exitCode)
	}
}

func Test_helpCommand_InHelpOutput(t *testing.T) {
	app := NewApp()
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
	app := NewApp()

	set := flag.NewFlagSet("test", 0)
	_ = set.Parse([]string{"foo"})

	c := NewContext(app, set, nil)

	err := helpSubcommand.Action.(func(*Context) error)(c)

	if err == nil {
		t.Fatalf("expected error from helpCommand.Action(), but got nil")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected ExitError from helpCommand.Action(), but instead got: %v", err.Error())
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
		Commands: []Command{
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
				Commands: []Command{
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
				Commands: []Command{
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
		Commands: []Command{
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
		Commands: []Command{
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
		Commands: []Command{
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
		Commands: []Command{
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

func TestShowSubcommandHelp_SubcommandUsageText(t *testing.T) {
	app := &App{
		Commands: []Command{
			{
				Name: "frobbly",
				Subcommands: []Command{
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

func TestShowAppHelp_HiddenCommand(t *testing.T) {
	app := &App{
		Commands: []Command{
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
		Commands: []Command{
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
