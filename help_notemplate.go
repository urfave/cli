//go:build urfave_cli_no_template

package cli

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

// templatesSupported is true unless built with the urfave_cli_no_template tag.
const templatesSupported = false

// Default template values, used to recognize which help is being requested.
// Package-level variable initialization happens before any init function,
// so these are the values before any modifications by the user.
var (
	defaultRootCommandHelpTemplate = RootCommandHelpTemplate
	defaultCommandHelpTemplate     = CommandHelpTemplate
	defaultSubcommandHelpTemplate  = SubcommandHelpTemplate
	defaultFishCompletionTemplate  = FishCompletionTemplate
)

var errNoTemplate = errors.New("custom templates are not supported when built with urfave_cli_no_template tag")

// DefaultPrintHelpCustom is the default implementation of HelpPrinterCustom.
//
// The customFuncs map will be combined with a default template.FuncMap to
// allow using arbitrary functions in template rendering.
//
// When built with the urfave_cli_no_template tag, text/template is not used
// (as it disables linker's dead code elimination), and the help is rendered
// directly. In this mode, only the default templates are supported, and the
// only custom functions used are "wrap" and "wrapAt".
func DefaultPrintHelpCustom(out io.Writer, templ string, data any, customFuncs map[string]any) {
	const maxLineLength = 10000

	cmd, ok := data.(*Command)
	if !ok {
		handleTemplateError(errNoTemplate)
		return
	}

	h := helpRenderer{
		wrap: func(input string, offset int) string { return wrap(input, offset, maxLineLength) },
	}
	if f, ok := customFuncs["wrap"].(func(string, int) string); ok {
		h.wrap = f
	}
	if wa, ok := customFuncs["wrapAt"]; ok {
		if wrapAtFunc, ok := wa.(func() int); ok {
			wrapAt := wrapAtFunc()
			h.wrap = func(input string, offset int) string {
				return wrap(input, offset, wrapAt)
			}
		}
	}

	switch templ {
	case defaultRootCommandHelpTemplate:
		renderRootCommandHelpTemplate(&h, cmd)
	case defaultCommandHelpTemplate:
		renderCommandHelpTemplate(&h, cmd)
	case defaultSubcommandHelpTemplate:
		renderSubcommandHelpTemplate(&h, cmd)
	default:
		handleTemplateError(errNoTemplate)
		return
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	_, err := io.WriteString(w, h.String())
	handleTemplateError(err)
	_ = w.Flush()
}

// helpRenderer is used by the code generated from the default templates
// (see help_notemplate_gen.go).
type helpRenderer struct {
	strings.Builder
	wrap func(string, int) string
}

// printableValue formats v the same way text/template prints a value.
func printableValue(v any) string {
	if v == nil {
		return "<no value>"
	}
	switch v.(type) {
	case error, fmt.Stringer:
		return fmt.Sprint(v)
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer && !rv.IsNil() {
		rv = rv.Elem()
	}
	return fmt.Sprint(rv.Interface())
}

func renderFishCompletion(w io.Writer, data *fishCommandCompletionTemplate) error {
	if FishCompletionTemplate != defaultFishCompletionTemplate {
		return errNoTemplate
	}

	var h helpRenderer
	renderFishCompletionTemplate(&h, data)

	_, err := io.WriteString(w, h.String())
	return err
}
