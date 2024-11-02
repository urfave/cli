# Migration Guide: v2 to v3

v3 has a number of breaking changes but converting is relatively
straightforward: make the changes documented below then resolve any
compiler errors. We hope this will be sufficient for most typical
users.

If you find any issues not covered by this document, please post a
comment on [Issue 921](https://github.com/urfave/cli/issues/921) or
consider sending a PR to help improve this guide.

# Import string changed

* OLD: `import "github.com/urfave/cli/v2"`
* NEW: `import "github.com/urfave/cli/v3"`

Check each file for this and make the change.

Shell command to find them all: `fgrep -rl github.com/urfave/cli/v2 *`

# FilePath

Change `FilePath: "XXXXX"` to `Sources: Files("XXXXX")`.

* OLD:
```go
cli.StringFlag{
        FilePath: "/path/to/foo",
}
```

* NEW:
```go
cli.StringFlag{
        Sources: Files("/path/to/foo"),
}
```

# EnvVars

Change `EnvVars: "XXXXX"` to `Sources: EnvVars("XXXXX")`.

* OLD:
```go
cli.StringFlag{
        EnvVars: []string{"APP_LANG"},
}
```

* NEW:
```go
cli.StringFlag{
        Sources: EnvVars("APP_LANG"),
}
```

# Altsrc has been moved out of the cli library into its own repo

* OLD: `import "github.com/urfave/cli/v2/altsrc"`
* NEW: `import "github.com/urfave/cli-altsrc/v3"`

# Altsrc is now a value source for cli

* OLD:
```go
altsrc.StringFlag{
    &cli.String{....}
}
```

* NEW:
```go
cli.StringFlag{
    Sources: altsrc.JSON("key", "/tmp/foo")
}
```

# Order of precedence of envvars, filepaths, altsrc now depends on the order in which they are defined

* OLD:
```go
cli.StringFlag{
        EnvVars: []string{"APP_LANG"},
}
cli.StringFlag{
        FilePath: "/path/to/foo",
}
```

* NEW:
```go
cli.StringFlag{
        Sources: cli.ValueSourceChain{
           Chain: {
                EnvVars("APP_LANG"),
                Files("/path/to/foo"),
                altsrc.JSON("foo", "/path/to/"),
           }                
        },
}
```

In the above case the Envs are checked first and if not found then files are looked at and then finally the altsrc

# cli.Context has been removed

All functions handled previously by cli.Context have been incorporated into cli.Command

* Change `cli.Context.IsSet` -> `cli.Command.IsSet`
* Change `cli.Context.NumFlags` -> `cli.Command.NumFlags`
* Change `cli.Context.FlagNames` -> `cli.Command.FlagNames`
* Change `cli.Context.LocalFlagNames` -> `cli.Command.LocalFlagNames`
* Change `cli.Context.Lineage` -> `cli.Command.Lineage`
* Change `cli.Context.Count` -> `cli.Command.Count`
* Change `cli.Context.Value` -> `cli.Command.Value`
* Change `cli.Context.Args` -> `cli.Command.Args`
* Change `cli.Context.NArg` -> `cli.Command.NArg`

# Handler func signatures have changed

All handler functions now take atleast 2 arguments a context.Context and a pointer to Cli.Command
in addition to other specific args. This allows handler functions to utilize context.Context for
blocking/time-specific operations and so on

* OLD: `type BeforeFunc func(*Context) error`
* NEW: `type BeforeFunc func(context.Context, *cli.Command) (context.Context, error)`

* OLD: `type AfterFunc func(*Context) error`
* NEW: `type AfterFunc func(context.Context, *cli.Command) error`

* OLD: `type ActionFunc func(*Context) error`
* NEW: `type ActionFunc func(context.Context, *cli.Command) error`

* OLD: `type CommandNotFoundFunc func(*Context, string) error`
* NEW: `type CommandNotFoundFunc func(context.Context, *cli.Command, string) error`

* OLD: `type OnUsageErrorFunc func(*Context, err error, isSubcommand bool) error`
* NEW: `type OnUsageErrorFunc func(context.Context, *cli.Command, err error, isSubcommand bool) error`

* OLD: `type InvalidAccessFunc func(*Context, string) error`
* NEW: `type InvalidAccessFunc func(context.Context, *cli.Command, string) error`

* OLD: `type ExitErrHandlerFunc func(*Context, err error) error`
* NEW: `type ExitErrHandlerFunc func(context.Context, *cli.Command, err error) error`

Compiler messages you might see(for ActionFunc):

```
cannot use func literal (type func(*cli.Context) error) as type cli.ActionFunc in field value
```
Similar messages would be shown for other funcs
