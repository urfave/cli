# Migration Guide: v1 to v2

v2 has a number of breaking changes but converting is relatively
straightforward: make the changes documented below then resolve any
compiler errors. We hope this will be sufficient for most typical
users.

If you find any issues not covered by this document, please post a
comment on [Issue 921](https://github.com/urfave/cli/issues/921) or
consider sending a PR to help improve this guide.

## Flags before args

In v2 flags must come before args. This is more POSIX-compliant.  You
may need to update scripts, user documentation, etc.

This will work:

```
cli hello --shout rick
```

This will not:

```
cli hello rick --shout
```

## Import string changed

=== "v1"

    `import "github.com/urfave/cli"`

=== "v2"

    `import "github.com/urfave/cli/v2"`

Check each file for this and make the change.

Shell command to find them all: `fgrep -rl github.com/urfave/cli *`

## Flag aliases are done differently

Change `Name: "foo, f"` to `Name: "foo", Aliases: []string{"f"}`

=== "v1"

    ```go
    cli.StringFlag{
            Name: "config, cfg"
    }
    ```

=== "v2"
    
    ```go
    cli.StringFlag{
            Name: "config",
            Aliases: []string{"cfg"},
    }
    ```

Sadly v2 doesn't warn you if a comma is in the name.
(https://github.com/urfave/cli/issues/1103)

## EnvVar is now a list (EnvVars)

Change `EnvVar: "XXXXX"` to `EnvVars: []string{"XXXXX"}` (plural).

=== "v1"

    ```go
    cli.StringFlag{
            EnvVar: "APP_LANG"
    }
    ```

=== "v2"

    ```go
    cli.StringFlag{
            EnvVars: []string{"APP_LANG"}
    }
    ```

## Actions returns errors

A command's `Action:` now returns an `error`.

=== "v1"

    `Action: func(c *cli.Context) {`

=== "v2"

    `Action: func(c *cli.Context) error {`

Compiler messages you might see:

```
cannot use func literal (type func(*cli.Context)) as type cli.ActionFunc in field value
```

## cli.Flag changed

`cli.Flag` is now a list of pointers.

What this means to you:

If you make a list of flags, add a `&` in front of each
item.   cli.BoolFlag, cli.StringFlag, etc.

=== "v1"

    ```go
            app.Flags = []cli.Flag{
                   cli.BoolFlag{
    ```

=== "v2"
    
    ```go
            app.Flags = []cli.Flag{
                   &cli.BoolFlag{
    ```

Compiler messages you might see:

```
	cli.StringFlag does not implement cli.Flag (Apply method has pointer receiver)
```

## Commands are now lists of pointers

Occurrences of `[]Command` have been changed to `[]*Command`.

What this means to you:

Look for `[]cli.Command{}` and change it to `[]*cli.Command{}`

Example:

=== "v1"

    `var commands = []cli.Command{}`

=== "v2"

    `var commands = []*cli.Command{}`

Compiler messages you might see:

```
cannot convert commands (type []cli.Command) to type cli.CommandsByName
cannot use commands (type []cli.Command) as type []*cli.Command in assignment
```

## Lists of commands should be pointers

If you are building up a list of commands, the individual items should
now be pointers.

=== "v1"

    `cli.Command{`

=== "v2"

    `&cli.Command{`

Compiler messages you might see:

```
cannot use cli.Command literal (type cli.Command) as type *cli.Command in argument to
```

## Appending Commands

Appending to a list of commands needs to be changed since the list is
now pointers.

=== "v1"

    `commands = append(commands, *c)`

=== "v2"

    `commands = append(commands, c)`

Compiler messages you might see:

```
cannot use c (type *cli.Command) as type cli.Command in append
```

## GlobalString, GlobalBool and its likes are deprecated

Use simply `String` instead of `GlobalString`, `Bool` instead of `GlobalBool` 

## BoolTFlag and BoolT are deprecated

BoolTFlag was a Bool Flag with its default value set to true and BoolT was used to find any BoolTFlag used locally, so both are deprecated.

=== "v1"

    ```go
    cli.BoolTFlag{
            Name:   FlagName,
            Usage:  FlagUsage,
            EnvVar: "FLAG_ENV_VAR",
    }
    ```

=== "v2"
    
    ```go
    cli.BoolFlag{
            Name:   FlagName,
            Value:  true,
            Usage:  FlagUsage,
            EnvVar: "FLAG_ENV_VAR",
    }
    ```

## &cli.StringSlice{""} replaced with cli.NewStringSlice("")

Example: 

=== "v1"

    ```go
    Value: &cli.StringSlice{""},
    ```

=== "v2"
    
    ```go
    Value: cli.NewStringSlice(""),
    ```

## Replace deprecated functions

`cli.NewExitError()` is deprecated.  Use `cli.Exit()` instead.  ([Staticcheck](https://staticcheck.io/) detects this automatically and recommends replacement code.)

## Everything else

Compile the code and work through any errors. Most should
relate to issues listed above.

Once it compiles, test the command. Review the output of `-h` or any
help messages to verify they match the intended flags and subcommands.
Then test the program itself.

If you find any issues not covered by this document please let us know
by submitting a comment on
[Issue 921](https://github.com/urfave/cli/issues/921)
so that others can benefit.
