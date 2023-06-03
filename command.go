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

func (cmd *Command) setupDefaults() {
	isRoot := cmd.parent == nil

	if cmd.ShellComplete == nil {
		cmd.ShellComplete = DefaultCompleteWithFlags(cmd)
	}

	if cmd.Name == "" && isRoot {
		cmd.Name = filepath.Base(os.Args[0])
	}

	if cmd.HelpName == "" {
		cmd.HelpName = cmd.Name
	}

	if cmd.Usage == "" && isRoot {
		cmd.Usage = "A new cli application"
	}

	if cmd.Version == "" {
		cmd.HideVersion = true
	}

	if cmd.Action == nil {
		cmd.Action = helpCommand.Action
	}

	if cmd.Reader == nil {
		cmd.Reader = os.Stdin
	}

	if cmd.Writer == nil {
		cmd.Writer = os.Stdout
	}

	if cmd.ErrWriter == nil {
		cmd.ErrWriter = os.Stderr
	}

	if cmd.AllowExtFlags {
		// add global flags added by other packages
		flag.VisitAll(func(f *flag.Flag) {
			// skip test flags
			if !strings.HasPrefix(f.Name, ignoreFlagPrefix) {
				cmd.Flags = append(cmd.Flags, &extFlag{f})
			}
		})
	}

	var newCommands []*Command

	for _, subCmd := range cmd.Commands {
		cname := subCmd.Name
		if subCmd.HelpName != "" {
			cname = subCmd.HelpName
		}
		subCmd.HelpName = fmt.Sprintf("%s %s", cmd.HelpName, cname)

		subCmd.flagCategories = newFlagCategoriesFromFlags(subCmd.Flags)
		newCommands = append(newCommands, subCmd)
	}

	cmd.Commands = newCommands

	if cmd.Command(helpCommand.Name) == nil && !cmd.HideHelp {
		if !cmd.HideHelpCommand {
			helpCommand.HelpName = fmt.Sprintf("%s %s", cmd.HelpName, helpName)
			cmd.appendCommand(helpCommand)
		}

		if HelpFlag != nil {
			cmd.appendFlag(HelpFlag)
		}
	}

	if !cmd.HideVersion && isRoot {
		cmd.appendFlag(VersionFlag)
	}

	if cmd.PrefixMatchCommands {
		if cmd.SuggestCommandFunc == nil {
			cmd.SuggestCommandFunc = suggestCommand
		}
	}

	if cmd.EnableShellCompletion {
		if cmd.ShellCompletionCommandName != "" {
			completionCommand.Name = cmd.ShellCompletionCommandName
		}

		cmd.appendCommand(completionCommand)
	}

	cmd.categories = newCommandCategories()

	for _, command := range cmd.Commands {
		cmd.categories.AddCommand(command.Category, command)
	}

	sort.Sort(cmd.categories.(*commandCategories))

	cmd.flagCategories = newFlagCategories()

	for _, fl := range cmd.Flags {
		if cf, ok := fl.(CategorizableFlag); ok {
			if cf.GetCategory() != "" {
				cmd.flagCategories.AddFlag(cf.GetCategory(), cf)
			}
		}
	}

	if cmd.Metadata == nil {
		cmd.Metadata = map[string]any{}
	}

	if len(cmd.SliceFlagSeparator) != 0 {
		defaultSliceFlagSeparator = cmd.SliceFlagSeparator
	}

	disableSliceFlagSeparator = cmd.DisableSliceFlagSeparator
}

func (c *Command) setup(ctx *Context) {
	if c.Command(helpCommand.Name) == nil && !c.HideHelp {
		if !c.HideHelpCommand {
			helpCommand.HelpName = fmt.Sprintf("%s %s", c.HelpName, helpName)
			c.Commands = append(c.Commands, helpCommand)
		}
	}

	if !c.HideHelp && HelpFlag != nil {
		// append help to flags
		c.appendFlag(HelpFlag)
	}

	if ctx.Command.UseShortOptionHandling {
		c.UseShortOptionHandling = true
	}

	c.categories = newCommandCategories()
	for _, command := range c.Commands {
		c.categories.AddCommand(command.Category, command)
	}

	sort.Sort(c.categories.(*commandCategories))

	var newCmds []*Command

	for _, scmd := range c.Commands {
		if scmd.HelpName == "" {
			scmd.HelpName = fmt.Sprintf("%s %s", c.HelpName, scmd.Name)
		}
		newCmds = append(newCmds, scmd)
	}

	c.Commands = newCmds
}

