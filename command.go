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

const (
	// ignoreFlagPrefix is to ignore test flags when adding flags from other packages
	ignoreFlagPrefix = "test."

	commandContextKey = contextKey("cli.context")
)

type contextKey string

// Command contains everything needed to run an application that
// accepts a string slice of arguments such as os.Args. A given
// Command may contain Flags and sub-commands in Commands.
type Command struct {
	// The name of the command
	Name string
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
	// List of flags to parse
	Flags []Flag
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
	// CustomRootCommandHelpTemplate the text template for app help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomRootCommandHelpTemplate string
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
	// CustomHelpTemplate the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomHelpTemplate string
	// Use longest prefix match for commands
	PrefixMatchCommands bool
	// Custom suggest command for matching
	SuggestCommandFunc SuggestCommandFunc
	// Flag exclusion group
	MutuallyExclusiveFlags []MutuallyExclusiveFlags

	// categories contains the categorized commands and is populated on app startup
	categories CommandCategories
	// flagCategories contains the categorized flags and is populated on app startup
	flagCategories FlagCategories
	// flags that have been applied in current parse
	appliedFlags []Flag
	// The parent of this command. This value will be nil for the
	// command at the root of the graph.
	parent *Command
	// the flag.FlagSet for this command
	flagSet *flag.FlagSet
	// track state of error handling
	isInError bool
	// track state of defaults
	didSetupDefaults bool
}

