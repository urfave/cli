package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"
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
	Name string `json:"name"`
	// A list of aliases for the command
	Aliases []string `json:"aliases"`
	// A short description of the usage of this command
	Usage string `json:"usage"`
	// Text to override the USAGE section of help
	UsageText string `json:"usageText"`
	// A short description of the arguments of this command
	ArgsUsage string `json:"argsUsage"`
	// Version of the command
	Version string `json:"version"`
	// Longer explanation of how the command works
	Description string `json:"description"`
	// DefaultCommand is the (optional) name of a command
	// to run if no command names are passed as CLI arguments.
	DefaultCommand string `json:"defaultCommand"`
	// The category the command is part of
	Category string `json:"category"`
	// List of child commands
	Commands []*Command `json:"commands"`
	// List of flags to parse
	Flags []Flag `json:"flags"`
	// Boolean to hide built-in help command and help flag
	HideHelp bool `json:"hideHelp"`
	// Ignored if HideHelp is true.
	HideHelpCommand bool `json:"hideHelpCommand"`
	// Boolean to hide built-in version flag and the VERSION section of help
	HideVersion bool `json:"hideVersion"`
	// Boolean to enable shell completion commands
	EnableShellCompletion bool `json:"-"`
	// Shell Completion generation command name
	ShellCompletionCommandName string `json:"-"`
	// The function to call when checking for shell command completions
	ShellComplete ShellCompleteFunc `json:"-"`
	// An action to execute before any subcommands are run, but after the context is ready
	// If a non-nil error is returned, no subcommands are run
	Before BeforeFunc `json:"-"`
	// An action to execute after any subcommands are run, but after the subcommand has finished
	// It is run even if Action() panics
	After AfterFunc `json:"-"`
	// The function to call when this command is invoked
	Action ActionFunc `json:"-"`
	// Execute this function if the proper command cannot be found
	CommandNotFound CommandNotFoundFunc `json:"-"`
	// Execute this function if a usage error occurs.
	OnUsageError OnUsageErrorFunc `json:"-"`
	// Execute this function when an invalid flag is accessed from the context
	InvalidFlagAccessHandler InvalidFlagAccessFunc `json:"-"`
	// Boolean to hide this command from help or completion
	Hidden bool `json:"hidden"`
	// List of all authors who contributed (string or fmt.Stringer)
	// TODO: ~string | fmt.Stringer when interface unions are available
	Authors []any `json:"authors"`
	// Copyright of the binary if any
	Copyright string `json:"copyright"`
	// Reader reader to write input to (useful for tests)
	Reader io.Reader `json:"-"`
	// Writer writer to write output to
	Writer io.Writer `json:"-"`
	// ErrWriter writes error output
	ErrWriter io.Writer `json:"-"`
	// ExitErrHandler processes any error encountered while running an App before
	// it is returned to the caller. If no function is provided, HandleExitCoder
	// is used as the default behavior.
	ExitErrHandler ExitErrHandlerFunc `json:"-"`
	// Other custom info
	Metadata map[string]interface{} `json:"metadata"`
	// Carries a function which returns app specific info.
	ExtraInfo func() map[string]string `json:"-"`
	// CustomRootCommandHelpTemplate the text template for app help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomRootCommandHelpTemplate string `json:"-"`
	// SliceFlagSeparator is used to customize the separator for SliceFlag, the default is ","
	SliceFlagSeparator string `json:"sliceFlagSeparator"`
	// DisableSliceFlagSeparator is used to disable SliceFlagSeparator, the default is false
	DisableSliceFlagSeparator bool `json:"disableSliceFlagSeparator"`
	// Boolean to enable short-option handling so user can combine several
	// single-character bool arguments into one
	// i.e. foobar -o -v -> foobar -ov
	UseShortOptionHandling bool `json:"useShortOptionHandling"`
	// Enable suggestions for commands and flags
	Suggest bool `json:"suggest"`
	// Allows global flags set by libraries which use flag.XXXVar(...) directly
	// to be parsed through this library
	AllowExtFlags bool `json:"allowExtFlags"`
	// Treat all flags as normal arguments if true
	SkipFlagParsing bool `json:"skipFlagParsing"`
	// CustomHelpTemplate the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomHelpTemplate string `json:"-"`
	// Use longest prefix match for commands
	PrefixMatchCommands bool `json:"prefixMatchCommands"`
	// Custom suggest command for matching
	SuggestCommandFunc SuggestCommandFunc `json:"-"`
	// Flag exclusion group
	MutuallyExclusiveFlags []MutuallyExclusiveFlags `json:"mutuallyExclusiveFlags"`
	// Arguments to parse for this command
	Arguments []Argument `json:"arguments"`
	// Whether to read arguments from stdin
	// applicable to root command only
	ReadArgsFromStdin bool `json:"readArgsFromStdin"`

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
	// parsed args
	parsedArgs Args
	// track state of error handling
	isInError bool
	// track state of defaults
	didSetupDefaults bool
	// whether in shell completion mode
	shellCompletion bool
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

