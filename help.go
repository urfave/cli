package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode/utf8"
)

const (
	helpName  = "help"
	helpAlias = "h"
)

// HelpPrinterFunc prints help for the Command.
type HelpPrinterFunc func(w io.Writer, templ string, data any)

// Prints help for the Command with custom template function.
type HelpPrinterCustomFunc func(w io.Writer, templ string, data any, customFunc map[string]any)

// HelpPrinter is a function that writes the help output. If not set explicitly,
// this calls HelpPrinterCustom using only the default template functions.
//
// If custom logic for printing help is required, this function can be
// overridden. If the ExtraInfo field is defined on a Command, this function
// should not be modified, as HelpPrinterCustom will be used directly in order
// to capture the extra information.
var HelpPrinter HelpPrinterFunc = DefaultPrintHelp

// HelpPrinterCustom is a function that writes the help output. It is used as
// the default implementation of HelpPrinter, and may be called directly if
// the ExtraInfo field is set on a Command.
//
// In the default implementation, if the customFuncs argument contains a
// "wrapAt" key, which is a function which takes no arguments and returns
// an int, this int value will be used to produce a "wrap" function used
// by the default template to wrap long lines.
var HelpPrinterCustom HelpPrinterCustomFunc = DefaultPrintHelpCustom

// VersionPrinter prints the version for the root Command.
var VersionPrinter = DefaultPrintVersion

// ShowRootCommandHelp is an action that displays help for the root command.
var ShowRootCommandHelp = DefaultShowRootCommandHelp

// ShowAppHelp is a backward-compatible name for ShowRootCommandHelp.
var ShowAppHelp = ShowRootCommandHelp

// ShowCommandHelp prints help for the given command
var ShowCommandHelp = DefaultShowCommandHelp

// ShowSubcommandHelp prints help for the given subcommand
var ShowSubcommandHelp = DefaultShowSubcommandHelp

// UsageCommandHelp is the text to override the USAGE section of the help command
var UsageCommandHelp = "Shows a list of commands or help for one command"

// ArgsUsageCommandHelp is a short description of the arguments of the help command
var ArgsUsageCommandHelp = "[command]"

func buildHelpCommand(withAction bool) *Command {
	cmd := &Command{
		Name:        helpName,
		Aliases:     []string{helpAlias},
		Usage:       UsageCommandHelp,
		ArgsUsage:   ArgsUsageCommandHelp,
		HideHelp:    true,
		builtInHelp: true,
	}

	if withAction {
		cmd.Action = helpCommandAction
	}

	return cmd
}

func helpCommandAction(ctx context.Context, cmd *Command) error {
	args := cmd.Args()
	firstArg := args.First()

	tracef("doing help for cmd %[1]q with args %[2]q", cmd, args)

	// helpCommandAction is triggered in several ways:
	//
	//   * the command has no user-defined Action (default action fallback)
	//   * the --help / -h flag was parsed (via cmd.checkHelp())
	//   * the "help" subcommand (or "h" alias) was dispatched
	//
	// Possible invocations:
	//
	//   $ app                  # default action; show root help
	//   $ app --help / -h      # flag; show root help (ignores subsequent args)
	//   $ app help / h         # subcommand; show root help
	//   $ app help / h foo     # subcommand; show help for subcommand "foo"
	//   $ app --help / -h foo  # flag; show help for subcommand "foo"
	//   $ app foo --help / -h  # flag on subcommand; show help for "foo"
	//   $ app foo help / h     # subcommand on subcommand; show help for "foo"
	//   $ app foo (no action)  # default action on subcommand; show help for "foo"

	// Case 4. when executing a help command set the context to parent
	// to allow resolution of subsequent args. This will transform
	// $ app help foo
	//     to
	// $ app foo
	// which will then be handled as case 3
	if cmd.parent != nil && cmd.builtInHelp {
		tracef("setting cmd to cmd.parent")
		cmd = cmd.parent
	}

	// Case 4. $ app help foo
	// foo is the command for which help needs to be shown
	if firstArg != "" {
		/*	if firstArg == "--" {
			return nil
		}*/
		tracef("returning ShowCommandHelp with %[1]q", firstArg)
		return ShowCommandHelp(ctx, cmd, firstArg)
	}

	// Case 1 & 2
	// Special case when running help on main app itself as opposed to individual
	// commands/subcommands
	if cmd.parent == nil {
		tracef("returning ShowRootCommandHelp")
		_ = ShowRootCommandHelp(cmd)
		return nil
	}

	// Case 3, 5
	if len(cmd.VisibleCommands()) == 0 {
		tracef("running HelpPrinter with command %[1]q", cmd.Name)
		return ShowCommandHelp(ctx, cmd.parent, cmd.Name)
	}

	tracef("running ShowSubcommandHelp")
	return ShowSubcommandHelp(cmd)
}

