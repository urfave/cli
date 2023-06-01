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

var (
	ignoreFlagPrefix = "test." // this is to ignore test flags when adding flags from other packages
)

// App is an alias for Command that is intended to help clarify
// which Command is considered the "root"
type App = Command

// Command is the primary type used in building a cli app and may
// contain child Commands
type Command struct {
	// The name of the command
	Name string
	// A list of aliases for the command
	Aliases []string
	// A short description of the usage of this command
	Usage string
	// Custom text to show on USAGE section of help
	UsageText string
	// A longer explanation of how the command works
	Description string
	// A short description of the arguments of this command
	ArgsUsage string
	// Version of the program
	Version string
	// DefaultCommand is the (optional) name of a command
	// to run if no command names are passed as CLI arguments.
	DefaultCommand string
	// The category the command is part of
	Category string
	// The function to call when checking for shell command completions
	ShellComplete ShellCompleteFunc
	// An action to execute before any sub-subcommands are run, but after the context is ready
	// If a non-nil error is returned, no sub-subcommands are run
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
	// List of all authors who contributed
	Authors []*Author
	// Copyright of the binary if any
	Copyright string
	// List of child commands
	Commands []*Command
	// List of flags to parse
	Flags          []Flag
	flagCategories FlagCategories
	// Boolean to enable shell completion commands
	EnableShellCompletion bool
	// Treat all flags as normal arguments if true
	SkipFlagParsing bool
	// Boolean to hide built-in help command and help flag
	HideHelp bool
	// Boolean to hide built-in help command but keep help flag
	// Ignored if HideHelp is true.
	HideHelpCommand bool
	// Boolean to hide built-in version flag and the VERSION section of help
	HideVersion bool
	// Boolean to hide this command from help or completion
	Hidden bool
	// Boolean to enable short-option handling so user can combine several
	// single-character bool arguments into one
	// i.e. foobar -o -v -> foobar -ov
	UseShortOptionHandling bool
	// Enable suggestions for commands and flags
	Suggest bool

	// Full name of command for help, defaults to full command name, including parent commands.
	HelpName        string
	commandNamePath []string

	// CustomHelpTemplate the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomHelpTemplate string

	// Use longest prefix match for commands
	PrefixMatchCommands bool

	// categories contains the categorized commands and is populated on app startup
	categories CommandCategories

	// if this is a root "special" command
	isRoot bool

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
	Metadata map[string]any
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
	// Allows global flags set by libraries which use flag.XXXVar(...) directly
	// to be parsed through this library
	AllowExtFlags bool

	// Flag exclusion group
	MutuallyExclusiveFlags []MutuallyExclusiveFlags
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
// For subcommands this ensures that parent commands are part of the command path
func (c *Command) FullName() string {
	if c.commandNamePath == nil {
		return c.Name
	}
	return strings.Join(c.commandNamePath, " ")
}

func (c *Command) setupDefaults() {
	if c.Name == "" {
		c.Name = filepath.Base(os.Args[0])
	}

	if c.HelpName == "" {
		c.HelpName = c.Name
	}

	if c.Usage == "" {
		c.Usage = "A new cli application"
	}

	if c.Version == "" {
		c.HideVersion = true
	}

	if c.ShellComplete == nil {
		c.ShellComplete = DefaultCompleteWithFlags(c)
	}

	if c.Action == nil {
		c.Action = helpCommand.Action
	}

	if c.Reader == nil {
		c.Reader = os.Stdin
	}

	if c.Writer == nil {
		c.Writer = os.Stdout
	}

	if c.ErrWriter == nil {
		c.ErrWriter = os.Stderr
	}

	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
}

func (c *Command) setup(cCtx *Context) {
	if cCtx.parentContext != nil && cCtx.parentContext.Command != nil {
		if cCtx.parentContext.Command.UseShortOptionHandling {
			c.UseShortOptionHandling = true
		}
	}

	if c.Reader == nil {
		c.Reader = cCtx.Reader()
	}
	if c.Writer == nil {
		c.Writer = cCtx.Writer()
	}
	if c.ErrWriter == nil {
		c.ErrWriter = cCtx.ErrWriter()
	}

	c.setupDefaults()

	if c.AllowExtFlags {
		// add global flags added by other packages
		flag.VisitAll(func(f *flag.Flag) {
			// skip test flags
			if !strings.HasPrefix(f.Name, ignoreFlagPrefix) {
				c.Flags = append(c.Flags, &extFlag{f})
			}
		})
	}

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

	c.categories = newCommandCategories()
	for _, command := range c.Commands {
		c.categories.AddCommand(command.Category, command)
	}
	sort.Sort(c.categories.(*commandCategories))

	var newCmds []*Command
	for _, scmd := range c.Commands {
		cname := scmd.Name
		if scmd.HelpName != "" {
			cname = scmd.HelpName
		}
		scmd.HelpName = fmt.Sprintf("%s %s", c.HelpName, cname)
		newCmds = append(newCmds, scmd)
	}
	c.Commands = newCmds

	if c.Command(helpCommand.Name) == nil && !c.HideHelp {
		if !c.HideHelpCommand {
			helpCommand.HelpName = fmt.Sprintf("%s %s", c.HelpName, helpName)
			c.appendCommand(helpCommand)
		}

		if HelpFlag != nil {
			c.appendFlag(HelpFlag)
		}
	}

	if c.isRoot && !c.HideVersion {
		c.appendFlag(VersionFlag)
	}

	c.categories = newCommandCategories()
	for _, command := range c.Commands {
		c.categories.AddCommand(command.Category, command)
	}
	sort.Sort(c.categories.(*commandCategories))

	c.flagCategories = newFlagCategories()
	for _, fl := range c.Flags {
		if cf, ok := fl.(CategorizableFlag); ok {
			if cf.GetCategory() != "" {
				c.flagCategories.AddFlag(cf.GetCategory(), cf)
			}
		}
	}

	if len(c.SliceFlagSeparator) != 0 {
		defaultSliceFlagSeparator = c.SliceFlagSeparator
	}

	disableSliceFlagSeparator = c.DisableSliceFlagSeparator
}

// Run is the entry point to the cli app. Parses the arguments slice and routes
// to the proper flag/args combination
func (c *Command) Run(arguments []string) error {
	return c.RunContext(context.Background(), arguments)
}

// RunContext is like Run except it takes a Context that will be
// passed to its commands and sub-commands. Through this, you can
// propagate timeouts and cancellation requests
func (c *Command) RunContext(ctx context.Context, arguments []string) (err error) {
	c.isRoot = true
	c.setupDefaults()

	// handle the completion flag separately from the flagset since
	// completion could be attempted after a flag, but before its value was put
	// on the command line. this causes the flagset to interpret the completion
	// flag name as the value of the flag before it which is undesirable
	// note that we can only do this because the shell autocomplete function
	// always appends the completion flag at the end of the command
	shellComplete, arguments := checkShellCompleteFlag(c, arguments)

	cCtx := NewContext(c, nil, newRootContext(ctx, shellComplete))

	return c.run(cCtx, arguments...)
}

func (c *Command) run(cCtx *Context, arguments ...string) (err error) {
	c.setup(cCtx)

	a := args(arguments)
	set, err := c.parseFlags(&a, cCtx)
	cCtx.flagSet = set

	if checkCompletions(cCtx) {
		return nil
	}

	if err != nil {
		if c.OnUsageError != nil {
			err = c.OnUsageError(cCtx, err, !c.isRoot)
			cCtx.Command.handleExitCoder(cCtx, err)
			return err
		}
		_, _ = fmt.Fprintf(cCtx.Command.Writer, "%s %s\n\n", "Incorrect Usage:", err.Error())
		if cCtx.root().Command.Suggest {
			if suggestion, err := c.suggestFlagFromError(err, ""); err == nil {
				fmt.Fprintf(cCtx.Command.Writer, "%s", suggestion)
			}
		}
		if !c.HideHelp {
			if c.isRoot {
				_ = ShowAppHelp(cCtx)
			} else {
				_ = ShowCommandHelp(cCtx.parentContext, c.Name)
			}
		}
		return err
	}

	if checkHelp(cCtx) {
		return helpCommand.Action(cCtx)
	}

	if c.isRoot && !cCtx.Command.HideVersion && checkVersion(cCtx) {
		ShowVersion(cCtx)
		return nil
	}

	if c.After != nil && !cCtx.shellComplete {
		defer func() {
			afterErr := c.After(cCtx)
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

	cerr := cCtx.checkRequiredFlags(c.Flags)
	if cerr != nil {
		_ = ShowSubcommandHelp(cCtx)
		return cerr
	}

	for _, grp := range c.MutuallyExclusiveFlags {
		if err := grp.check(cCtx); err != nil {
			_ = ShowSubcommandHelp(cCtx)
			return err
		}
	}

	if c.Before != nil && !cCtx.shellComplete {
		beforeErr := c.Before(cCtx)
		if beforeErr != nil {
			cCtx.Command.handleExitCoder(cCtx, beforeErr)
			err = beforeErr
			return err
		}
	}

	if err = runFlagActions(cCtx, c.Flags); err != nil {
		return err
	}

	var cmd *Command
	args := cCtx.Args()
	if args.Present() {
		name := args.First()
		if SuggestCommand != nil {
			name = SuggestCommand(c.Commands, name)
		}
		cmd = c.Command(name)
		if cmd == nil {
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
					cmd = cCtx.Command.Command(argsWithDefault.First())
				}
			}
		}
	} else if c.isRoot && cCtx.Command.DefaultCommand != "" {
		if dc := cCtx.Command.Command(cCtx.Command.DefaultCommand); dc != c {
			cmd = dc
		}
	}

	if cmd != nil {
		newcCtx := NewContext(cmd, nil, cCtx)
		return cmd.run(newcCtx, cCtx.Args().Slice()...)
	}

	if c.Action == nil {
		c.Action = helpCommand.Action
	}

	err = c.Action(cCtx)

	cCtx.root().Command.handleExitCoder(cCtx, err)
	return err
}

// Command returns the named command on App. Returns nil if the command does not exist
func (c *Command) Command(name string) *Command {
	for _, cmd := range c.Commands {
		if cmd.HasName(name) {
			return cmd
		}
	}

	return nil
}

func (c *Command) appendCommand(cmd *Command) {
	if !hasCommand(c.Commands, cmd) {
		c.Commands = append(c.Commands, cmd)
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

func (c *Command) parseFlags(args Args, cCtx *Context) (*flag.FlagSet, error) {
	set, err := c.newFlagSet()
	if err != nil {
		return nil, err
	}

	if c.SkipFlagParsing {
		return set, set.Parse(append([]string{"--"}, args.Tail()...))
	}

	for pCtx := cCtx.parentContext; pCtx != nil; pCtx = pCtx.parentContext {
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

	if err := parseIter(set, c, args.Tail(), cCtx.shellComplete); err != nil {
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

func hasCommand(commands []*Command, command *Command) bool {
	for _, existing := range commands {
		if command == existing {
			return true
		}
	}

	return false
}

func runFlagActions(c *Context, fs []Flag) error {
	for _, f := range fs {
		isSet := false
		for _, name := range f.Names() {
			if c.IsSet(name) {
				isSet = true
				break
			}
		}
		if isSet {
			if af, ok := f.(ActionableFlag); ok {
				if err := af.RunAction(c); err != nil {
					return err
				}
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