// FullName returns the full name of the command.
// For commands with parents this ensures that the parent commands
// are part of the command path.
func (cmd *Command) FullName() string {
	namePath := []string{}

	if cmd.parent != nil {
		namePath = append(namePath, cmd.parent.FullName())
	}

	return strings.Join(append(namePath, cmd.Name), " ")
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
	if cmd.didSetupDefaults {
		tracef("already did setup")
		return
	}

	cmd.didSetupDefaults = true

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

func (cmd *Command) setupCommandGraph() {
	for _, subCmd := range cmd.Commands {
		subCmd.parent = cmd
		subCmd.setupSubcommand()
		subCmd.setupCommandGraph()
	}
}

func (cmd *Command) setupSubcommand() {
	cmd.ensureHelp()

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

	if v, ok := ctx.Value(commandContextKey).(*Command); ok {
		cmd.parent = v
	}

	// handle the completion flag separately from the flagset since
	// completion could be attempted after a flag, but before its value was put
	// on the command line. this causes the flagset to interpret the completion
	// flag name as the value of the flag before it which is undesirable
	// note that we can only do this because the shell autocomplete function
	// always appends the completion flag at the end of the command
	enableShellCompletion, arguments := checkShellCompleteFlag(cmd, arguments)
	cmd.EnableShellCompletion = enableShellCompletion

	ctx = context.WithValue(ctx, commandContextKey, cmd)

	if cmd.parent == nil {
		cmd.setupCommandGraph()
	}

	argsArguments := args(arguments)
	set, err := cmd.parseFlags(&argsArguments)
	cmd.flagSet = set

	if checkCompletions(ctx, cmd) {
		return nil
	}

	if err != nil {
		tracef("setting deferErr from %[1]v", err)
		deferErr = err

		cmd.isInError = true
		if cmd.OnUsageError != nil {
			err = cmd.OnUsageError(ctx, cmd, err, cmd.parent != nil)
			err = cmd.handleExitCoder(ctx, err)
			return err
		}
		fmt.Fprintf(cmd.Root().ErrWriter, "Incorrect Usage: %s\n\n", err.Error())
		if cmd.Suggest {
			if suggestion, err := cmd.suggestFlagFromError(err, ""); err == nil {
				fmt.Fprintf(cmd.Root().ErrWriter, "%s", suggestion)
			}
		}
		if !cmd.HideHelp {
			if cmd.parent == nil {
				tracef("running ShowAppHelp")
				if err := ShowAppHelp(cmd); err != nil {
					tracef("SILENTLY IGNORING ERROR running ShowAppHelp %[1]v", err)
				}
			} else {
				tracef("running ShowCommandHelp with %[1]q", cmd.Name)
				if err := ShowCommandHelp(ctx, cmd, cmd.Name); err != nil {
					tracef("SILENTLY IGNORING ERROR running ShowCommandHelp with %[1]q %[2]v", cmd.Name, err)
				}
			}
		}

		return err
	}

	if checkHelp(cmd) {
		return helpCommandAction(ctx, cmd)
	}

	if cmd.parent == nil && !cmd.HideVersion && checkVersion(cmd) {
		ShowVersion(cmd)
		return nil
	}

	if cmd.After != nil && !cmd.EnableShellCompletion {
		defer func() {
			if err := cmd.After(ctx, cmd); err != nil {
				err = cmd.handleExitCoder(ctx, err)

				if deferErr != nil {
					deferErr = newMultiError(deferErr, err)
				} else {
					deferErr = err
				}
			}
		}()
	}

	if err := cmd.checkRequiredFlags(); err != nil {
		cmd.isInError = true
		_ = ShowSubcommandHelp(cmd)
		return err
	}

	for _, grp := range cmd.MutuallyExclusiveFlags {
		if err := grp.check(cmd); err != nil {
			_ = ShowSubcommandHelp(cmd)
			return err
		}
	}

	if cmd.Before != nil && !cmd.EnableShellCompletion {
		if err := cmd.Before(ctx, cmd); err != nil {
			deferErr = cmd.handleExitCoder(ctx, err)
			return deferErr
		}
	}

	if err := runFlagActions(ctx, cmd, cmd.appliedFlags); err != nil {
		return err
	}

	var subCmd *Command
	if argsArguments.Present() {
		name := argsArguments.First()
		if cmd.SuggestCommandFunc != nil {
			name = cmd.SuggestCommandFunc(cmd.Commands, name)
		}
		subCmd = cmd.Command(name)
		if subCmd == nil {
			hasDefault := cmd.DefaultCommand != ""
			isFlagName := checkStringSliceIncludes(name, cmd.FlagNames())

			var (
				isDefaultSubcommand   = false
				defaultHasSubcommands = false
			)

			if hasDefault {
				dc := cmd.Command(cmd.DefaultCommand)
				defaultHasSubcommands = len(dc.Commands) > 0
				for _, dcSub := range dc.Commands {
					if checkStringSliceIncludes(name, dcSub.Names()) {
						isDefaultSubcommand = true
						break
					}
				}
			}

			if isFlagName || (hasDefault && (defaultHasSubcommands && isDefaultSubcommand)) {
				argsWithDefault := cmd.argsWithDefaultCommand(&argsArguments)
				if !reflect.DeepEqual(argsArguments, argsWithDefault) {
					subCmd = cmd.Command(argsWithDefault.First())
				}
			}
		}
	} else if cmd.parent == nil && cmd.DefaultCommand != "" {
		if dc := cmd.Command(cmd.DefaultCommand); dc != cmd {
			subCmd = dc
		}
	}

	if subCmd != nil {
		return subCmd.Run(ctx, cmd.Args().Slice())
	}

	if cmd.Action == nil {
		cmd.Action = helpCommandAction
	}

	if err := cmd.Action(ctx, cmd); err != nil {
		tracef("calling handleExitCoder with %[1]v", err)
		deferErr = cmd.handleExitCoder(ctx, err)
	}

	tracef("returning deferErr")
	return deferErr
}

func (cmd *Command) newFlagSet() (*flag.FlagSet, error) {
	cmd.appliedFlags = append(cmd.appliedFlags, cmd.allFlags()...)
	return flagSet(cmd.Name, cmd.allFlags())
}

func (cmd *Command) allFlags() []Flag {
	var flags []Flag
	flags = append(flags, cmd.Flags...)
	for _, grpf := range cmd.MutuallyExclusiveFlags {
		for _, f1 := range grpf.Flags {
			flags = append(flags, f1...)
		}
	}
	return flags
}

func (cmd *Command) useShortOptionHandling() bool {
	return cmd.UseShortOptionHandling
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

func (cmd *Command) parseFlags(argsArguments Args) (*flag.FlagSet, error) {
	set, err := cmd.newFlagSet()
	if err != nil {
		return nil, err
	}

	if cmd.SkipFlagParsing {
		return set, set.Parse(append([]string{"--"}, argsArguments.Tail()...))
	}

	for pCmd := cmd.parent; pCmd != nil; pCmd = pCmd.parent {
		for _, fl := range pCmd.Flags {
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

			cmd.appliedFlags = append(cmd.appliedFlags, fl)
		}
	}

	if err := parseIter(set, cmd, argsArguments.Tail(), cmd.EnableShellCompletion); err != nil {
		return nil, err
	}

	if err := normalizeFlags(cmd.Flags, set); err != nil {
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
func (cmd *Command) VisibleCategories() []CommandCategory {
	ret := []CommandCategory{}
	for _, category := range cmd.categories.Categories() {
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
func (cmd *Command) VisibleCommands() []*Command {
	var ret []*Command
	for _, command := range cmd.Commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}

// VisibleFlagCategories returns a slice containing all the visible flag categories with the flags they contain
func (cmd *Command) VisibleFlagCategories() []VisibleFlagCategory {
	if cmd.flagCategories == nil {
		cmd.flagCategories = newFlagCategoriesFromFlags(cmd.Flags)
	}
	return cmd.flagCategories.VisibleCategories()
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (cmd *Command) VisibleFlags() []Flag {
	return visibleFlags(cmd.Flags)
}

func (cmd *Command) appendFlag(fl Flag) {
	if !hasFlag(cmd.Flags, fl) {
		cmd.Flags = append(cmd.Flags, fl)
	}
}

func (cmd *Command) appendCommand(aCmd *Command) {
	if !hasCommand(cmd.Commands, aCmd) {
		aCmd.parent = cmd
		cmd.Commands = append(cmd.Commands, aCmd)
	}
}

func (cmd *Command) handleExitCoder(ctx context.Context, err error) error {
	if cmd.parent != nil {
		return cmd.parent.handleExitCoder(ctx, err)
	}

	if cmd.ExitErrHandler != nil {
		cmd.ExitErrHandler(ctx, cmd, err)
		return err
	}

	HandleExitCoder(err)
	return err
}

func (cmd *Command) argsWithDefaultCommand(oldArgs Args) Args {
	if cmd.DefaultCommand != "" {
		rawArgs := append([]string{cmd.DefaultCommand}, oldArgs.Slice()...)
		newArgs := args(rawArgs)

		return &newArgs
	}

	return oldArgs
}

// Root returns the Command at the root of the graph
func (cmd *Command) Root() *Command {
	if cmd.parent == nil {
		return cmd
	}

	return cmd.parent.Root()
}

func (cmd *Command) lookupFlag(name string) Flag {
	for _, pCmd := range cmd.Lineage() {
		for _, f := range pCmd.Flags {
			for _, n := range f.Names() {
				if n == name {
					return f
				}
			}
		}
	}

	return nil
}

func (cmd *Command) lookupFlagSet(name string) *flag.FlagSet {
	for _, pCmd := range cmd.Lineage() {
		if pCmd.flagSet == nil {
			continue
		}
		if f := pCmd.flagSet.Lookup(name); f != nil {
			return pCmd.flagSet
		}
	}
	cmd.onInvalidFlag(context.TODO(), name)
	return nil
}

func (cmd *Command) checkRequiredFlags() requiredFlagsErr {
	var missingFlags []string
	for _, f := range cmd.Flags {
		if rf, ok := f.(RequiredFlag); ok && rf.IsRequired() {
			var flagPresent bool
			var flagName string

			for _, key := range f.Names() {
				flagName = key

				if cmd.IsSet(strings.TrimSpace(key)) {
					flagPresent = true
				}
			}

			if !flagPresent && flagName != "" {
				missingFlags = append(missingFlags, flagName)
			}
		}
	}

	if len(missingFlags) != 0 {
		return &errRequiredFlags{missingFlags: missingFlags}
	}

	return nil
}

func (cmd *Command) onInvalidFlag(ctx context.Context, name string) {
	for cmd != nil {
		if cmd.InvalidFlagAccessHandler != nil {
			cmd.InvalidFlagAccessHandler(ctx, cmd, name)
			break
		}
		cmd = cmd.parent
	}
}

// NumFlags returns the number of flags set
func (cmd *Command) NumFlags() int {
	return cmd.flagSet.NFlag()
}

// Set sets a context flag to a value.
func (cmd *Command) Set(name, value string) error {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		return fs.Set(name, value)
	}

	return fmt.Errorf("no such flag -%s", name)
}

// IsSet determines if the flag was actually set
func (cmd *Command) IsSet(name string) bool {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		isSet := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == name {
				isSet = true
			}
		})
		if isSet {
			return true
		}

		f := cmd.lookupFlag(name)
		if f == nil {
			return false
		}

		return f.IsSet()
	}

	return false
}

// LocalFlagNames returns a slice of flag names used in this
// command.
func (cmd *Command) LocalFlagNames() []string {
	var names []string
	cmd.flagSet.Visit(makeFlagNameVisitor(&names))
	// Check the flags which have been set via env or file
	if cmd.Flags != nil {
		for _, f := range cmd.Flags {
			if f.IsSet() {
				names = append(names, f.Names()...)
			}
		}
	}

	// Sort out the duplicates since flag could be set via multiple
	// paths
	m := map[string]struct{}{}
	var unames []string
	for _, name := range names {
		if _, ok := m[name]; !ok {
			m[name] = struct{}{}
			unames = append(unames, name)
		}
	}

	return unames
}

// FlagNames returns a slice of flag names used by the this command
// and all of its parent commands.
func (cmd *Command) FlagNames() []string {
	var names []string
	for _, pCmd := range cmd.Lineage() {
		names = append(names, pCmd.LocalFlagNames()...)
	}
	return names
}

// Lineage returns *this* command and all of its ancestor commands
// in order from child to parent
func (cmd *Command) Lineage() []*Command {
	var lineage []*Command

	for cur := cmd; cur != nil; cur = cur.parent {
		lineage = append(lineage, cur)
	}

	return lineage
}

// Count returns the num of occurrences of this flag
func (cmd *Command) Count(name string) int {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		if cf, ok := fs.Lookup(name).Value.(Countable); ok {
			return cf.Count()
		}
	}
	return 0
}

// Value returns the value of the flag corresponding to `name`
func (cmd *Command) Value(name string) interface{} {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		return fs.Lookup(name).Value.(flag.Getter).Get()
	}
	return nil
}

// Args returns the command line arguments associated with the
// command.
func (cmd *Command) Args() Args {
	ret := args(cmd.flagSet.Args())
	return &ret
}

// NArg returns the number of the command line arguments.
func (cmd *Command) NArg() int {
	return cmd.Args().Len()
}

func hasCommand(commands []*Command, command *Command) bool {
	for _, existing := range commands {
		if command == existing {
			return true
		}
	}

	return false
}

func runFlagActions(ctx context.Context, cmd *Command, flags []Flag) error {
	for _, fl := range flags {
		isSet := false

		for _, name := range fl.Names() {
			if cmd.IsSet(name) {
				isSet = true
				break
			}
		}

		if !isSet {
			continue
		}

		if af, ok := fl.(ActionableFlag); ok {
			if err := af.RunAction(ctx, cmd); err != nil {
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

func makeFlagNameVisitor(names *[]string) func(*flag.Flag) {
	return func(f *flag.Flag) {
		nameParts := strings.Split(f.Name, ",")
		name := strings.TrimSpace(nameParts[0])

		for _, part := range nameParts {
			part = strings.TrimSpace(part)
			if len(part) > len(name) {
				name = part
			}
		}

		if name != "" {
			*names = append(*names, name)
		}
	}
}