func (cmd *Command) setupDefaults(osArgs []string) {
	if cmd.didSetupDefaults {
		tracef("already did setup (cmd=%[1]q)", cmd.Name)
		return
	}

	cmd.didSetupDefaults = true

	isRoot := cmd.parent == nil
	tracef("isRoot? %[1]v (cmd=%[2]q)", isRoot, cmd.Name)

	if cmd.ShellComplete == nil {
		tracef("setting default ShellComplete (cmd=%[1]q)", cmd.Name)
		cmd.ShellComplete = DefaultCompleteWithFlags
	}

	if cmd.Name == "" && isRoot {
		name := filepath.Base(osArgs[0])
		tracef("setting cmd.Name from first arg basename (cmd=%[1]q)", name)
		cmd.Name = name
	}

	if cmd.Usage == "" && isRoot {
		tracef("setting default Usage (cmd=%[1]q)", cmd.Name)
		cmd.Usage = "A new cli application"
	}

	if cmd.Version == "" {
		tracef("setting HideVersion=true due to empty Version (cmd=%[1]q)", cmd.Name)
		cmd.HideVersion = true
	}

	if cmd.Action == nil {
		tracef("setting default Action as help command action (cmd=%[1]q)", cmd.Name)
		cmd.Action = helpCommandAction
	}

	if cmd.Reader == nil {
		tracef("setting default Reader as os.Stdin (cmd=%[1]q)", cmd.Name)
		cmd.Reader = os.Stdin
	}

	if cmd.Writer == nil {
		tracef("setting default Writer as os.Stdout (cmd=%[1]q)", cmd.Name)
		cmd.Writer = os.Stdout
	}

	if cmd.ErrWriter == nil {
		tracef("setting default ErrWriter as os.Stderr (cmd=%[1]q)", cmd.Name)
		cmd.ErrWriter = os.Stderr
	}

	if cmd.AllowExtFlags {
		tracef("visiting all flags given AllowExtFlags=true (cmd=%[1]q)", cmd.Name)
		// add global flags added by other packages
		flag.VisitAll(func(f *flag.Flag) {
			// skip test flags
			if !strings.HasPrefix(f.Name, ignoreFlagPrefix) {
				cmd.Flags = append(cmd.Flags, &extFlag{f})
			}
		})
	}

	for _, subCmd := range cmd.Commands {
		tracef("setting sub-command (cmd=%[1]q) parent as self (cmd=%[2]q)", subCmd.Name, cmd.Name)
		subCmd.parent = cmd
	}

	cmd.ensureHelp()

	if !cmd.HideVersion && isRoot {
		tracef("appending version flag (cmd=%[1]q)", cmd.Name)
		cmd.appendFlag(VersionFlag)
	}

	if cmd.PrefixMatchCommands && cmd.SuggestCommandFunc == nil {
		tracef("setting default SuggestCommandFunc (cmd=%[1]q)", cmd.Name)
		cmd.SuggestCommandFunc = suggestCommand
	}

	if cmd.EnableShellCompletion || cmd.Root().shellCompletion {
		completionCommand := buildCompletionCommand(cmd.Name)

		if cmd.ShellCompletionCommandName != "" {
			tracef(
				"setting completion command name (%[1]q) from "+
					"cmd.ShellCompletionCommandName (cmd=%[2]q)",
				cmd.ShellCompletionCommandName, cmd.Name,
			)
			completionCommand.Name = cmd.ShellCompletionCommandName
		}

		tracef("appending completionCommand (cmd=%[1]q)", cmd.Name)
		cmd.appendCommand(completionCommand)
	}

	tracef("setting command categories (cmd=%[1]q)", cmd.Name)
	cmd.categories = newCommandCategories()

	for _, subCmd := range cmd.Commands {
		cmd.categories.AddCommand(subCmd.Category, subCmd)
	}

	tracef("sorting command categories (cmd=%[1]q)", cmd.Name)
	sort.Sort(cmd.categories.(*commandCategories))

	tracef("setting category on mutually exclusive flags (cmd=%[1]q)", cmd.Name)
	for _, grp := range cmd.MutuallyExclusiveFlags {
		grp.propagateCategory()
	}

	tracef("setting flag categories (cmd=%[1]q)", cmd.Name)
	cmd.flagCategories = newFlagCategoriesFromFlags(cmd.allFlags())

	if cmd.Metadata == nil {
		tracef("setting default Metadata (cmd=%[1]q)", cmd.Name)
		cmd.Metadata = map[string]any{}
	}

	if len(cmd.SliceFlagSeparator) != 0 {
		tracef("setting defaultSliceFlagSeparator from cmd.SliceFlagSeparator (cmd=%[1]q)", cmd.Name)
		defaultSliceFlagSeparator = cmd.SliceFlagSeparator
	}

	tracef("setting disableSliceFlagSeparator from cmd.DisableSliceFlagSeparator (cmd=%[1]q)", cmd.Name)
	disableSliceFlagSeparator = cmd.DisableSliceFlagSeparator
}

