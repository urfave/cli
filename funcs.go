package cli

// An action to execute when the bash-completion flag is set
type BashCompleteFn func(*Context)

// An action to execute before any subcommands are run, but after the context is ready
// If a non-nil error is returned, no subcommands are run
type BeforeFn func(*Context) (int, error)

// An action to execute after any subcommands are run, but after the subcommand has finished
// It is run even if Action() panics
type AfterFn func(*Context) (int, error)

// The action to execute when no subcommands are specified
type ActionFn func(*Context) int

// Execute this function if the proper command cannot be found
type CommandNotFoundFn func(*Context, string)