// Run is the entry point to the command graph. The positional
// arguments are parsed according to the Flag and Command
// definitions and the matching Action functions are run.
func (cmd *Command) Run(ctx context.Context, arguments []string) error {
	cmd.setupDefaults()

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

	if cmd.parent != nil {
		cmd.setup(cCtx)
	}

	a := args(arguments)
	set, err := cmd.parseFlags(&a, cCtx)
	cCtx.flagSet = set

	if checkCompletions(cCtx) {
		return nil
	}

	if err != nil {
		cCtx.Command.isInError = true
		if cmd.OnUsageError != nil {
			err = cmd.OnUsageError(cCtx, err, cmd.parent != nil)
			cCtx.Command.handleExitCoder(cCtx, err)
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
				_ = ShowCommandHelp(cCtx.parentContext, cmd.Name)
			}
		}
		return err
	}

	if checkHelp(cCtx) {
		return helpCommand.Action(cCtx)
	}

	if cmd.parent == nil && !cCtx.Command.HideVersion && checkVersion(cCtx) {
		ShowVersion(cCtx)
		return nil
	}

	if cmd.After != nil && !cCtx.shellComplete {
		defer func() {
			afterErr := cmd.After(cCtx)
			if afterErr != nil {
				cCtx.Command.handleExitCoder(cCtx, err)
				if err != nil {
					err = newMultiError(err, afterErr)
				} else {
					err = afterErr
				}
			}
		}()
	}

	cerr := cCtx.checkRequiredFlags(cmd.Flags)
	if cerr != nil {
		cCtx.Command.isInError = true
		_ = ShowSubcommandHelp(cCtx)
		return cerr
	}

	for _, grp := range cmd.MutuallyExclusiveFlags {
		if err := grp.check(cCtx); err != nil {
			_ = ShowSubcommandHelp(cCtx)
			return err
		}
	}

	if cmd.Before != nil && !cCtx.shellComplete {
		beforeErr := cmd.Before(cCtx)
		if beforeErr != nil {
			cCtx.Command.handleExitCoder(cCtx, beforeErr)
			err = beforeErr
			return err
		}
	}

	if err = runFlagActions(cCtx, cmd.Flags); err != nil {
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
		cmd.Action = helpCommand.Action
	}

	err = cmd.Action(cCtx)

	cCtx.Command.handleExitCoder(cCtx, err)
	return err
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

func (c *Command) suggestFlagFromError(err error, command string) (string, error) {
	flag, parseErr := flagFromError(err)
	if parseErr != nil {
		return "", err
	}

	flags := c.Flags
	hideHelp := c.HideHelp
	if command != "" {
		cmd := c.Command(command)
		if cmd == nil {
			return "", err
		}
		flags = cmd.Flags
		hideHelp = hideHelp || cmd.HideHelp
	}

	suggestion := SuggestFlag(flags, flag, hideHelp)
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

	for pCtx := ctx.parentContext; pCtx != nil; pCtx = pCtx.parentContext {
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
func (c *Command) Names() []string {
	return append([]string{c.Name}, c.Aliases...)
}

// HasName returns true if Command.Name matches given name
func (c *Command) HasName(name string) bool {
	for _, n := range c.Names() {
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

func (c *Command) handleExitCoder(cCtx *Context, err error) {
	if c.ExitErrHandler != nil {
		c.ExitErrHandler(cCtx, err)
	} else {
		HandleExitCoder(err)
	}
}

func (c *Command) argsWithDefaultCommand(oldArgs Args) Args {
	if c.DefaultCommand != "" {
		rawArgs := append([]string{c.DefaultCommand}, oldArgs.Slice()...)
		newArgs := args(rawArgs)

		return &newArgs
	}

	return oldArgs
}

func (c *Command) writer() io.Writer {
	if c.isInError {
		// this can happen in test but not in normal usage
		if c.ErrWriter == nil {
			return os.Stderr
		}
		return c.ErrWriter
	}
	return c.Writer
}

func hasCommand(commands []*Command, command *Command) bool {
	for _, existing := range commands {
		if command == existing {
			return true
		}
	}

	return false
}