func (cmd *Command) setupCommandGraph() {
	tracef("setting up command graph (cmd=%[1]q)", cmd.Name)

	for _, subCmd := range cmd.Commands {
		subCmd.parent = cmd
		subCmd.setupSubcommand()
		subCmd.setupCommandGraph()
	}
}

func (cmd *Command) setupSubcommand() {
	tracef("setting up self as sub-command (cmd=%[1]q)", cmd.Name)

	cmd.ensureHelp()

	tracef("setting command categories (cmd=%[1]q)", cmd.Name)
	cmd.categories = newCommandCategories()

	for _, subCmd := range cmd.Commands {
		cmd.categories.AddCommand(subCmd.Category, subCmd)
	}

	tracef("sorting command categories (cmd=%[1]q)", cmd.Name)
	sort.Sort(cmd.categories.(*commandCategories))

	tracef("setting category on mutually exclusive flags (cmd=%[1]q)", cmd.Name)
	for _, grp := range cmd.MutuallyExclusiveFlags {
		grp.propagateCategory()
	}

	tracef("setting flag categories (cmd=%[1]q)", cmd.Name)
	cmd.flagCategories = newFlagCategoriesFromFlags(cmd.allFlags())
}

func (cmd *Command) hideHelp() bool {
	tracef("hide help (cmd=%[1]q)", cmd.Name)
	for c := cmd; c != nil; c = c.parent {
		if c.HideHelp {
			return true
		}
	}

	return false
}

func (cmd *Command) ensureHelp() {
	tracef("ensuring help (cmd=%[1]q)", cmd.Name)

	helpCommand := buildHelpCommand(true)

	if !cmd.hideHelp() {
		if cmd.Command(helpCommand.Name) == nil {
			if !cmd.HideHelpCommand {
				tracef("appending helpCommand (cmd=%[1]q)", cmd.Name)
				cmd.appendCommand(helpCommand)
			}
		}

		if HelpFlag != nil {
			tracef("appending HelpFlag (cmd=%[1]q)", cmd.Name)
			cmd.appendFlag(HelpFlag)
		}
	}
}

