package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// App is the main structure of a cli application.
type App struct {
	// The name of the program. Defaults to path.Base(os.Args[0])
	Name string
	// Full name of command for help, defaults to Name
	HelpName string
	// Description of the program.
	Usage string
	// Text to override the USAGE section of help
	UsageText string
	// Description of the program argument format.
	ArgsUsage string
	// Version of the program
	Version string
	// Description of the program
	Description string
	// DefaultCommand is the (optional) name of a command
	// to run if no command names are passed as CLI arguments.
	DefaultCommand string
	// List of commands to execute
	Commands []*Command
	// List of flags to parse
	Flags []Flag
	// Boolean to enable shell completion commands
	EnableShellCompletion bool
	// Shell Completion generation command name
	ShellCompletionCommandName string
	// Boolean to hide built-in help command and help flag
	HideHelp bool
	// Boolean to hide built-in help command but keep help flag.
	// Ignored if HideHelp is true.
	HideHelpCommand bool
	// Boolean to hide built-in version flag and the VERSION section of help
	HideVersion bool
	// categories contains the categorized commands and is populated on app startup
	categories CommandCategories
	// flagCategories contains the categorized flags and is populated on app startup
	flagCategories FlagCategories
	// An action to execute when the shell completion flag is set
	ShellComplete ShellCompleteFunc
	// An action to execute before any subcommands are run, but after the context is ready
	// If a non-nil error is returned, no subcommands are run
	Before BeforeFunc
	// An action to execute after any subcommands are run, but after the subcommand has finished
	// It is run even if Action() panics
	After AfterFunc
	// The action to execute when no subcommands are specified
	Action ActionFunc
	// Execute this function if the proper command cannot be found
	CommandNotFound CommandNotFoundFunc
	// Execute this function if a usage error occurs
	OnUsageError OnUsageErrorFunc
	// Execute this function when an invalid flag is accessed from the context
	InvalidFlagAccessHandler InvalidFlagAccessFunc
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
	// Use longest prefix match for commands
	PrefixMatchCommands bool
	// Custom suggest command for matching
	SuggestCommandFunc SuggestCommandFunc

	didSetup bool

	rootCommand *Command

	// if the app is in error mode
	isInError bool
}

// Setup runs initialization code to ensure all data structures are ready for
// `Run` or inspection prior to `Run`.  It is internally called by `Run`, but
// will return early if setup has already happened.
func (a *App) Setup() {
	if a.didSetup {
		return
	}

	a.didSetup = true

	if a.Name == "" {
		a.Name = filepath.Base(os.Args[0])
	}

	if a.HelpName == "" {
		a.HelpName = a.Name
	}

	if a.Usage == "" {
		a.Usage = "A new cli application"
	}

	if a.Version == "" {
		a.HideVersion = true
	}

	if a.ShellComplete == nil {
		a.ShellComplete = DefaultAppComplete
	}

	if a.Action == nil {
		a.Action = helpCommand.Action
	}

	if a.Reader == nil {
		a.Reader = os.Stdin
	}

	if a.Writer == nil {
		a.Writer = os.Stdout
	}

	if a.ErrWriter == nil {
		a.ErrWriter = os.Stderr
	}

	if a.AllowExtFlags {
		// add global flags added by other packages
		flag.VisitAll(func(f *flag.Flag) {
			// skip test flags
			if !strings.HasPrefix(f.Name, ignoreFlagPrefix) {
				a.Flags = append(a.Flags, &extFlag{f})
			}
		})
	}

	var newCommands []*Command

	for _, c := range a.Commands {
		cname := c.Name
		if c.HelpName != "" {
			cname = c.HelpName
		}
		c.HelpName = fmt.Sprintf("%s %s", a.HelpName, cname)

		c.flagCategories = newFlagCategoriesFromFlags(c.Flags)
		newCommands = append(newCommands, c)
	}
	a.Commands = newCommands

	if a.Command(helpCommand.Name) == nil && !a.HideHelp {
		if !a.HideHelpCommand {
			helpCommand.HelpName = fmt.Sprintf("%s %s", a.HelpName, helpName)
			a.appendCommand(helpCommand)
		}

		if HelpFlag != nil {
			a.appendFlag(HelpFlag)
		}
	}

	if !a.HideVersion {
		a.appendFlag(VersionFlag)
	}

	if a.PrefixMatchCommands {
		if a.SuggestCommandFunc == nil {
			a.SuggestCommandFunc = suggestCommand
		}
	}
	if a.EnableShellCompletion {
		if a.ShellCompletionCommandName != "" {
			completionCommand.Name = a.ShellCompletionCommandName
		}
		a.appendCommand(completionCommand)
	}

	a.categories = newCommandCategories()
	for _, command := range a.Commands {
		a.categories.AddCommand(command.Category, command)
	}
	sort.Sort(a.categories.(*commandCategories))

	a.flagCategories = newFlagCategories()
	for _, fl := range a.Flags {
		if cf, ok := fl.(CategorizableFlag); ok {
			if cf.GetCategory() != "" {
				a.flagCategories.AddFlag(cf.GetCategory(), cf)
			}
		}
	}

	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}

	if len(a.SliceFlagSeparator) != 0 {
		defaultSliceFlagSeparator = a.SliceFlagSeparator
	}

	disableSliceFlagSeparator = a.DisableSliceFlagSeparator
}

