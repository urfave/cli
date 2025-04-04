# Migration Guide: v2 to v3

v3 has a number of breaking changes but converting is relatively
straightforward: make the changes documented below then resolve any
compiler errors. We hope this will be sufficient for most typical
users.

If you find any issues not covered by this document, please post a
comment on [the discussion](https://github.com/urfave/cli/discussions/2084) or
consider sending a PR to help improve this guide.

## Import string changed

=== "v2"

    `import "github.com/urfave/cli/v2"`

=== "v3"

    `import "github.com/urfave/cli/v3"`

Check each file for this and make the change.

Shell command to find them all: `fgrep -rl github.com/urfave/cli/v2 *`

## FilePath

Change `FilePath: "XXXXX"` to `Sources: Files("XXXXX")`.

=== "v2"

    ```go
    cli.StringFlag{
            FilePath: "/path/to/foo",
    }
    ```

=== "v3"

    ```go
    cli.StringFlag{
            Sources: Files("/path/to/foo"),
    }
    ```

## EnvVars

Change `EnvVars: "XXXXX"` to `Sources: EnvVars("XXXXX")`.

=== "v2"

    ```go
    cli.StringFlag{
            EnvVars: []string{"APP_LANG"},
    }
    ```

=== "v3"

    ```go
    cli.StringFlag{
            Sources: EnvVars("APP_LANG"),
    }
    ```

## Altsrc has been moved out of the cli library into its own repo

=== "v2"

    `import "github.com/urfave/cli/v2/altsrc"`

=== "v3"

    `import "github.com/urfave/cli-altsrc/v3"`

## Altsrc is now a value source for cli

=== "v2"
    
    ```go
    altsrc.StringFlag{
        &cli.String{....}
    }
    ```

=== "v3"
    
    ```go
    cli.StringFlag{
        Sources: altsrc.JSON("key", "/tmp/foo")
    }
    ```

## Order of precedence of envvars, filepaths, altsrc now depends on the order in which they are defined


=== "v2"

    ```go
    cli.StringFlag{
            EnvVars: []string{"APP_LANG"},
    }
    cli.StringFlag{
            FilePath: "/path/to/foo",
    }
    ```

=== "v3"

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

In the above case the Envs are checked first and if not found then files are looked at and then finally the `altsrc`

## cli.Context has been removed

All functions handled previously by cli.Context have been incorporated into `cli.Command`:

* Change `cli.Context.IsSet` -> `cli.Command.IsSet`
* Change `cli.Context.NumFlags` -> `cli.Command.NumFlags`
* Change `cli.Context.FlagNames` -> `cli.Command.FlagNames`
* Change `cli.Context.LocalFlagNames` -> `cli.Command.LocalFlagNames`
* Change `cli.Context.Lineage` -> `cli.Command.Lineage`
* Change `cli.Context.Count` -> `cli.Command.Count`
* Change `cli.Context.Value` -> `cli.Command.Value`
* Change `cli.Context.Args` -> `cli.Command.Args`
* Change `cli.Context.NArg` -> `cli.Command.NArg`

## Handler func signatures have changed

All handler functions now take at least 2 arguments a `context.Context` and a pointer to `Cli.Command`
in addition to other specific args. This allows handler functions to utilize `context.Context` for
blocking/time-specific operations and so on.

### BeforeFunc

=== "v2"

    `type BeforeFunc func(*Context) error`

=== "v3"

    `type BeforeFunc func(context.Context, *cli.Command) (context.Context, error)`

### AfterFunc

=== "v2"

    `type AfterFunc func(*Context) error`

=== "v3"

    `type AfterFunc func(context.Context, *cli.Command) error`

### ActionFunc

=== "v2"

    `type ActionFunc func(*Context) error`

=== "v3"

    `type ActionFunc func(context.Context, *cli.Command) error`

### CommandNotFoundFunc

=== "v2"

    `type CommandNotFoundFunc func(*Context, string) error`

=== "v3"

    `type CommandNotFoundFunc func(context.Context, *cli.Command, string) error`

### OnUsageErrorFunc

=== "v2"

    `type OnUsageErrorFunc func(*Context, err error, isSubcommand bool) error`

=== "v3"

    `type OnUsageErrorFunc func(context.Context, *cli.Command, err error, isSubcommand bool) error`

### InvalidAccessFunc

=== "v2"

    `type InvalidAccessFunc func(*Context, string) error`

=== "v3"

    `type InvalidAccessFunc func(context.Context, *cli.Command, string) error`

### ExitErrHandlerFunc

=== "v2"

    `type ExitErrHandlerFunc func(*Context, err error) error`

=== "v3"

    `type ExitErrHandlerFunc func(context.Context, *cli.Command, err error) error`

Compiler messages you might see(for ActionFunc):

```
cannot use func literal (type func(*cli.Context) error) as type cli.ActionFunc in field value
```

Similar messages would be shown for other funcs.

## TimestampFlag

=== "v2"

    ```go
    &cli.TimestampFlag{
        Name:   "foo",
        Layout: time.RFC3339,
    }
    ```

=== "v3"

    ```go
    &cli.TimestampFlag{
        Name:  "foo",
        Config: cli.TimestampConfig{
            Layouts: []string{time.RFC3339},
        },
    }
    ```
