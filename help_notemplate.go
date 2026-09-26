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

// With urfave_cli_no_template tag, templates are not used, and these are
// only used to tell which help is requested.
const (
	// RootCommandHelpTemplate identifies the help for the root command.
	RootCommandHelpTemplate = "root"
	// CommandHelpTemplate identifies the help for a command.
	CommandHelpTemplate = "command"
	// SubcommandHelpTemplate identifies the help for a command with subcommands.
	SubcommandHelpTemplate = "subcommand"
	// FishCompletionTemplate is not used.
	FishCompletionTemplate = "fish"
)

var errNoTemplate = errors.New("custom templates are not supported when built with urfave_cli_no_template tag")

// DefaultPrintHelpCustom is the default implementation of HelpPrinterCustom.
//
// The customFuncs map will be combined with a default template.FuncMap to
// allow using arbitrary functions in template rendering.
//
// When built with the urfave_cli_no_template tag, text/template is not used
// (as it disables linker's dead code elimination), and the help is rendered
// directly. In this mode, only the default templates are supported (it panics
// otherwise), and the only custom functions used are "wrap" and "wrapAt".
func DefaultPrintHelpCustom(out io.Writer, templ string, data any, customFuncs map[string]any) {
	const maxLineLength = 10000

	cmd, ok := data.(*Command)
	if !ok {
		panic(errNoTemplate)
	}

	h := helpRenderer{
		cmd:  cmd,
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
	case RootCommandHelpTemplate:
		h.root()
	case CommandHelpTemplate:
		h.command()
	case SubcommandHelpTemplate:
		h.subcommand()
	default:
		panic(errNoTemplate)
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	_, err := io.WriteString(w, h.String())
	handleTemplateError(err)
	_ = w.Flush()
}

// helpRenderer renders the default help templates (see template.go)
// without using text/template.
type helpRenderer struct {
	strings.Builder
	cmd  *Command
	wrap func(string, int) string
}

func (h *helpRenderer) root() {
	c := h.cmd
	h.WriteString("NAME:\n   ")
	h.helpName()
	h.WriteString("\n\nUSAGE:\n   ")
	if c.UsageText != "" {
		h.WriteString(h.wrap(c.UsageText, 3))
	} else {
		h.WriteString(c.FullName() + " ")
		if len(c.VisibleFlags()) > 0 {
			h.WriteString("[global options]")
		}
		if len(c.VisibleCommands()) > 0 {
			h.WriteString(" [command [command options]]")
		}
		h.argsUsage()
	}
	if c.Version != "" && !c.HideVersion {
		h.WriteString("\n\nVERSION:\n   " + c.Version)
	}
	h.description()
	if len(c.Authors) > 0 {
		h.WriteString("\n\nAUTHOR")
		h.authors()
	}
	if len(c.VisibleCommands()) > 0 {
		h.WriteString("\n\nCOMMANDS:")
		h.visibleCommandCategories()
	}
	h.flags("GLOBAL OPTIONS:")
	if c.Copyright != "" {
		h.WriteString("\n\nCOPYRIGHT:\n   " + h.wrap(c.Copyright, 3))
	}
	h.WriteString("\n")
}

func (h *helpRenderer) command() {
	h.WriteString("NAME:\n   ")
	h.helpName()
	h.WriteString("\n\nUSAGE:\n   ")
	h.usage()
	h.category()
	h.description()
	h.flags("OPTIONS:")
	h.persistentFlags()
	h.WriteString("\n")
}

func (h *helpRenderer) subcommand() {
	c := h.cmd
	h.WriteString("NAME:\n   ")
	h.helpName()
	h.WriteString("\n\nUSAGE:\n   ")
	if c.UsageText != "" {
		h.WriteString(h.wrap(c.UsageText, 3))
	} else {
		h.WriteString(c.FullName())
		if len(c.VisibleCommands()) > 0 {
			h.WriteString(" [command [command options]]")
		}
		h.argsUsage()
	}
	h.category()
	h.description()
	if cmds := c.VisibleCommands(); len(cmds) > 0 {
		h.WriteString("\n\nCOMMANDS:")
		h.visibleCommands(cmds)
	}
	h.flags("OPTIONS:")
	h.persistentFlags()
	h.WriteString("\n")
}

// helpName implements helpNameTemplate.
func (h *helpRenderer) helpName() {
	name := h.cmd.FullName()
	h.WriteString(h.wrap(name, 3))
	if h.cmd.Usage != "" {
		h.WriteString(" - " + h.wrap(h.cmd.Usage, offset(name, 6)))
	}
}

// usage implements usageTemplate.
func (h *helpRenderer) usage() {
	c := h.cmd
	if c.UsageText != "" {
		h.WriteString(h.wrap(c.UsageText, 3))
		return
	}
	h.WriteString(c.FullName())
	if len(c.VisibleFlags()) > 0 {
		h.WriteString(" [options]")
	}
	if len(c.VisibleCommands()) > 0 {
		h.WriteString(" [command [command options]]")
	}
	if c.ArgsUsage != "" {
		h.WriteString(" " + c.ArgsUsage)
	} else if len(c.Arguments) > 0 {
		// argsTemplate.
		h.WriteString(" ")
		for _, arg := range c.Arguments {
			h.WriteString(arg.Usage() + " ")
		}
	}
}

// argsUsage implements the part of root and subcommand usage
// that follows the command name and options.
func (h *helpRenderer) argsUsage() {
	if h.cmd.ArgsUsage != "" {
		h.WriteString(" " + h.cmd.ArgsUsage)
	} else if len(h.cmd.Arguments) > 0 {
		h.WriteString(" [arguments...]")
	}
}

func (h *helpRenderer) category() {
	if h.cmd.Category != "" {
		h.WriteString("\n\nCATEGORY:\n   " + h.cmd.Category)
	}
}

// description implements descriptionTemplate.
func (h *helpRenderer) description() {
	if h.cmd.Description != "" {
		h.WriteString("\n\nDESCRIPTION:\n   " + h.wrap(h.cmd.Description, 3))
	}
}

// authors implements authorsTemplate.
func (h *helpRenderer) authors() {
	if len(h.cmd.Authors) != 1 {
		h.WriteString("S")
	}
	h.WriteString(":\n   ")
	for i, author := range h.cmd.Authors {
		if i > 0 {
			h.WriteString("\n   ")
		}
		h.WriteString(printableValue(author))
	}
}

// visibleCommandCategories implements visibleCommandCategoryTemplate.
func (h *helpRenderer) visibleCommandCategories() {
	for _, cat := range h.cmd.VisibleCategories() {
		if cat.Name() == "" {
			h.visibleCommands(cat.VisibleCommands())
			continue
		}
		h.WriteString("\n\n   " + cat.Name() + ":")
		for _, cmd := range cat.VisibleCommands() {
			h.WriteString("\n     " + strings.Join(cmd.Names(), ", ") + "\t" + cmd.Usage)
		}
	}
}

// visibleCommands implements visibleCommandTemplate.
func (h *helpRenderer) visibleCommands(cmds []*Command) {
	cv := offsetCommands(cmds, 5)
	for _, cmd := range cmds {
		s := strings.Join(cmd.Names(), ", ")
		h.WriteString("\n   " + s + indent(subtract(cv, offset(s, 3)), "") + h.wrap(cmd.Usage, cv))
	}
}

// flags implements the OPTIONS (or GLOBAL OPTIONS) section,
// using visibleFlagCategoryTemplate or visibleFlagTemplate.
func (h *helpRenderer) flags(title string) {
	if cats := h.cmd.VisibleFlagCategories(); len(cats) > 0 {
		h.WriteString("\n\n" + title)
		for _, cat := range cats {
			h.WriteString("\n   ")
			if cat.Name() != "" {
				h.WriteString(cat.Name() + "\n\n   ")
			}
			flags := cat.Flags()
			for i, f := range flags {
				h.WriteString(printableValue(f))
				if i == len(flags)-1 {
					h.WriteString("\n")
				} else {
					h.WriteString("\n   ")
				}
			}
		}
	} else if flags := h.cmd.VisibleFlags(); len(flags) > 0 {
		h.WriteString("\n\n" + title)
		h.visibleFlags(flags)
	}
}

// persistentFlags implements the GLOBAL OPTIONS section
// of command and subcommand help.
func (h *helpRenderer) persistentFlags() {
	if flags := h.cmd.VisiblePersistentFlags(); len(flags) > 0 {
		h.WriteString("\n\nGLOBAL OPTIONS:")
		h.visibleFlags(flags)
	}
}

// visibleFlags implements visibleFlagTemplate
// and visiblePersistentFlagTemplate.
func (h *helpRenderer) visibleFlags(flags []Flag) {
	for _, f := range flags {
		h.WriteString("\n   " + h.wrap(f.String(), 6))
	}
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
	name := data.Command.Name
	var b strings.Builder
	b.WriteString("# " + name + " fish shell completion\n\n" +
		"function __fish_" + name + "_no_subcommand --description 'Test if there has been any subcommand yet'\n" +
		"    for i in (commandline -opc)\n" +
		"        if contains -- $i")
	for _, c := range data.AllCommands {
		b.WriteString(" " + c)
	}
	b.WriteString("\n" +
		"            return 1\n" +
		"        end\n" +
		"    end\n" +
		"    return 0\n" +
		"end\n\n")
	for _, c := range data.Completions {
		b.WriteString(c + "\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}