// ShowRootCommandHelpAndExit prints the list of subcommands and exits with exit code.
func ShowRootCommandHelpAndExit(cmd *Command, exitCode int) {
	_ = ShowRootCommandHelp(cmd)
	OsExiter(exitCode)
}

// ShowAppHelpAndExit is a backward-compatible name for ShowRootCommandHelp.
var ShowAppHelpAndExit = ShowRootCommandHelpAndExit

// DefaultShowRootCommandHelp is the default implementation of ShowRootCommandHelp.
func DefaultShowRootCommandHelp(cmd *Command) error {
	tmpl := cmd.CustomRootCommandHelpTemplate
	if tmpl == "" {
		tracef("using RootCommandHelpTemplate")
		tmpl = RootCommandHelpTemplate
	}

	if cmd.ExtraInfo == nil {
		HelpPrinter(cmd.Root().Writer, tmpl, cmd.Root())
		return nil
	}

	tracef("setting ExtraInfo in customAppData")
	customAppData := func() map[string]any {
		return map[string]any{
			"ExtraInfo": cmd.ExtraInfo,
		}
	}
	HelpPrinterCustom(cmd.Root().Writer, tmpl, cmd.Root(), customAppData())

	return nil
}

// DefaultRootCommandComplete prints the list of subcommands as the default completion method.
func DefaultRootCommandComplete(ctx context.Context, cmd *Command) {
	DefaultCompleteWithFlags(ctx, cmd)
}

// DefaultAppComplete is a backward-compatible name for DefaultRootCommandComplete.
var DefaultAppComplete = DefaultRootCommandComplete

func printCommandSuggestions(commands []*Command, writer io.Writer) {
	for _, command := range commands {
		if command.Hidden {
			continue
		}
		if len(command.Usage) > 0 {
			_, _ = fmt.Fprintf(writer, "%s:%s\n", command.Name, command.Usage)
		} else {
			_, _ = fmt.Fprintf(writer, "%s\n", command.Name)
		}
	}
}

func cliArgContains(flagName string, args []string) bool {
	for _, name := range strings.Split(flagName, ",") {
		name = strings.TrimSpace(name)
		count := utf8.RuneCountInString(name)
		if count > 2 {
			count = 2
		}
		flag := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)
		if slices.Contains(args, flag) {
			return true
		}
	}
	return false
}

func printFlagSuggestions(lastArg string, flags []Flag, writer io.Writer) {
	// Trim to handle both "-short" and "--long" flags.
	cur := strings.TrimLeft(lastArg, "-")
	for _, flag := range flags {
		if bflag, ok := flag.(*BoolFlag); ok && bflag.Hidden {
			continue
		}

		usage := ""
		if docFlag, ok := flag.(DocGenerationFlag); ok {
			usage = docFlag.GetUsage()
		}

		name := strings.TrimSpace(flag.Names()[0])
		// this will get total count utf8 letters in flag name
		count := utf8.RuneCountInString(name)
		if count > 2 {
			count = 2 // reuse this count to generate single - or -- in flag completion
		}
		// if flag name has more than one utf8 letter and last argument in cli has -- prefix then
		// skip flag completion for short flags example -v or -x
		if strings.HasPrefix(lastArg, "--") && count == 1 {
			continue
		}
		// match if last argument matches this flag and it is not repeated
		if strings.HasPrefix(name, cur) && cur != name /* && !cliArgContains(name, os.Args)*/ {
			flagCompletion := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)
			if usage != "" {
				flagCompletion = fmt.Sprintf("%s:%s", flagCompletion, usage)
			}
			fmt.Fprintln(writer, flagCompletion)
		}
	}
}

