package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var (
	// DefaultSuccessExitCode is the default for use with os.Exit intended to
	// indicate success
	DefaultSuccessExitCode = 0
	// DefaultErrorExitCode is the default for use with os.Exit intended to
	// indicate an error
	DefaultErrorExitCode = 1
)

// App is the main structure of a cli application. It is recommended that
// an app be created with the cli.NewApp() function
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
	// List of commands to execute
	Commands []Command
	// List of flags to parse
	Flags []Flag
	// Boolean to enable bash completion commands
	EnableBashCompletion bool
	// Boolean to hide built-in help command
	HideHelp bool
	// Boolean to hide built-in version flag and the VERSION section of help
	HideVersion bool
	// Populate on app startup, only gettable throught method Categories()
	categories CommandCategories
	// An action to execute when the bash-completion flag is set
	BashComplete BashCompleteFn
	// An action to execute before any subcommands are run, but after the context is ready
	// If a non-nil error is returned, no subcommands are run
	Before BeforeFn
	// An action to execute after any subcommands are run, but after the subcommand has finished
	// It is run even if Action() panics
	After AfterFn
	// The action to execute when no subcommands are specified
	Action ActionFn
	// Execute this function if the proper command cannot be found
	CommandNotFound CommandNotFoundFn
	// Execute this function, if an usage error occurs. This is useful for displaying customized usage error messages.
	// This function is able to replace the original error messages.
	// If this function is not set, the "Incorrect usage" is displayed and the execution is interrupted.
	OnUsageError func(context *Context, err error, isSubcommand bool) error
	// Compilation date
	Compiled time.Time
	// List of all authors who contributed
	Authors []Author
	// Copyright of the binary if any
	Copyright string
	// Name of Author (Note: Use App.Authors, this is deprecated)
	Author string
	// Email of Author (Note: Use App.Authors, this is deprecated)
	Email string
	// Writer writer to write output to
	Writer io.Writer
}

// Tries to find out when this binary was compiled.
// Returns the current time if it fails to find it.
func compileTime() time.Time {
	info, err := os.Stat(os.Args[0])
	if err != nil {
		return time.Now()
	}
	return info.ModTime()
}

// Creates a new cli Application with some reasonable defaults for Name, Usage, Version and Action.
func NewApp() *App {
	return &App{
		Name:         filepath.Base(os.Args[0]),
		HelpName:     filepath.Base(os.Args[0]),
		Usage:        "A new cli application",
		UsageText:    "",
		Version:      "0.0.0",
		BashComplete: DefaultAppComplete,
		Action:       helpCommand.Action,
		Compiled:     compileTime(),
		Writer:       os.Stdout,
	}
}

// Entry point to the cli app. Parses the arguments slice and routes to the proper flag/args combination
func (a *App) Run(arguments []string) (ec int, err error) {
	if a.Author != "" || a.Email != "" {
		a.Authors = append(a.Authors, Author{Name: a.Author, Email: a.Email})
	}

	newCmds := []Command{}
	for _, c := range a.Commands {
		if c.HelpName == "" {
			c.HelpName = fmt.Sprintf("%s %s", a.HelpName, c.Name)
		}
		newCmds = append(newCmds, c)
	}
	a.Commands = newCmds

	a.categories = CommandCategories{}
	for _, command := range a.Commands {
		a.categories = a.categories.AddCommand(command.Category, command)
	}
	sort.Sort(a.categories)

	// append help to commands
	if a.Command(helpCommand.Name) == nil && !a.HideHelp {
		a.Commands = append(a.Commands, helpCommand)
		if (HelpFlag != BoolFlag{}) {
			a.appendFlag(HelpFlag)
		}
	}

	//append version/help flags
	if a.EnableBashCompletion {
		a.appendFlag(BashCompletionFlag)
	}

	if !a.HideVersion {
		a.appendFlag(VersionFlag)
	}

	// parse flags
	set := flagSet(a.Name, a.Flags)
	set.SetOutput(ioutil.Discard)
	err = set.Parse(arguments[1:])
	nerr := normalizeFlags(a.Flags, set)
	context := NewContext(a, set, nil)
	if nerr != nil {
		fmt.Fprintln(a.Writer, nerr)
		ShowAppHelp(context)
		return DefaultErrorExitCode, nerr
	}

	if checkCompletions(context) {
		return DefaultSuccessExitCode, nil
	}

	if err != nil {
		if a.OnUsageError != nil {
			err := a.OnUsageError(context, err, false)
			if err != nil {
				return DefaultErrorExitCode, err
			}
			return DefaultSuccessExitCode, err
		} else {
			fmt.Fprintf(a.Writer, "%s\n\n", "Incorrect Usage.")
			ShowAppHelp(context)
			return DefaultErrorExitCode, err
		}
	}

	if !a.HideHelp && checkHelp(context) {
		ShowAppHelp(context)
		return DefaultSuccessExitCode, nil
	}

	if !a.HideVersion && checkVersion(context) {
		ShowVersion(context)
		return DefaultSuccessExitCode, nil
	}

	if a.After != nil {
		defer func() {
			afterEc, afterErr := a.After(context)
			if afterErr != nil {
				if err != nil {
					err = NewMultiError(err, afterErr)
				} else {
					err = afterErr
				}
			}
			ec = afterEc
		}()
	}

	if a.Before != nil {
		ec, err = a.Before(context)
		if err != nil {
			fmt.Fprintf(a.Writer, "%v\n\n", err)
			ShowAppHelp(context)
			return ec, err
		}
	}

	args := context.Args()
	if args.Present() {
		name := args.First()
		c := a.Command(name)
		if c != nil {
			return c.Run(context)
		}
	}

	// Run default Action
	return a.Action(context), nil
}

