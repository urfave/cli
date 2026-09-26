//go:build urfave_cli_no_template

package cli

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoTemplate_CustomHelpTemplate(t *testing.T) {
	tests := []struct {
		name string
		cmd  *Command
		args []string
	}{
		{
			name: "root",
			cmd:  &Command{Name: "foo", CustomRootCommandHelpTemplate: "{{.Name}}"},
			args: []string{"foo", "--help"},
		},
		{
			name: "subcommand",
			cmd: &Command{Name: "foo", Commands: []*Command{
				{Name: "bar", CustomHelpTemplate: "{{.Name}}"},
			}},
			args: []string{"foo", "bar", "--help"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cmd.Writer = io.Discard
			assert.PanicsWithError(t, errNoTemplate.Error(), func() {
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

func TestNoTemplate_DefaultPrintHelpCustom(t *testing.T) {
	tests := []struct {
		name  string
		templ string
		data  any
	}{
		{name: "custom template", templ: "{{.Name}}", data: &Command{Name: "foo"}},
		{name: "not a command", templ: RootCommandHelpTemplate, data: "not a command"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.PanicsWithError(t, errNoTemplate.Error(), func() {
				DefaultPrintHelpCustom(io.Discard, tc.templ, tc.data, nil)
			})
		})
	}
}
