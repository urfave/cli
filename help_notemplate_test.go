//go:build urfave_cli_no_template

package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoTemplate_CustomHelpTemplate(t *testing.T) {
	tests := []struct {
		name string
		cmd  *Command
		args []string
		want string
	}{
		{
			name: "root",
			cmd:  &Command{Name: "foo", CustomRootCommandHelpTemplate: "{{.Name}}"},
			args: []string{"foo", "--help"},
			want: `command "foo": ` + errNoTemplate.Error(),
		},
		{
			name: "subcommand",
			cmd: &Command{Name: "foo", Commands: []*Command{
				{Name: "bar", CustomHelpTemplate: "{{.Name}}"},
			}},
			args: []string{"foo", "bar", "--help"},
			want: `command "foo bar": ` + errNoTemplate.Error(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cmd.Writer = io.Discard
			assertPanicsNoTemplate(t, tc.want, func() {
				_ = tc.cmd.Run(context.Background(), tc.args)
			})
		})
	}
}

// A custom template is fine as long as it is handled by a custom help
// printer, or never used.
func TestNoTemplate_CustomHelpPrinter(t *testing.T) {
	defer func(old HelpPrinterFunc) { HelpPrinter = old }(HelpPrinter)
	HelpPrinter = func(w io.Writer, _ string, data any) {
		_, _ = io.WriteString(w, "custom help for "+data.(*Command).Name+"\n")
	}

	for _, args := range [][]string{{"foo"}, {"foo", "--help"}} {
		var out, errOut bytes.Buffer
		ran := false
		cmd := &Command{
			Name:                          "foo",
			Writer:                        &out,
			ErrWriter:                     &errOut,
			CustomRootCommandHelpTemplate: "rendered by the custom printer",
			Action: func(context.Context, *Command) error {
				ran = true
				return nil
			},
		}
		require.NoError(t, cmd.Run(context.Background(), args))
		assert.Empty(t, errOut.String())
		if len(args) == 1 {
			assert.True(t, ran)
		} else {
			assert.Equal(t, "custom help for foo\n", out.String())
		}
	}
}

// A custom help printer receives "root", "command" or "subcommand" as templ,
// and can tell commands with the same name apart using data.
func TestNoTemplate_CustomHelpPrinterTempl(t *testing.T) {
	defer func(old HelpPrinterFunc) { HelpPrinter = old }(HelpPrinter)
	HelpPrinter = func(w io.Writer, templ string, data any) {
		_, _ = io.WriteString(w, templ+": "+data.(*Command).FullName()+"\n")
	}

	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"app", "--help"}, want: "root: app"},
		{args: []string{"app", "a", "--help"}, want: "subcommand: app a"},
		{args: []string{"app", "a", "x", "--help"}, want: "command: app a x"},
		{args: []string{"app", "b", "x", "--help"}, want: "command: app b x"},
		{args: []string{"app", "b", "help", "x"}, want: "command: app b x"},
	}
	for _, tc := range tests {
		t.Run(strings.Join(tc.args[1:], " "), func(t *testing.T) {
			var out bytes.Buffer
			cmd := &Command{
				Name:   "app",
				Writer: &out,
				Commands: []*Command{
					{Name: "a", Commands: []*Command{{Name: "x"}}},
					{Name: "b", Commands: []*Command{{Name: "x"}}},
				},
			}
			require.NoError(t, cmd.Run(context.Background(), tc.args))
			assert.Equal(t, tc.want+"\n", out.String())
		})
	}
}

func TestNoTemplate_DefaultPrintHelpCustom(t *testing.T) {
	tests := []struct {
		name  string
		templ string
		data  any
		want  string
	}{
		{
			name:  "custom template",
			templ: "{{.Name}}",
			data:  &Command{Name: "foo"},
			want:  `command "foo": ` + errNoTemplate.Error(),
		},
		{
			name:  "not a command",
			templ: RootCommandHelpTemplate,
			data:  "not a command",
			want:  errNoTemplate.Error(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertPanicsNoTemplate(t, tc.want, func() {
				DefaultPrintHelpCustom(io.Discard, tc.templ, tc.data, nil)
			})
		})
	}
}

// assertPanicsNoTemplate checks that f panics with errNoTemplate,
// and the panic message is want.
func assertPanicsNoTemplate(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		err, _ := recover().(error)
		require.ErrorIs(t, err, errNoTemplate)
		assert.Equal(t, want, err.Error())
	}()
	f()
}