// Another entry point to the cli app, takes care of passing arguments and error handling
func (a *App) RunAndExitOnError() {
	if exitCode, err := a.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode)
	}
}

// Invokes the subcommand given the context, parses ctx.Args() to generate command-specific flags
func (a *App) RunAsSubcommand(ctx *Context) (ec int, err error) {
	// append help to commands
	if len(a.Commands) > 0 {
		if a.Command(helpCommand.Name) == nil && !a.HideHelp {
			a.Commands = append(a.Commands, helpCommand)
			if (HelpFlag != BoolFlag{}) {
				a.appendFlag(HelpFlag)
			}
		}
	}

	newCmds := []Command{}
	for _, c := range a.Commands {
		if c.HelpName == "" {
			c.HelpName = fmt.Sprintf("%s %s", a.HelpName, c.Name)
		}
		newCmds = append(newCmds, c)
	}
	a.Commands = newCmds

	// append flags
	if a.EnableBashCompletion {
		a.appendFlag(BashCompletionFlag)
	}

	// parse flags
	set := flagSet(a.Name, a.Flags)
	set.SetOutput(ioutil.Discard)
	err = set.Parse(ctx.Args().Tail())
	nerr := normalizeFlags(a.Flags, set)
	context := NewContext(a, set, ctx)

	if nerr != nil {
		fmt.Fprintln(a.Writer, nerr)
		fmt.Fprintln(a.Writer)
		if len(a.Commands) > 0 {
			ShowSubcommandHelp(context)
		} else {
			ShowCommandHelp(ctx, context.Args().First())
		}
		return DefaultErrorExitCode, nerr
	}

	if checkCompletions(context) {
		return DefaultSuccessExitCode, nil
	}

	if err != nil {
		if a.OnUsageError != nil {
			err = a.OnUsageError(context, err, true)
			if err != nil {
				return DefaultErrorExitCode, err
			}
			return DefaultSuccessExitCode, err
		} else {
			fmt.Fprintf(a.Writer, "%s\n\n", "Incorrect Usage.")
			ShowSubcommandHelp(context)
			return DefaultErrorExitCode, err
		}
	}

	if len(a.Commands) > 0 {
		if checkSubcommandHelp(context) {
			return DefaultSuccessExitCode, nil
		}
	} else {
		if checkCommandHelp(ctx, context.Args().First()) {
			return DefaultSuccessExitCode, nil
		}
	}

	if a.After != nil {
		defer func() {
			afterEc, afterErr := a.After(context)
			if afterErr != nil {
				if err != nil {
					err = NewMultiError(err, afterErr)
				} else {
					err = afterErr
				}
			}
			ec = afterEc
		}()
	}

	if a.Before != nil {
		ec, err = a.Before(context)
		if err != nil {
			return ec, err
		}
	}

	args := context.Args()
	if args.Present() {
		name := args.First()
		c := a.Command(name)
		if c != nil {
			return c.Run(context)
		}
	}

	// Run default Action
	return a.Action(context), nil
}

// Returns the named command on App. Returns nil if the command does not exist
func (a *App) Command(name string) *Command {
	for _, c := range a.Commands {
		if c.HasName(name) {
			return &c
		}
	}

	return nil
}

// Returnes the array containing all the categories with the commands they contain
func (a *App) Categories() CommandCategories {
	return a.categories
}

func (a *App) hasFlag(flag Flag) bool {
	for _, f := range a.Flags {
		if flag == f {
			return true
		}
	}

	return false
}

func (a *App) appendFlag(flag Flag) {
	if !a.hasFlag(flag) {
		a.Flags = append(a.Flags, flag)
	}
}

// Author represents someone who has contributed to a cli project.
type Author struct {
	Name  string // The Authors name
	Email string // The Authors email
}

// String makes Author comply to the Stringer interface, to allow an easy print in the templating process
func (a Author) String() string {
	e := ""
	if a.Email != "" {
		e = "<" + a.Email + "> "
	}

	return fmt.Sprintf("%v %v", a.Name, e)
}