func (cmd *Command) parseArgsFromStdin() ([]string, error) {
	type state int
	const (
		stateSearchForToken  state = -1
		stateSearchForString state = 0
	)

	st := stateSearchForToken
	linenum := 1
	token := ""
	args := []string{}

	breader := bufio.NewReader(cmd.Reader)

outer:
	for {
		ch, _, err := breader.ReadRune()
		if err == io.EOF {
			switch st {
			case stateSearchForToken:
				if token != "--" {
					args = append(args, token)
				}
			case stateSearchForString:
				// make sure string is not empty
				for _, t := range token {
					if !unicode.IsSpace(t) {
						args = append(args, token)
					}
				}
			}
			break outer
		}
		if err != nil {
			return nil, err
		}
		switch st {
		case stateSearchForToken:
			if unicode.IsSpace(ch) || ch == '"' {
				if ch == '\n' {
					linenum++
				}
				if token != "" {
					// end the processing here
					if token == "--" {
						break outer
					}
					args = append(args, token)
					token = ""
				}
				if ch == '"' {
					st = stateSearchForString
				}
				continue
			}
			token += string(ch)
		case stateSearchForString:
			if ch != '"' {
				token += string(ch)
			} else {
				if token != "" {
					args = append(args, token)
					token = ""
				}
				/*else {
					//TODO. Should we pass in empty strings ?
				}*/
				st = stateSearchForToken
			}
		}
	}

	tracef("parsed stdin args as %v (cmd=%[2]q)", args, cmd.Name)

	return args, nil
}

