package cli

import (
	"bytes"
	"strings"
	"testing"
)

func Test_ShowAppHelp_NoAuthor(t *testing.T) {
	output := new(bytes.Buffer)
	app := NewApp()
	app.Writer = output

	c := NewContext(app, nil, nil)

	ShowAppHelp(c)

	if bytes.Index(output.Bytes(), []byte("AUTHOR(S):")) != -1 {
		t.Errorf("expected\n%snot to include %s", output.String(), "AUTHOR(S):")
	}
}

func Test_ShowAppHelp_NoVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := NewApp()
	app.Writer = output

	app.Version = ""

	c := NewContext(app, nil, nil)

	ShowAppHelp(c)

	if bytes.Index(output.Bytes(), []byte("VERSION:")) != -1 {
		t.Errorf("expected\n%snot to include %s", output.String(), "VERSION:")
	}
}

func Test_ShowAppHelp_HideVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := NewApp()
	app.Writer = output

	app.HideVersion = true

	c := NewContext(app, nil, nil)

	ShowAppHelp(c)

	if bytes.Index(output.Bytes(), []byte("VERSION:")) != -1 {
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
	app.Run([]string{"test", "-h"})
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
	app.Run([]string{"test", "-v"})
	if output.Len() > 0 {
		t.Errorf("unexpected output: %s", output.String())
	}
}

func TestShowAppHelp_HiddenCommand(t *testing.T) {
	app := &App{
		Commands: []Command{
			Command{
				Name: "frobbly",
				Action: func(ctx *Context) error {
					return nil
				},
			},
			Command{
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
	app.Run([]string{"app", "--help"})

	if strings.Contains(output.String(), "secretfrob") {
		t.Fatalf("expected output to exclude \"secretfrob\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "frobbly") {
		t.Fatalf("expected output to include \"frobbly\"; got: %q", output.String())
	}
}
