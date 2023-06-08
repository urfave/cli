package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// ignoreFlagPrefix is to ignore test flags when adding flags from other packages
const ignoreFlagPrefix = "test."

const maxCommandGraphDepth = 256

// Command is a subcommand for a cli.App.
type Command struct {
	// The name of the command
	Name string
	// Full name of command for help, defaults to full command name, including parent commands.
	HelpName        string
	commandNamePath []string
	// A list of aliases for the command
	Aliases []string
	// A short description of the usage of this command
	Usage string
	// Text to override the USAGE section of help
	UsageText string
	// A short description of the arguments of this command
	ArgsUsage string
	// Version of the command
	Version string
	// Longer explanation of how the command works
	Description string
	// DefaultCommand is the (optional) name of a command
	// to run if no command names are passed as CLI arguments.
	DefaultCommand string
	// The category the command is part of
	Category string
	// List of child commands
	Commands []*Command
	// categories contains the categorized commands and is populated on app startup
	categories CommandCategories
	// List of flags to parse
	Flags []Flag
	// flagCategories contains the categorized flags and is populated on app startup
	flagCategories FlagCategories
	// Boolean to hide built-in help command and help flag
	HideHelp bool
	// Ignored if HideHelp is true.
	HideHelpCommand bool
	// Boolean to hide built-in version flag and the VERSION section of help
	HideVersion bool
	// Boolean to enable shell completion commands
	EnableShellCompletion bool
	// Shell Completion generation command name
	ShellCompletionCommandName string
	// The function to call when checking for shell command completions
	ShellComplete ShellCompleteFunc
	// An action to execute before any subcommands are run, but after the context is ready
	// If a non-nil error is returned, no subcommands are run
	Before BeforeFunc
	// An action to execute after any subcommands are run, but after the subcommand has finished
	// It is run even if Action() panics
	After AfterFunc
	// The function to call when this command is invoked
	Action ActionFunc
	// Execute this function if the proper command cannot be found
	CommandNotFound CommandNotFoundFunc
	// Execute this function if a usage error occurs.
	OnUsageError OnUsageErrorFunc
	// Execute this function when an invalid flag is accessed from the context
	InvalidFlagAccessHandler InvalidFlagAccessFunc
	// Boolean to hide this command from help or completion
	Hidden bool
	// List of all authors who contributed (string or fmt.Stringer)
	Authors []any // TODO: ~string | fmt.Stringer when interface unions are available
	// Copyright of the binary if any
	Copyright string
	// Reader reader to write input to (useful for tests)
	Reader io.Reader
	// Writer writer to write output to
	Writer io.Writer
	// ErrWriter writes error output
	ErrWriter io.Writer
	// ExitErrHandler processes any error encountered while running an App before
	// it is returned to the caller. If no function is provided, HandleExitCoder
	// is used as the default behavior.
	ExitErrHandler ExitErrHandlerFunc
	// Other custom info
	Metadata map[string]interface{}
	// Carries a function which returns app specific info.
	ExtraInfo func() map[string]string
	// CustomAppHelpTemplate the text template for app help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomAppHelpTemplate string
	// SliceFlagSeparator is used to customize the separator for SliceFlag, the default is ","
	SliceFlagSeparator string
	// DisableSliceFlagSeparator is used to disable SliceFlagSeparator, the default is false
	DisableSliceFlagSeparator bool
	// Boolean to enable short-option handling so user can combine several
	// single-character bool arguments into one
	// i.e. foobar -o -v -> foobar -ov
	UseShortOptionHandling bool
	// Enable suggestions for commands and flags
	Suggest bool
	// Allows global flags set by libraries which use flag.XXXVar(...) directly
	// to be parsed through this library
	AllowExtFlags bool
	// Treat all flags as normal arguments if true
	SkipFlagParsing bool
	// Flag exclusion group
	MutuallyExclusiveFlags []MutuallyExclusiveFlags

	// CustomHelpTemplate the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomHelpTemplate string

	// Use longest prefix match for commands
	PrefixMatchCommands bool
	// Custom suggest command for matching
	SuggestCommandFunc SuggestCommandFunc

	// The parent of this command. This value will be nil for the
	// command at the root of the graph.
	parent *Command

	// track state of error handling
	isInError bool
}

type Commands []*Command

type CommandsByName []*Command

func (c CommandsByName) Len() int {
	return len(c)
}

func (c CommandsByName) Less(i, j int) bool {
	return lexicographicLess(c[i].Name, c[j].Name)
}