// Run is the entry point to the command graph. The positional
// arguments are parsed according to the Flag and Command
// definitions and the matching Action functions are run.
func (cmd *Command) Run(ctx context.Context, osArgs []string) (deferErr error) {
	tracef("running with arguments %[1]q (cmd=%[2]q)", osArgs, cmd.Name)
	cmd.setupDefaults(osArgs)

	if v, ok := ctx.Value(commandContextKey).(*Command); ok {
		tracef("setting parent (cmd=%[1]q) command from context.Context value (cmd=%[2]q)", v.Name, cmd.Name)
		cmd.parent = v
	}

	if cmd.parent == nil {
		if cmd.ReadArgsFromStdin {
			if args, err := cmd.parseArgsFromStdin(); err != nil {
				return err
			} else {
				osArgs = append(osArgs, args...)
			}
		}
		// handle the completion flag separately from the flagset since
		// completion could be attempted after a flag, but before its value was put
		// on the command line. this causes the flagset to interpret the completion
		// flag name as the value of the flag before it which is undesirable
		// note that we can only do this because the shell autocomplete function
		// always appends the completion flag at the end of the command
		tracef("checking osArgs %v (cmd=%[2]q)", osArgs, cmd.Name)
		cmd.shellCompletion, osArgs = checkShellCompleteFlag(cmd, osArgs)

		tracef("setting cmd.shellCompletion=%[1]v from checkShellCompleteFlag (cmd=%[2]q)", cmd.shellCompletion && cmd.EnableShellCompletion, cmd.Name)
		cmd.shellCompletion = cmd.EnableShellCompletion && cmd.shellCompletion
	}

	tracef("using post-checkShellCompleteFlag arguments %[1]q (cmd=%[2]q)", osArgs, cmd.Name)

	tracef("setting self as cmd in context (cmd=%[1]q)", cmd.Name)
	ctx = context.WithValue(ctx, commandContextKey, cmd)

	if cmd.parent == nil {
		cmd.setupCommandGraph()
	}

	args, err := cmd.parseFlags(&stringSliceArgs{v: osArgs})

	tracef("using post-parse arguments %[1]q (cmd=%[2]q)", args, cmd.Name)

	if checkCompletions(ctx, cmd) {
		return nil
	}

	if err != nil {
		tracef("setting deferErr from %[1]q (cmd=%[2]q)", err, cmd.Name)
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
		if !cmd.hideHelp() {
			if cmd.parent == nil {
				tracef("running ShowAppHelp")
				if err := ShowAppHelp(cmd); err != nil {
					tracef("SILENTLY IGNORING ERROR running ShowAppHelp %[1]v (cmd=%[2]q)", err, cmd.Name)
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

	if cmd.checkHelp() {
		return helpCommandAction(ctx, cmd)
	} else {
		tracef("no help is wanted (cmd=%[1]q)", cmd.Name)
	}

	if cmd.parent == nil && !cmd.HideVersion && checkVersion(cmd) {
		ShowVersion(cmd)
		return nil
	}

	if cmd.After != nil && !cmd.Root().shellCompletion {
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

	for _, grp := range cmd.MutuallyExclusiveFlags {
		if err := grp.check(cmd); err != nil {
			_ = ShowSubcommandHelp(cmd)
			return err
		}
	}

	if cmd.Before != nil && !cmd.Root().shellCompletion {
		if bctx, err := cmd.Before(ctx, cmd); err != nil {
			deferErr = cmd.handleExitCoder(ctx, err)
			return deferErr
		} else if bctx != nil {
			ctx = bctx
		}
	}

	tracef("running flag actions (cmd=%[1]q)", cmd.Name)

	if err := cmd.runFlagActions(ctx); err != nil {
		return err
	}

	var subCmd *Command

	if args.Present() {
		tracef("checking positional args %[1]q (cmd=%[2]q)", args, cmd.Name)

		name := args.First()

		tracef("using first positional argument as sub-command name=%[1]q (cmd=%[2]q)", name, cmd.Name)

		if cmd.SuggestCommandFunc != nil {
			name = cmd.SuggestCommandFunc(cmd.Commands, name)
		}
		subCmd = cmd.Command(name)
		if subCmd == nil {
			hasDefault := cmd.DefaultCommand != ""
			isFlagName := checkStringSliceIncludes(name, cmd.FlagNames())

			if hasDefault {
				tracef("using default command=%[1]q (cmd=%[2]q)", cmd.DefaultCommand, cmd.Name)
			}

			if isFlagName || hasDefault {
				argsWithDefault := cmd.argsWithDefaultCommand(args)
				tracef("using default command args=%[1]q (cmd=%[2]q)", argsWithDefault, cmd.Name)
				if !reflect.DeepEqual(args, argsWithDefault) {
					subCmd = cmd.Command(argsWithDefault.First())
				}
			}
		}
	} else if cmd.parent == nil && cmd.DefaultCommand != "" {
		tracef("no positional args present; checking default command %[1]q (cmd=%[2]q)", cmd.DefaultCommand, cmd.Name)

		if dc := cmd.Command(cmd.DefaultCommand); dc != cmd {
			subCmd = dc
		}
	}

	if subCmd != nil {
		tracef("running sub-command %[1]q with arguments %[2]q (cmd=%[3]q)", subCmd.Name, cmd.Args(), cmd.Name)
		return subCmd.Run(ctx, cmd.Args().Slice())
	}

	if cmd.Action == nil {
		cmd.Action = helpCommandAction
	} else {
		if err := cmd.checkAllRequiredFlags(); err != nil {
			cmd.isInError = true
			_ = ShowSubcommandHelp(cmd)
			return err
		}

		if len(cmd.Arguments) > 0 {
			rargs := cmd.Args().Slice()
			tracef("calling argparse with %[1]v", rargs)
			for _, arg := range cmd.Arguments {
				var err error
				rargs, err = arg.Parse(rargs)
				if err != nil {
					tracef("calling with %[1]v (cmd=%[2]q)", err, cmd.Name)
					return err
				}
			}
			cmd.parsedArgs = &stringSliceArgs{v: rargs}
		}
	}

	if err := cmd.Action(ctx, cmd); err != nil {
		tracef("calling handleExitCoder with %[1]v (cmd=%[2]q)", err, cmd.Name)
		deferErr = cmd.handleExitCoder(ctx, err)
	}

	tracef("returning deferErr (cmd=%[1]q) %[2]q", cmd.Name, deferErr)
	return deferErr
}

func (cmd *Command) checkHelp() bool {
	tracef("checking if help is wanted (cmd=%[1]q)", cmd.Name)

	if HelpFlag == nil {
		return false
	}

	for _, name := range HelpFlag.Names() {
		if cmd.Bool(name) {
			return true
		}
	}

	return false
}

func (cmd *Command) newFlagSet() (*flag.FlagSet, error) {
	allFlags := cmd.allFlags()

	cmd.appliedFlags = append(cmd.appliedFlags, allFlags...)

	tracef("making new flag set (cmd=%[1]q)", cmd.Name)

	return newFlagSet(cmd.Name, allFlags)
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

// useShortOptionHandling traverses Lineage() for *any* ancestors
// with UseShortOptionHandling
func (cmd *Command) useShortOptionHandling() bool {
	for _, pCmd := range cmd.Lineage() {
		if pCmd.UseShortOptionHandling {
			return true
		}
	}

	return false
}

func (cmd *Command) suggestFlagFromError(err error, commandName string) (string, error) {
	fl, parseErr := flagFromError(err)
	if parseErr != nil {
		return "", err
	}

	flags := cmd.Flags
	hideHelp := cmd.hideHelp()

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

func (cmd *Command) parseFlags(args Args) (Args, error) {
	tracef("parsing flags from arguments %[1]q (cmd=%[2]q)", args, cmd.Name)

	cmd.parsedArgs = nil
	if v, err := cmd.newFlagSet(); err != nil {
		return args, err
	} else {
		cmd.flagSet = v
	}

	if cmd.SkipFlagParsing {
		tracef("skipping flag parsing (cmd=%[1]q)", cmd.Name)

		return cmd.Args(), cmd.flagSet.Parse(append([]string{"--"}, args.Tail()...))
	}

	tracef("walking command lineage for persistent flags (cmd=%[1]q)", cmd.Name)

	for pCmd := cmd.parent; pCmd != nil; pCmd = pCmd.parent {
		tracef(
			"checking ancestor command=%[1]q for persistent flags (cmd=%[2]q)",
			pCmd.Name, cmd.Name,
		)

		for _, fl := range pCmd.Flags {
			flNames := fl.Names()

			pfl, ok := fl.(LocalFlag)
			if !ok || pfl.IsLocal() {
				tracef("skipping non-persistent flag %[1]q (cmd=%[2]q)", flNames, cmd.Name)
				continue
			}

			tracef(
				"checking for applying persistent flag=%[1]q pCmd=%[2]q (cmd=%[3]q)",
				flNames, pCmd.Name, cmd.Name,
			)

			applyPersistentFlag := true

			cmd.flagSet.VisitAll(func(f *flag.Flag) {
				for _, name := range flNames {
					if name == f.Name {
						applyPersistentFlag = false
						break
					}
				}
			})

			if !applyPersistentFlag {
				tracef("not applying as persistent flag=%[1]q (cmd=%[2]q)", flNames, cmd.Name)

				continue
			}

			tracef("applying as persistent flag=%[1]q (cmd=%[2]q)", flNames, cmd.Name)

			if err := fl.Apply(cmd.flagSet); err != nil {
				return cmd.Args(), err
			}

			tracef("appending to applied flags flag=%[1]q (cmd=%[2]q)", flNames, cmd.Name)
			cmd.appliedFlags = append(cmd.appliedFlags, fl)
		}
	}

	tracef("parsing flags iteratively tail=%[1]q (cmd=%[2]q)", args.Tail(), cmd.Name)
	defer tracef("done parsing flags (cmd=%[1]q)", cmd.Name)

	rargs := args.Tail()
	posArgs := []string{}
	for {
		tracef("rearrange:1 (cmd=%[1]q) %[2]q", cmd.Name, rargs)
		for {
			tracef("rearrange:2 (cmd=%[1]q) %[2]q %[3]q", cmd.Name, posArgs, rargs)

			// no more args to parse. Break out of inner loop
			if len(rargs) == 0 {
				break
			}

			if strings.TrimSpace(rargs[0]) == "" {
				break
			}

			// stop parsing once we see a "--"
			if rargs[0] == "--" {
				posArgs = append(posArgs, rargs...)
				cmd.parsedArgs = &stringSliceArgs{posArgs}
				return cmd.parsedArgs, nil
			}

			// let flagset parse this
			if rargs[0][0] == '-' {
				break
			}

			tracef("rearrange-3 (cmd=%[1]q) check %[2]q", cmd.Name, rargs[0])

			// if there is a command by that name let the command handle the
			// rest of the parsing
			if cmd.Command(rargs[0]) != nil {
				posArgs = append(posArgs, rargs...)
				cmd.parsedArgs = &stringSliceArgs{posArgs}
				return cmd.parsedArgs, nil
			}

			posArgs = append(posArgs, rargs[0])

			// if this is the sole argument then
			// break from inner loop
			if len(rargs) == 1 {
				rargs = []string{}
				break
			}

			rargs = rargs[1:]
		}
		if err := parseIter(cmd.flagSet, cmd, rargs, cmd.Root().shellCompletion); err != nil {
			posArgs = append(posArgs, cmd.flagSet.Args()...)
			tracef("returning-1 (cmd=%[1]q) args %[2]q", cmd.Name, posArgs)
			cmd.parsedArgs = &stringSliceArgs{posArgs}
			return cmd.parsedArgs, err
		}
		tracef("rearrange-4 (cmd=%[1]q) check %[2]q", cmd.Name, cmd.flagSet.Args())
		rargs = cmd.flagSet.Args()
		if len(rargs) == 0 || strings.TrimSpace(rargs[0]) == "" || rargs[0] == "-" {
			break
		}
	}

	posArgs = append(posArgs, cmd.flagSet.Args()...)
	tracef("returning-2 (cmd=%[1]q) args %[2]q", cmd.Name, posArgs)
	cmd.parsedArgs = &stringSliceArgs{posArgs}
	return cmd.parsedArgs, nil
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
		if command.Hidden || command.Name == helpName {
			continue
		}
		ret = append(ret, command)
	}
	return ret
}

// VisibleFlagCategories returns a slice containing all the visible flag categories with the flags they contain
func (cmd *Command) VisibleFlagCategories() []VisibleFlagCategory {
	if cmd.flagCategories == nil {
		cmd.flagCategories = newFlagCategoriesFromFlags(cmd.allFlags())
	}
	return cmd.flagCategories.VisibleCategories()
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (cmd *Command) VisibleFlags() []Flag {
	return visibleFlags(cmd.allFlags())
}

func (cmd *Command) appendFlag(fl Flag) {
	if !hasFlag(cmd.Flags, fl) {
		cmd.Flags = append(cmd.Flags, fl)
	}
}

// VisiblePersistentFlags returns a slice of [LocalFlag] with Persistent=true and Hidden=false.
func (cmd *Command) VisiblePersistentFlags() []Flag {
	var flags []Flag
	for _, fl := range cmd.Root().Flags {
		pfl, ok := fl.(LocalFlag)
		if !ok || pfl.IsLocal() {
			continue
		}
		flags = append(flags, fl)
	}
	return visibleFlags(flags)
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
		newArgs := &stringSliceArgs{v: rawArgs}

		return newArgs
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
					tracef("flag found for name %[1]q (cmd=%[2]q)", name, cmd.Name)
					return f
				}
			}
		}
	}

	tracef("flag NOT found for name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}

func (cmd *Command) lookupFlagSet(name string) *flag.FlagSet {
	for _, pCmd := range cmd.Lineage() {
		if pCmd.flagSet == nil {
			continue
		}

		if f := pCmd.flagSet.Lookup(name); f != nil {
			tracef("matching flag set found for name %[1]q (cmd=%[2]q)", name, cmd.Name)
			return pCmd.flagSet
		}
	}

	tracef("matching flag set NOT found for name %[1]q (cmd=%[2]q)", name, cmd.Name)
	cmd.onInvalidFlag(context.TODO(), name)
	return nil
}

func (cmd *Command) checkRequiredFlag(f Flag) (bool, string) {
	if rf, ok := f.(RequiredFlag); ok && rf.IsRequired() {
		flagPresent := false
		flagName := ""

		for _, key := range f.Names() {
			// use the first name to return since that is the
			// primary flag name
			if flagName == "" {
				flagName = key
			}

			if cmd.IsSet(strings.TrimSpace(key)) {
				flagPresent = true
				break
			}
		}

		if !flagPresent && flagName != "" {
			return false, flagName
		}
	}
	return true, ""
}

func (cmd *Command) checkAllRequiredFlags() requiredFlagsErr {
	for pCmd := cmd; pCmd != nil; pCmd = pCmd.parent {
		if err := pCmd.checkRequiredFlags(); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *Command) checkRequiredFlags() requiredFlagsErr {
	tracef("checking for required flags (cmd=%[1]q)", cmd.Name)

	missingFlags := []string{}

	for _, f := range cmd.appliedFlags {
		if ok, name := cmd.checkRequiredFlag(f); !ok {
			missingFlags = append(missingFlags, name)
		}
	}

	if len(missingFlags) != 0 {
		tracef("found missing required flags %[1]q (cmd=%[2]q)", missingFlags, cmd.Name)

		return &errRequiredFlags{missingFlags: missingFlags}
	}

	tracef("all required flags set (cmd=%[1]q)", cmd.Name)

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
	flSet := cmd.lookupFlagSet(name)

	if flSet == nil {
		return false
	}

	isSet := false

	flSet.Visit(func(f *flag.Flag) {
		if f.Name == name {
			isSet = true
		}
	})

	if isSet {
		tracef("flag with name %[1]q found via flag set lookup (cmd=%[2]q)", name, cmd.Name)
		return true
	}

	fl := cmd.lookupFlag(name)
	if fl == nil {
		tracef("flag with name %[1]q NOT found; assuming not set (cmd=%[2]q)", name, cmd.Name)
		return false
	}

	isSet = fl.IsSet()
	if isSet {
		tracef("flag with name %[1]q is set (cmd=%[2]q)", name, cmd.Name)
	} else {
		tracef("flag with name %[1]q is NOT set (cmd=%[2]q)", name, cmd.Name)
	}

	return isSet
}

// LocalFlagNames returns a slice of flag names used in this
// command.
func (cmd *Command) LocalFlagNames() []string {
	names := []string{}

	cmd.flagSet.Visit(makeFlagNameVisitor(&names))

	// Check the flags which have been set via env or file
	for _, f := range cmd.Flags {
		if f.IsSet() {
			names = append(names, f.Names()...)
		}
	}

	// Sort out the duplicates since flag could be set via multiple
	// paths
	m := map[string]struct{}{}
	uniqNames := []string{}

	for _, name := range names {
		if _, ok := m[name]; !ok {
			m[name] = struct{}{}
			uniqNames = append(uniqNames, name)
		}
	}

	return uniqNames
}

// FlagNames returns a slice of flag names used by the this command
// and all of its parent commands.
func (cmd *Command) FlagNames() []string {
	names := cmd.LocalFlagNames()

	if cmd.parent != nil {
		names = append(cmd.parent.FlagNames(), names...)
	}

	return names
}

// Lineage returns *this* command and all of its ancestor commands
// in order from child to parent
func (cmd *Command) Lineage() []*Command {
	lineage := []*Command{cmd}

	if cmd.parent != nil {
		lineage = append(lineage, cmd.parent.Lineage()...)
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
		tracef("value found for name %[1]q (cmd=%[2]q)", name, cmd.Name)
		return fs.Lookup(name).Value.(flag.Getter).Get()
	}

	tracef("value NOT found for name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}

// Args returns the command line arguments associated with the
// command.
func (cmd *Command) Args() Args {
	if cmd.parsedArgs != nil {
		return cmd.parsedArgs
	}
	return &stringSliceArgs{v: cmd.flagSet.Args()}
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

func (cmd *Command) runFlagActions(ctx context.Context) error {
	for _, fl := range cmd.appliedFlags {
		isSet := false

		// check only local flagset for running local flag actions
		for _, name := range fl.Names() {
			cmd.flagSet.Visit(func(f *flag.Flag) {
				if f.Name == name {
					isSet = true
				}
			})
			if isSet {
				break
			}
		}

		// If the flag hasnt been set on cmd line then we need to further
		// check if it has been set via other means. If however it has
		// been set by other means but it is persistent(and not set via current cmd)
		// do not run the flag action
		if !isSet {
			if !fl.IsSet() {
				continue
			}
			if pf, ok := fl.(LocalFlag); ok && !pf.IsLocal() {
				continue
			}
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