/*
func (a *App) newRootCommand() *Command {
	return &Command{
		Name:                   a.Name,
		Usage:                  a.Usage,
		UsageText:              a.UsageText,
		Description:            a.Description,
		ArgsUsage:              a.ArgsUsage,
		ShellComplete:          a.ShellComplete,
		Before:                 a.Before,
		After:                  a.After,
		Action:                 a.Action,
		OnUsageError:           a.OnUsageError,
		Commands:               a.Commands,
		Flags:                  a.Flags,
		flagCategories:         a.flagCategories,
		HideHelp:               a.HideHelp,
		HideHelpCommand:        a.HideHelpCommand,
		UseShortOptionHandling: a.UseShortOptionHandling,
		HelpName:               a.HelpName,
		CustomHelpTemplate:     a.CustomAppHelpTemplate,
		categories:             a.categories,
		SkipFlagParsing:        a.SkipFlagParsing,
		isRoot:                 true,
		MutuallyExclusiveFlags: a.MutuallyExclusiveFlags,
		PrefixMatchCommands:    a.PrefixMatchCommands,
	}
}
*/

func (a *App) newFlagSet() (*flag.FlagSet, error) {
	return flagSet(a.Name, a.Flags)
}

func (a *App) useShortOptionHandling() bool {
	return a.UseShortOptionHandling
}

// Run is the entry point to the cli app. Parses the arguments slice and routes
// to the proper flag/args combination
func (a *App) Run(arguments []string) error {
	return a.RunContext(context.Background(), arguments)
}

// RunContext is like Run except it takes a Context that will be
// passed to its commands and sub-commands. Through this, you can
// propagate timeouts and cancellation requests
func (a *App) RunContext(ctx context.Context, arguments []string) (err error) {
	a.Setup()

	/*
		// handle the completion flag separately from the flagset since
		// completion could be attempted after a flag, but before its value was put
		// on the command line. this causes the flagset to interpret the completion
		// flag name as the value of the flag before it which is undesirable
		// note that we can only do this because the shell autocomplete function
		// always appends the completion flag at the end of the command
		shellComplete, arguments := checkShellCompleteFlag(a, arguments)

		cCtx := NewContext(a, nil, &Context{Context: ctx})
		cCtx.shellComplete = shellComplete

		a.rootCommand = a.newRootCommand()
		cCtx.Command = a.rootCommand
	*/

	return a.rootCommand.Run(ctx, arguments)
}

func (a *App) suggestFlagFromError(err error, command string) (string, error) {
	flag, parseErr := flagFromError(err)
	if parseErr != nil {
		return "", err
	}

	flags := a.Flags
	hideHelp := a.HideHelp
	if command != "" {
		cmd := a.Command(command)
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

	return fmt.Sprintf(SuggestDidYouMeanTemplate+"\n\n", suggestion), nil
}

// Command returns the named command on App. Returns nil if the command does not exist
func (a *App) Command(name string) *Command {
	for _, c := range a.Commands {
		if c.HasName(name) {
			return c
		}
	}

	return nil
}

// VisibleCategories returns a slice of categories and commands that are
// Hidden=false
func (a *App) VisibleCategories() []CommandCategory {
	ret := []CommandCategory{}
	for _, category := range a.categories.Categories() {
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
func (a *App) VisibleCommands() []*Command {
	var ret []*Command
	for _, command := range a.Commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}

// VisibleFlagCategories returns a slice containing all the categories with the flags they contain
func (a *App) VisibleFlagCategories() []VisibleFlagCategory {
	if a.flagCategories == nil {
		return []VisibleFlagCategory{}
	}
	return a.flagCategories.VisibleCategories()
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (a *App) VisibleFlags() []Flag {
	return visibleFlags(a.Flags)
}

func (a *App) appendFlag(fl Flag) {
	if !hasFlag(a.Flags, fl) {
		a.Flags = append(a.Flags, fl)
	}
}

func (a *App) appendCommand(c *Command) {
	if !hasCommand(a.Commands, c) {
		a.Commands = append(a.Commands, c)
	}
}

func (a *App) handleExitCoder(cCtx *Context, err error) {
	if a.ExitErrHandler != nil {
		a.ExitErrHandler(cCtx, err)
	} else {
		HandleExitCoder(err)
	}
}

func (a *App) argsWithDefaultCommand(oldArgs Args) Args {
	if a.DefaultCommand != "" {
		rawArgs := append([]string{a.DefaultCommand}, oldArgs.Slice()...)
		newArgs := args(rawArgs)

		return &newArgs
	}

	return oldArgs
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

func (a *App) writer() io.Writer {
	if a.isInError {
		// this can happen in test but not in normal usage
		if a.ErrWriter == nil {
			return os.Stderr
		}
		return a.ErrWriter
	}
	return a.Writer
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
