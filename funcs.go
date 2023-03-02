package cli

// ShellCompleteFunc is an action to execute when the shell completion flag is set
type ShellCompleteFunc func(*Context)

// BeforeFunc is an action that executes prior to any subcommands being run once
// the context is ready.  If a non-nil error is returned, no subcommands are
// run.
type BeforeFunc func(*Context) error

// AfterFunc is an action that executes after any subcommands are run and have
// finished. The AfterFunc is run even if Action() panics.
type AfterFunc func(*Context) error

// ActionFunc is the action to execute when no subcommands are specified
type ActionFunc func(*Context) error

// CommandNotFoundFunc is executed if the proper command cannot be found
type CommandNotFoundFunc func(*Context, string)

// OnUsageErrorFunc is executed if a usage error occurs. This is useful for displaying
// customized usage error messages.  This function is able to replace the
// original error messages.  If this function is not set, the "Incorrect usage"
// is displayed and the execution is interrupted.
type OnUsageErrorFunc func(cCtx *Context, err error, isSubcommand bool) error

// InvalidFlagAccessFunc is executed when an invalid flag is accessed from the context.
type InvalidFlagAccessFunc func(*Context, string)

// ExitErrHandlerFunc is executed if provided in order to handle exitError values
// returned by Actions and Before/After functions.
type ExitErrHandlerFunc func(cCtx *Context, err error)

// FlagStringFunc is used by the help generation to display a flag, which is
// expected to be a single line.
type FlagStringFunc func(Flag) string

// FlagNamePrefixFunc is used by the default FlagStringFunc to create prefix
// text for a flag's full name.
type FlagNamePrefixFunc func(fullName []string, placeholder string) string

// FlagEnvHintFunc is used by the default FlagStringFunc to annotate flag help
// with the environment variable details.
type FlagEnvHintFunc func(envVars []string, str string) string

// FlagFileHintFunc is used by the default FlagStringFunc to annotate flag help
// with the file path details.
type FlagFileHintFunc func(filePath, str string) string