func DefaultCompleteWithFlags(ctx context.Context, cmd *Command) {
	args := os.Args
	if cmd != nil && cmd.parent != nil {
		args = cmd.Args().Slice()
		tracef("running default complete with flags[%v] on command %[2]q", args, cmd.Name)
	} else {
		tracef("running default complete with os.Args flags[%v]", args)
	}

	if cmd == nil {
		return
	}

	req := cmd.Root().completion

	// Everything after "--" is a positional argument of whatever the command runs, so
	// this command's flags and subcommands are no answer to it.
	// https://unix.stackexchange.com/a/11382
	if req != nil && req.terminated {
		tracef("not suggesting past a \"--\" on command %[1]q", cmd.Name)
		return
	}

	lastArg := ""
	if req != nil && req.wordKnown {
		// The request says which word is being completed, so there is nothing to work
		// out from the position of the arguments.
		lastArg = req.word
	} else if argsLen := len(args); argsLen > 1 {
		// A request in the deprecated form leaves the word out unless it starts with
		// "-", and the parent command still has completionFlag on it, so the word is
		// looked for one before the end.
		lastArg = args[argsLen-2]
	} else if argsLen > 0 {
		lastArg = args[argsLen-1]
	}

	if lastArg == completionFlag {
		lastArg = ""
	}

	if strings.HasPrefix(lastArg, "-") {
		tracef("printing flag suggestion for flag[%v] on command %[1]q", lastArg, cmd.Name)
		printFlagSuggestions(lastArg, cmd.Flags, cmd.Root().Writer)
		return
	}

	tracef("printing command suggestions on command %[1]q", cmd.Name)
	printCommandSuggestions(cmd.Commands, cmd.Root().Writer)
}

// ShowCommandHelpAndExit exits with code after showing help via ShowCommandHelp.
func ShowCommandHelpAndExit(ctx context.Context, cmd *Command, command string, code int) {
	_ = ShowCommandHelp(ctx, cmd, command)
	OsExiter(code)
}

// DefaultShowCommandHelp is the default implementation of ShowCommandHelp.
func DefaultShowCommandHelp(ctx context.Context, cmd *Command, commandName string) error {
	for _, subCmd := range cmd.Commands {
		if !subCmd.HasName(commandName) {
			continue
		}

		tmpl := subCmd.CustomHelpTemplate
		if tmpl == "" {
			if len(subCmd.VisibleCommands()) == 0 {
				tracef("using CommandHelpTemplate")
				tmpl = CommandHelpTemplate
			} else {
				tracef("using SubcommandHelpTemplate")
				tmpl = SubcommandHelpTemplate
			}
		}

		tracef("running HelpPrinter")
		HelpPrinter(cmd.Root().Writer, tmpl, subCmd)

		tracef("returning nil after printing help")
		return nil
	}

	tracef("no matching command found")

	if cmd.CommandNotFound == nil {
		errMsg := fmt.Sprintf("No help topic for '%v'", commandName)

		if cmd.Suggest {
			if suggestion := SuggestCommand(cmd.Commands, commandName); suggestion != "" {
				errMsg += ". " + suggestion
			}
		}

		tracef("exiting 3 with errMsg %[1]q", errMsg)
		return Exit(errMsg, 3)
	}

	tracef("running CommandNotFound func for %[1]q", commandName)
	cmd.CommandNotFound(ctx, cmd, commandName)

	return nil
}

// ShowSubcommandHelpAndExit prints help for the given subcommand via ShowSubcommandHelp and exits with exit code.
func ShowSubcommandHelpAndExit(cmd *Command, exitCode int) {
	_ = ShowSubcommandHelp(cmd)
	OsExiter(exitCode)
}

// DefaultShowSubcommandHelp is the default implementation of ShowSubcommandHelp.
func DefaultShowSubcommandHelp(cmd *Command) error {
	HelpPrinter(cmd.Root().Writer, SubcommandHelpTemplate, cmd)
	return nil
}

// ShowVersion prints the version number of the root Command.
func ShowVersion(cmd *Command) {
	tracef("showing version via VersionPrinter (cmd=%[1]q)", cmd.Name)
	VersionPrinter(cmd)
}

// DefaultPrintVersion is the default implementation of VersionPrinter.
func DefaultPrintVersion(cmd *Command) {
	_, _ = fmt.Fprintf(cmd.Root().Writer, "%v version %v\n", cmd.Name, cmd.Version)
}

func handleTemplateError(err error) {
	if err != nil {
		tracef("error encountered during template parse: %[1]v", err)
		// If the writer is closed, t.Execute will fail, and there's nothing
		// we can do to recover.
		if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
			_, _ = fmt.Fprintf(ErrWriter, "CLI TEMPLATE ERROR: %#v\n", err)
		}
		return
	}
}

