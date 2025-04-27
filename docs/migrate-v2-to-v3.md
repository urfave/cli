# Migration Guide: v2 to v3

v3 has a number of breaking changes but converting is relatively
straightforward: make the changes documented below then resolve any
compiler errors. We hope this will be sufficient for most typical
users.

If you find any issues not covered by this document, please post a
comment on [the discussion](https://github.com/urfave/cli/discussions/2084) or
consider sending a PR to help improve this guide.

## New Import

=== "v2"

    `import "github.com/urfave/cli/v2"`

=== "v3"

    `import "github.com/urfave/cli/v3"`

Check each file for this and make the change.

Shell command to find them all: `fgrep -rl github.com/urfave/cli/v2 *`

## New Names

### cli.App

=== "v2"

    ```go
    cli.App{
            // ...
    }
    ```

=== "v3"

    ```go
    cli.Command{
            // ...
    }
    ```

### cli.App.EnableBashCompletion

=== "v2"

    ```go
    cli.App{
            EnableBashCompletion: true,
    }
    ```

=== "v3"

    ```go
    cli.Command{
            EnableShellCompletion: true,
    }
    ```

### cli.App.CustomAppHelpTemplate

=== "v2"

    ```go
    cli.App{
            CustomAppHelpTemplate: "...",
    }
    ```

=== "v3"

    ```go
    cli.Command{
            CustomRootCommandHelpTemplate: "...",
    }
    ```

### cli.App.RunContext

=== "v2"

    ```go
    (&cli.App{}).RunContext(context.Background(), os.Args)
    ```

=== "v3"

    ```go
    (&cli.Command{}).Run(context.Background(), os.Args)
    ```

### cli.App.BashComplete

=== "v2"

    ```go
    cli.App{
            BashComplete: func(ctx *cli.Context) {},
    }
    ```

=== "v3"

    ```go
    cli.Command{
            ShellComplete: func(ctx context.Context, cmd *cli.Command) {},
    }
    ```

### cli.Command.Subcommands

=== "v2"

    ```go
    cli.Command{
            Subcommands: []*cli.Command{},
    }
    ```

=== "v3"

    ```go
    cli.Command{
            Commands: []*cli.Command{},
    }
    ```

## Sources

### FilePath

=== "v2"

    ```go
    cli.StringFlag{
            FilePath: "/path/to/foo",
    }
    ```

=== "v3"

    ```go
    cli.StringFlag{
            Sources: cli.Files("/path/to/foo"),
    }
    ```

    or 

    ```go
    cli.StringFlag{
        Sources: cli.NewValueSourceChain(
            cli.File("/path/to/foo"),
        ),
    }
    ```

### EnvVars

=== "v2"

    ```go
    cli.StringFlag{
            EnvVars: []string{"APP_LANG"},
    }
    ```

=== "v3"

    ```go
    cli.StringFlag{
            Sources: cli.EnvVars("APP_LANG"),
    }
    ```

    or 

    ```go
    cli.StringFlag{
        Sources: cli.NewValueSourceChain(
           cli.EnvVar("APP_LANG"),
        ),
    }
    ```

### Altsrc

#### Altsrc is now a dedicated module

=== "v2"

    `import "github.com/urfave/cli/v2/altsrc"`

=== "v3"

    `import altsrc "github.com/urfave/cli-altsrc/v3"`

#### Altsrc is now a value source for CLI

=== "v2"
    
    ```go
    altsrc.NewStringFlag(
        &cli.StringFlag{
            Name:        "key",
            Value:       "/tmp/foo",
        },
    ),
    ```

=== "v3"
    
    Requires to use at least `github.com/urfave/cli-altsrc/v3@v3.0.0-alpha2.0.20250227140532-11fbec4d81a7`

    ```go
    cli.StringFlag{
        Sources: cli.NewValueSourceChain(altsrcjson.JSON("key", altsrc.StringSourcer("/path/to/foo.json"))),
    }
    ```

### Order of precedence of envvars, filepaths, altsrc now depends on the order in which they are defined

=== "v2"

    ```go
    altsrc.NewStringFlag(
        &cli.StringFlag{
            Name:     "key",
            EnvVars:  []string{"APP_LANG"},
            FilePath: "/path/to/foo",
        },
    ),
    ```

=== "v3"

    Requires to use at least `github.com/urfave/cli-altsrc/v3@v3.0.0-alpha2.0.20250227140532-11fbec4d81a7` 

    ```go
    import altsrcjson "github.com/urfave/cli-altsrc/v3/json"
    
    // ...

    &cli.StringFlag{
        Name: "key",
        Sources: cli.NewValueSourceChain(
            cli.EnvVar("APP_LANG"),
            cli.File("/path/to/foo"),
            altsrcjson.JSON("key", altsrc.StringSourcer("/path/to/foo.json")),
        ),
    },
    ```

In the above case the Envs are checked first and if not found then files are looked at and then finally the `altsrc`

## cli.Context has been removed

All functions handled previously by `cli.Context` have been incorporated into `cli.Command`:

| v2                           | v3                           |
|------------------------------|------------------------------|
| `cli.Context.IsSet`          | `cli.Command.IsSet`          |
| `cli.Context.NumFlags`       | `cli.Command.NumFlags`       |
| `cli.Context.FlagNames`      | `cli.Command.FlagNames`      |
| `cli.Context.LocalFlagNames` | `cli.Command.LocalFlagNames` |
| `cli.Context.Lineage`        | `cli.Command.Lineage`        |
| `cli.Context.Count`          | `cli.Command.Count`          |
| `cli.Context.Value`          | `cli.Command.Value`          |
| `cli.Context.Args`           | `cli.Command.Args`           |
| `cli.Context.NArg`           | `cli.Command.NArg`           |

## Handler Function Signatures Changes

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

## Authors

=== "v2"

    ```go
    &cli.App{
        Authors: []*cli.Author{
            {Name: "Some Guy", Email: "someguy@example.com"},
        },
    }
    ```

=== "v3"

    ```go
    // import "net/mail"
    &cli.Command{
        Authors: []any{
            mail.Address{Name: "Some Guy", Address: "someguy@example.com"},
        },
    }
    ```