func (c CommandsByName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// FullName returns the full name of the command.
// For commands with parets this ensures that the parent commands
// are part of the command path.
func (cmd *Command) FullName() string {
	if cmd.commandNamePath == nil {
		return cmd.Name
	}

	return strings.Join(cmd.commandNamePath, " ")
}

func (cmd *Command) Command(name string) *Command {
	for _, subCmd := range cmd.Commands {
		if subCmd.HasName(name) {
			return subCmd
		}
	}

	return nil
}

func (cmd *Command) setupDefaults(arguments []string) {
	isRoot := cmd.parent == nil

	tracef("isRoot? %[1]v", isRoot)

	if cmd.ShellComplete == nil {
		tracef("setting default ShellComplete")
		cmd.ShellComplete = DefaultCompleteWithFlags(cmd)
	}

	if cmd.Name == "" && isRoot {
		tracef("setting cmd.Name from first arg basename")
		cmd.Name = filepath.Base(arguments[0])
	}

	if cmd.HelpName == "" {
		tracef("setting cmd.HelpName to cmd.Name %[1]q", cmd.Name)
		cmd.HelpName = cmd.Name
	}

	if cmd.Usage == "" && isRoot {
		tracef("setting default Usage")
		cmd.Usage = "A new cli application"
	}

	if cmd.Version == "" {
		tracef("setting HideVersion=true due to empty Version")
		cmd.HideVersion = true
	}

	if cmd.Action == nil {
		tracef("setting default Action as help command action")
		cmd.Action = helpCommandAction
	}

	if cmd.Reader == nil {
		tracef("setting default Reader as os.Stdin")
		cmd.Reader = os.Stdin
	}

	if cmd.Writer == nil {
		tracef("setting default Writer as os.Stdout")
		cmd.Writer = os.Stdout
	}

	if cmd.ErrWriter == nil {
		tracef("setting default ErrWriter as os.Stderr")
		cmd.ErrWriter = os.Stderr
	}

	if cmd.AllowExtFlags {
		tracef("visiting all flags given AllowExtFlags=true")
		// add global flags added by other packages
		flag.VisitAll(func(f *flag.Flag) {
			// skip test flags
			if !strings.HasPrefix(f.Name, ignoreFlagPrefix) {
				cmd.Flags = append(cmd.Flags, &extFlag{f})
			}
		})
	}

	for _, subCmd := range cmd.Commands {
		tracef("setting sub-command parent as self")
		subCmd.parent = cmd

		/*
			subCmdName := subCmd.Name

			if subCmd.HelpName != "" {
				tracef("setting subCmdName to %[1]q (cmd=%[2]q)", subCmd.HelpName, cmd.Name)
				subCmdName = subCmd.HelpName
			}
		*/

		newSubCmdHelpName := fmt.Sprintf("%[1]s %[2]s", cmd.HelpName, subCmd.Name)

		tracef("setting subCmd.HelpName to %[1]q (cmd=%[2]q)", newSubCmdHelpName, cmd.Name)
		subCmd.HelpName = newSubCmdHelpName

		tracef("assigning subCmd.flagCategories from subCmd.Flags")
		subCmd.flagCategories = newFlagCategoriesFromFlags(subCmd.Flags)
	}

	tracef("ensuring help command and flag")
	cmd.ensureHelp()

	if !cmd.HideVersion && isRoot {
		tracef("appending version flag")
		cmd.appendFlag(VersionFlag)
	}

	if cmd.PrefixMatchCommands && cmd.SuggestCommandFunc == nil {
		tracef("setting default SuggestCommandFunc")
		cmd.SuggestCommandFunc = suggestCommand
	}

	if cmd.EnableShellCompletion {
		completionCommand := buildCompletionCommand()

		if cmd.ShellCompletionCommandName != "" {
			tracef("setting completion command name from ShellCompletionCommandName")
			completionCommand.Name = cmd.ShellCompletionCommandName
		}

		tracef("appending completionCommand")
		cmd.appendCommand(completionCommand)
	}

	tracef("setting command categories")
	cmd.categories = newCommandCategories()

	for _, subCmd := range cmd.Commands {
		cmd.categories.AddCommand(subCmd.Category, subCmd)
	}

	tracef("sorting command categories")
	sort.Sort(cmd.categories.(*commandCategories))

	tracef("setting flag categories")
	cmd.flagCategories = newFlagCategoriesFromFlags(cmd.Flags)

	if cmd.Metadata == nil {
		tracef("setting default Metadata")
		cmd.Metadata = map[string]any{}
	}

	if len(cmd.SliceFlagSeparator) != 0 {
		tracef("setting defaultSliceFlagSeparator from cmd.SliceFlagSeparator")
		defaultSliceFlagSeparator = cmd.SliceFlagSeparator
	}

	tracef("setting disableSliceFlagSeparator from cmd.DisableSliceFlagSeparator")
	disableSliceFlagSeparator = cmd.DisableSliceFlagSeparator
}

func (cmd *Command) setupCommandGraph(cCtx *Context) {
	for _, subCmd := range cmd.Commands {
		subCmd.parent = cmd
		subCmd.setupSubcommand(cCtx)
		subCmd.setupCommandGraph(cCtx)
	}
}

func (cmd *Command) setupSubcommand(cCtx *Context) {
	cmd.ensureHelp()

	if cCtx.Command.UseShortOptionHandling {
		cmd.UseShortOptionHandling = true
	}

	for _, subCmd := range cmd.Commands {
		if subCmd.HelpName == "" {
			newSubCmdHelpName := fmt.Sprintf("%[1]s %[2]s", cmd.HelpName, subCmd.Name)

			tracef("setting subCmd.HelpName to %[1]q (cmd=%[2]q)", newSubCmdHelpName, cmd.Name)
			subCmd.HelpName = newSubCmdHelpName
		}
	}

	tracef("setting command categories")
	cmd.categories = newCommandCategories()

	for _, subCmd := range cmd.Commands {
		cmd.categories.AddCommand(subCmd.Category, subCmd)
	}

	tracef("sorting command categories")
	sort.Sort(cmd.categories.(*commandCategories))

	tracef("setting flag categories")
	cmd.flagCategories = newFlagCategoriesFromFlags(cmd.Flags)
}

func (cmd *Command) ensureHelp() {
	helpCommand := buildHelpCommand(true)

	if cmd.Command(helpCommand.Name) == nil && !cmd.HideHelp {
		if !cmd.HideHelpCommand {
			newHelpCmdName := fmt.Sprintf("%[1]s %[2]s", cmd.HelpName, helpName)

			tracef("setting helpCommand.HelpName to %[1]q (cmd=%[2]q)", newHelpCmdName, cmd.Name)
			helpCommand.HelpName = newHelpCmdName

			tracef("appending helpCommand")
			cmd.appendCommand(helpCommand)
		}
	}

	if HelpFlag != nil && !cmd.HideHelp {
		tracef("appending HelpFlag")
		cmd.appendFlag(HelpFlag)
	}
}

// Run is the entry point to the command graph. The positional
// arguments are parsed according to the Flag and Command
// definitions and the matching Action functions are run.
func (cmd *Command) Run(ctx context.Context, arguments []string) (deferErr error) {
	cmd.setupDefaults(arguments)

	parentContext := &Context{Context: ctx}
	if v, ok := ctx.Value(contextContextKey).(*Context); ok {
		parentContext = v
	}

	// handle the completion flag separately from the flagset since
	// completion could be attempted after a flag, but before its value was put
	// on the command line. this causes the flagset to interpret the completion
	// flag name as the value of the flag before it which is undesirable
	// note that we can only do this because the shell autocomplete function
	// always appends the completion flag at the end of the command
	shellComplete, arguments := checkShellCompleteFlag(cmd, arguments)

	cCtx := NewContext(cmd, nil, parentContext)
	cCtx.shellComplete = shellComplete

	cCtx.Command = cmd

	ctx = context.WithValue(ctx, contextContextKey, cCtx)

	if cmd.parent == nil {
		cmd.setupCommandGraph(cCtx)
	}

	a := args(arguments)
	set, err := cmd.parseFlags(&a, cCtx)
	cCtx.flagSet = set

	if checkCompletions(cCtx) {
		return nil
	}

	if err != nil {
		tracef("setting deferErr from %[1]v", err)
		deferErr = err

		cCtx.Command.isInError = true
		if cmd.OnUsageError != nil {
			err = cmd.OnUsageError(cCtx, err, cmd.parent != nil)
			err = cCtx.Command.handleExitCoder(cCtx, err)
			return err
		}
		_, _ = fmt.Fprintf(cCtx.Command.writer(), "%s %s\n\n", "Incorrect Usage:", err.Error())
		if cCtx.Command.Suggest {
			if suggestion, err := cmd.suggestFlagFromError(err, ""); err == nil {
				fmt.Fprintf(cCtx.Command.writer(), "%s", suggestion)
			}
		}
		if !cmd.HideHelp {
			if cmd.parent == nil {
				_ = ShowAppHelp(cCtx)
			} else {
				_ = ShowCommandHelp(cCtx.parent, cmd.Name)
			}
		}
		return err
	}

	if checkHelp(cCtx) {
		return helpCommandAction(cCtx)
	}

	if cmd.parent == nil && !cCtx.Command.HideVersion && checkVersion(cCtx) {
		ShowVersion(cCtx)
		return nil
	}

	if cmd.After != nil && !cCtx.shellComplete {
		defer func() {
			if err := cmd.After(cCtx); err != nil {
				err = cCtx.Command.handleExitCoder(cCtx, err)

				if deferErr != nil {
					deferErr = newMultiError(deferErr, err)
				} else {
					deferErr = err
				}
			}
		}()
	}

	if err := cCtx.checkRequiredFlags(cmd.Flags); err != nil {
		cCtx.Command.isInError = true
		_ = ShowSubcommandHelp(cCtx)
		return err
	}

	for _, grp := range cmd.MutuallyExclusiveFlags {
		if err := grp.check(cCtx); err != nil {
			_ = ShowSubcommandHelp(cCtx)
			return err
		}
	}

	if cmd.Before != nil && !cCtx.shellComplete {
		if err := cmd.Before(cCtx); err != nil {
			deferErr = cCtx.Command.handleExitCoder(cCtx, err)
			return deferErr
		}
	}

	if err := runFlagActions(cCtx, cmd.Flags); err != nil {
		return err
	}

	var subCmd *Command
	args := cCtx.Args()
	if args.Present() {
		name := args.First()
		if cCtx.Command.SuggestCommandFunc != nil {
			name = cCtx.Command.SuggestCommandFunc(cmd.Commands, name)
		}
		subCmd = cmd.Command(name)
		if subCmd == nil {
			hasDefault := cCtx.Command.DefaultCommand != ""
			isFlagName := checkStringSliceIncludes(name, cCtx.FlagNames())

			var (
				isDefaultSubcommand   = false
				defaultHasSubcommands = false
			)

			if hasDefault {
				dc := cCtx.Command.Command(cCtx.Command.DefaultCommand)
				defaultHasSubcommands = len(dc.Commands) > 0
				for _, dcSub := range dc.Commands {
					if checkStringSliceIncludes(name, dcSub.Names()) {
						isDefaultSubcommand = true
						break
					}
				}
			}

			if isFlagName || (hasDefault && (defaultHasSubcommands && isDefaultSubcommand)) {
				argsWithDefault := cCtx.Command.argsWithDefaultCommand(args)
				if !reflect.DeepEqual(args, argsWithDefault) {
					subCmd = cCtx.Command.Command(argsWithDefault.First())
				}
			}
		}
	} else if cmd.parent == nil && cCtx.Command.DefaultCommand != "" {
		if dc := cCtx.Command.Command(cCtx.Command.DefaultCommand); dc != cmd {
			subCmd = dc
		}
	}

	if subCmd != nil {
		/*
			newcCtx := NewContext(cCtx.Command, nil, cCtx)
			newcCtx.Command = cmd
		*/
		return subCmd.Run(ctx, cCtx.Args().Slice())
	}

	if cmd.Action == nil {
		cmd.Action = helpCommandAction
	}

	if err := cmd.Action(cCtx); err != nil {
		tracef("calling handleExitCoder with %[1]v", err)
		deferErr = cCtx.Command.handleExitCoder(cCtx, err)
	}

	tracef("returning deferErr")
	return deferErr
}

func (c *Command) newFlagSet() (*flag.FlagSet, error) {
	return flagSet(c.Name, c.allFlags())
}

func (c *Command) allFlags() []Flag {
	var flags []Flag
	flags = append(flags, c.Flags...)
	for _, grpf := range c.MutuallyExclusiveFlags {
		for _, f1 := range grpf.Flags {
			flags = append(flags, f1...)
		}
	}
	return flags
}

func (c *Command) useShortOptionHandling() bool {
	return c.UseShortOptionHandling
}

func (cmd *Command) suggestFlagFromError(err error, commandName string) (string, error) {
	fl, parseErr := flagFromError(err)
	if parseErr != nil {
		return "", err
	}

	flags := cmd.Flags
	hideHelp := cmd.HideHelp

	if commandName != "" {
		subCmd := cmd.Command(commandName)
		if subCmd == nil {
			return "", err
		}
		flags = subCmd.Flags
		hideHelp = hideHelp || subCmd.HideHelp
	}

	suggestion := SuggestFlag(flags, fl, hideHelp)
	if len(suggestion) == 0 {
		return "", err
	}

	return fmt.Sprintf(SuggestDidYouMeanTemplate, suggestion) + "\n\n", nil
}

func (c *Command) parseFlags(args Args, ctx *Context) (*flag.FlagSet, error) {
	set, err := c.newFlagSet()
	if err != nil {
		return nil, err
	}

	if c.SkipFlagParsing {
		return set, set.Parse(append([]string{"--"}, args.Tail()...))
	}

	for pCtx := ctx.parent; pCtx != nil; pCtx = pCtx.parent {
		if pCtx.Command == nil {
			continue
		}

		for _, fl := range pCtx.Command.Flags {
			pfl, ok := fl.(PersistentFlag)
			if !ok || !pfl.IsPersistent() {
				continue
			}

			applyPersistentFlag := true
			set.VisitAll(func(f *flag.Flag) {
				for _, name := range fl.Names() {
					if name == f.Name {
						applyPersistentFlag = false
						break
					}
				}
			})

			if !applyPersistentFlag {
				continue
			}

			if err := fl.Apply(set); err != nil {
				return nil, err
			}
		}
	}

	if err := parseIter(set, c, args.Tail(), ctx.shellComplete); err != nil {
		return nil, err
	}

	if err := normalizeFlags(c.Flags, set); err != nil {
		return nil, err
	}

	return set, nil
}

// Names returns the names including short names and aliases.
func (cmd *Command) Names() []string {
	return append([]string{cmd.Name}, cmd.Aliases...)
}

// HasName returns true if Command.Name matches given name
func (cmd *Command) HasName(name string) bool {
	for _, n := range cmd.Names() {
		if n == name {
			return true
		}
	}

	return false
}

// VisibleCategories returns a slice of categories and commands that are
// Hidden=false
func (c *Command) VisibleCategories() []CommandCategory {
	ret := []CommandCategory{}
	for _, category := range c.categories.Categories() {
		if visible := func() CommandCategory {
			if len(category.VisibleCommands()) > 0 {
				return category
			}
			return nil
		}(); visible != nil {
			ret = append(ret, visible)
		}
	}
	return ret
}

// VisibleCommands returns a slice of the Commands with Hidden=false
func (c *Command) VisibleCommands() []*Command {
	var ret []*Command
	for _, command := range c.Commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}

// VisibleFlagCategories returns a slice containing all the visible flag categories with the flags they contain
func (c *Command) VisibleFlagCategories() []VisibleFlagCategory {
	if c.flagCategories == nil {
		c.flagCategories = newFlagCategoriesFromFlags(c.Flags)
	}
	return c.flagCategories.VisibleCategories()
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (c *Command) VisibleFlags() []Flag {
	return visibleFlags(c.Flags)
}

func (c *Command) appendFlag(fl Flag) {
	if !hasFlag(c.Flags, fl) {
		c.Flags = append(c.Flags, fl)
	}
}

func (cmd *Command) appendCommand(aCmd *Command) {
	if !hasCommand(cmd.Commands, aCmd) {
		cmd.Commands = append(cmd.Commands, aCmd)
	}
}

func (cmd *Command) handleExitCoder(cCtx *Context, err error) error {
	if cmd.ExitErrHandler != nil {
		cmd.ExitErrHandler(cCtx, err)
		return err
	}

	HandleExitCoder(err)
	return err
}

func (c *Command) argsWithDefaultCommand(oldArgs Args) Args {
	if c.DefaultCommand != "" {
		rawArgs := append([]string{c.DefaultCommand}, oldArgs.Slice()...)
		newArgs := args(rawArgs)

		return &newArgs
	}

	return oldArgs
}

func (cmd *Command) errWriter() io.Writer {
	if cmd.parent != nil {
		return cmd.parent.errWriter()
	}

	if cmd.ErrWriter == nil {
		return os.Stderr
	}

	return cmd.ErrWriter
}

func (cmd *Command) writer() io.Writer {
	if cmd.parent != nil {
		return cmd.parent.writer()
	}

	if cmd.isInError {
		if cmd.ErrWriter == nil {
			return os.Stderr
		}

		return cmd.ErrWriter
	}

	if cmd.Writer == nil {
		return os.Stdout
	}

	return cmd.Writer
}

func hasCommand(commands []*Command, command *Command) bool {
	for _, existing := range commands {
		if command == existing {
			return true
		}
	}

	return false
}

func runFlagActions(cCtx *Context, flags []Flag) error {
	for _, fl := range flags {
		isSet := false

		for _, name := range fl.Names() {
			if cCtx.IsSet(name) {
				isSet = true
				break
			}
		}

		if !isSet {
			continue
		}

		if af, ok := fl.(ActionableFlag); ok {
			if err := af.RunAction(cCtx); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkStringSliceIncludes(want string, sSlice []string) bool {
	found := false
	for _, s := range sSlice {
		if want == s {
			found = true
			break
		}
	}

	return found
}