// DefaultPrintHelpCustom is the default implementation of HelpPrinterCustom.
//
// The customFuncs map will be combined with a default template.FuncMap to
// allow using arbitrary functions in template rendering.
func DefaultPrintHelpCustom(out io.Writer, templ string, data any, customFuncs map[string]any) {
	const maxLineLength = 10000

	tracef("building default funcMap")
	funcMap := template.FuncMap{
		"join":           strings.Join,
		"subtract":       subtract,
		"indent":         indent,
		"nindent":        nindent,
		"trim":           strings.TrimSpace,
		"wrap":           func(input string, offset int) string { return wrap(input, offset, maxLineLength) },
		"offset":         offset,
		"offsetCommands": offsetCommands,
	}

	if wa, ok := customFuncs["wrapAt"]; ok {
		if wrapAtFunc, ok := wa.(func() int); ok {
			wrapAt := wrapAtFunc()
			customFuncs["wrap"] = func(input string, offset int) string {
				return wrap(input, offset, wrapAt)
			}
		}
	}

	for key, value := range customFuncs {
		funcMap[key] = value
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))

	if _, err := t.New("helpNameTemplate").Parse(helpNameTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("argsTemplate").Parse(argsTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("usageTemplate").Parse(usageTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("descriptionTemplate").Parse(descriptionTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleCommandTemplate").Parse(visibleCommandTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("copyrightTemplate").Parse(copyrightTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("versionTemplate").Parse(versionTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleFlagCategoryTemplate").Parse(visibleFlagCategoryTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleFlagTemplate").Parse(visibleFlagTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visiblePersistentFlagTemplate").Parse(visiblePersistentFlagTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleGlobalFlagCategoryTemplate").Parse(strings.ReplaceAll(visibleFlagCategoryTemplate, "OPTIONS", "GLOBAL OPTIONS")); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("authorsTemplate").Parse(authorsTemplate); err != nil {
		handleTemplateError(err)
	}

	if _, err := t.New("visibleCommandCategoryTemplate").Parse(visibleCommandCategoryTemplate); err != nil {
		handleTemplateError(err)
	}

	tracef("executing template")
	handleTemplateError(t.Execute(w, data))

	_ = w.Flush()
}

// DefaultPrintHelp is the default implementation of HelpPrinter.
func DefaultPrintHelp(out io.Writer, templ string, data any) {
	HelpPrinterCustom(out, templ, data, nil)
}

func checkVersion(cmd *Command) bool {
	return cmd.versionFlag != nil && cmd.versionFlag.IsSet()
}

// completionRequest is what a shell completion request says about the word being
// completed.
type completionRequest struct {
	// word is the word the shell is completing. wordKnown says whether the request
	// carried it: the deprecated request form does not.
	word      string
	wordKnown bool
	// terminated says whether a "--" precedes the word, which makes that word a
	// positional argument of whatever the command runs rather than one this command
	// has any suggestion for.
	terminated bool
}

// parseShellCompleteRequest reports whether arguments are a shell completion request
// and returns the arguments to parse. What the request says about the word being
// completed is recorded on c, which is the root command.
//
// Two request forms are understood. The current one names the request up front:
//
//	<cmd> __complete <word>... <word being completed>
//
// The completion scripts send every word before the cursor, then the word under the
// cursor, which is the empty string when the cursor sits on a fresh word. Naming the
// request in the first argument is what keeps it out of reach of "--": everything
// after that terminator is a positional argument, so a request appended at the end of
// the command line cannot be told apart from a positional argument that happens to
// look like one. See the deprecated form below for what that ambiguity costs.
//
// The deprecated form appends completionFlag to the command line. Scripts generated
// before this change still use it, so it keeps working, with one caveat it cannot
// escape: a command line holding "--" is answered as an ordinary run rather than as a
// completion, because after "--" the flag is a positional argument that belongs to
// whatever the command runs. That is what a wrapper command needs (see
// https://github.com/urfave/cli/issues/1932), and it is why a shell that appends the
// flag after a "--" runs the command instead of completing it (see
// https://github.com/urfave/cli/issues/1993). Regenerating the completion script and
// sourcing it again resolves that in favor of completing, since the request is then
// no longer something a command line can imitate.
func parseShellCompleteRequest(c *Command, arguments []string) (bool, []string) {
	// Whatever the previous run of this Command recorded says nothing about this one.
	c.completion = nil

	if (c.parent == nil && !c.EnableShellCompletion) || (c.parent != nil && !c.Root().shellCompletion) {
		return false, arguments
	}

	// A command of that name, if the app happens to have one, is what was asked for:
	// the request form is understood only where it shadows nothing.
	if len(arguments) > 1 && arguments[1] == completionCommandRequest && c.Command(completionCommandRequest) == nil {
		return true, c.parseCompletionRequest(arguments)
	}

	pos := len(arguments) - 1
	lastArg := arguments[pos]

	if lastArg != completionFlag {
		return false, arguments
	}

	// The word being completed is at position pos-1, immediately before
	// completionFlag, so only the arguments before that position are checked and
	// completing "--" itself still works.
	// https://unix.stackexchange.com/a/11382
	if pos >= 1 && slices.Contains(arguments[:pos-1], "--") {
		// The flag is a positional argument here, so it is left in place for the
		// command to pass on, and the command runs.
		return false, arguments
	}

	// This request form does not say which word is being completed, so
	// DefaultCompleteWithFlags works it out from the arguments.
	c.completion = &completionRequest{}
	return true, arguments[:pos]
}

// parseCompletionRequest records what a request naming completionCommandRequest says
// about the word being completed, and returns the arguments to parse.
//
// The word is kept in those arguments when it starts with "-", and dropped from them
// otherwise, which is the shape the deprecated request form produced. A ShellComplete
// function reading cmd.Args() therefore sees the same thing under both forms.
func (cmd *Command) parseCompletionRequest(arguments []string) []string {
	// arguments[0] is the program, arguments[1] is completionCommandRequest, and the
	// word being completed is last. A request holding neither, which no script sends,
	// is read as an empty word on an empty command line.
	var words []string
	word := ""
	if len(arguments) > 2 {
		words = arguments[2 : len(arguments)-1]
		word = arguments[len(arguments)-1]
	}
	cmd.completion = &completionRequest{
		word:      word,
		wordKnown: true,
		// Everything after a "--" is a positional argument of whatever the command
		// runs. A "--" being completed is not one: it is the word itself, and flags
		// still answer it.
		terminated: slices.Contains(words, "--"),
	}

	args := make([]string, 0, len(arguments)-1)
	args = append(args, arguments[0])
	args = append(args, words...)
	if strings.HasPrefix(word, "-") {
		args = append(args, word)
	}
	return args
}

func shouldRunCompletion(cmd *Command) bool {
	tracef("checking completions on command %[1]q", cmd.Name)

	if !cmd.Root().shellCompletion {
		tracef("completion not enabled skipping %[1]q", cmd.Name)
		return false
	}

	if argsArguments := cmd.Args(); argsArguments.Present() {
		name := argsArguments.First()
		if cmd := cmd.Command(name); cmd != nil {
			// let the command handle the completion
			return false
		}
	}

	tracef("no subcommand found for completion %[1]q", cmd.Name)
	return true
}

func runCompletion(ctx context.Context, cmd *Command) {
	if cmd.ShellComplete != nil {
		tracef("running shell completion func for command %[1]q", cmd.Name)
		cmd.ShellComplete(ctx, cmd)
	}
}

func subtract(a, b int) int {
	return a - b
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.ReplaceAll(v, "\n", "\n"+pad)
}

func nindent(spaces int, v string) string {
	return "\n" + indent(spaces, v)
}

func wrap(input string, offset int, wrapAt int) string {
	var ss []string

	lines := strings.Split(input, "\n")

	padding := strings.Repeat(" ", offset)

	for i, line := range lines {
		if line == "" {
			ss = append(ss, line)
		} else {
			wrapped := wrapLine(line, offset, wrapAt, padding)
			if i == 0 {
				ss = append(ss, wrapped)
			} else {
				ss = append(ss, padding+wrapped)
			}

		}
	}

	return strings.Join(ss, "\n")
}

func wrapLine(input string, offset int, wrapAt int, padding string) string {
	if wrapAt <= offset || len(input) <= wrapAt-offset {
		return input
	}

	lineWidth := wrapAt - offset
	words := strings.Fields(input)
	if len(words) == 0 {
		return input
	}

	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + padding + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return wrapped
}

func offset(input string, fixed int) int {
	return len(input) + fixed
}

// this function tries to find the max width of the names column
// so say we have the following rows for help
//
//	foo1, foo2, foo3  some string here
//	bar1, b2 some other string here
//
// We want to offset the 2nd row usage by some amount so that everything
// is aligned
//
//	foo1, foo2, foo3  some string here
//	bar1, b2          some other string here
//
// to find that offset we find the length of all the rows and use the max
// to calculate the offset
func offsetCommands(cmds []*Command, fixed int) int {
	max := 0
	for _, cmd := range cmds {
		s := strings.Join(cmd.Names(), ", ")
		if len(s) > max {
			max = len(s)
		}
	}
	return max + fixed
}
