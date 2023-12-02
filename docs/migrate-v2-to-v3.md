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

# EnvVar is now a value source

Change `EnvVars: "XXXXX"` to `Sources: EnvVars("XXXXX")` (plural).

* OLD:
```go
cli.StringFlag{
        EnvVars: []string{"APP_LANG"}
}
```

* NEW:
```go
cli.StringFlag{
        Sources: EnvVars("APP_LANG")
}
```

# Actions signatures have changed

A command's `Action:` now takes 2 arguments.

* OLD: `Action: func(ctx *cli.Context) error {`
* NEW: `Action: func(ctx context.Context, cmd *cli.Command) error {`

Compiler messages you might see:

```
cannot use func literal (type func(*cli.Context) error) as type cli.ActionFunc in field value
```
