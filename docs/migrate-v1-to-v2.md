Migration Guide: v1 to v2
===


v2 has a number of breaking changes but converting is relatively
straightforward: make the changes documented below then resolve any
compiler errors. We hope this will be sufficient for most typical
users.

If you find any issues not covered by this document, please post a
comment on [Issue 921](https://github.com/urfave/cli/issues/921) or
consider sending a PR to help improve this guide.

<!-- toc -->

  * [Flags before args](#flags-before-args)
  * [Import string changed](#import-string-changed)
  * [Flag aliases are done differently.](#flag-aliases-are-done-differently)
  * [EnvVar is now a list (EnvVars)](#envvar-is-now-a-list-envvars)
  * [Commands are now lists of pointers](#commands-are-now-lists-of-pointers)
  * [Lists of commands should be pointers](#lists-of-commands-should-be-pointers)
  * [cli.Flag changed](#cliflag-changed)
  * [Appending Commands](#appending-commands)
  * [Actions returns errors](#actions-returns-errors)
  * [Everything else](#everything-else)
  * [Full API Example](#full-api-example)

<!-- tocstop -->

# Flags before args

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

# Import string changed

* OLD: `import "github.com/urfave/cli"`
* NEW: `import "github.com/urfave/cli/v2"`

Check each file for this and make the change.

Shell command to find them all: `fgrep -rl github.com/urfave/cli *`

# Flag aliases are done differently.

Change `Name: "foo, f"` to `Name: "foo", Aliases: []string{"f"}`

* OLD:
```go
cli.StringFlag{
        Name: "config, cfg"
}
```

* NEW:
```go
cli.StringFlag{
        Name: "config",
        Aliases: []string{"cfg"},
}
```

Sadly v2 doesn't warn you if a comma is in the name.
(https://github.com/urfave/cli/issues/1103)

# EnvVar is now a list (EnvVars)

Change `EnvVar: "XXXXX"` to `EnvVars: []string{"XXXXX"}` (plural).

* OLD:
```go
cli.StringFlag{
        EnvVar: "APP_LANG"
}
```

* NEW:
```go
cli.StringFlag{
        EnvVars: []string{"APP_LANG"}
}
```

# Actions returns errors

A command's `Action:` now returns an `error`.

* OLD: `Action: func(c *cli.Context) {`
* NEW: `Action: func(c *cli.Context) error {`

Compiler messages you might see:

```
cannot use func literal (type func(*cli.Context)) as type cli.ActionFunc in field value
```

# cli.Flag changed

`cli.Flag` is now a list of pointers.

What this means to you:

If you make a list of flags, add a `&` in front of each
item.   cli.BoolFlag, cli.StringFlag, etc.

* OLD:
```go
        app.Flags = []cli.Flag{
               cli.BoolFlag{
```

* NEW:
```go
        app.Flags = []cli.Flag{
               &cli.BoolFlag{
```

Compiler messages you might see:

```
	cli.StringFlag does not implement cli.Flag (Apply method has pointer receiver)
```

# Commands are now lists of pointers

Occurrences of `[]Command` have been changed to `[]*Command`.

What this means to you:

Look for `[]cli.Command{}` and change it to `[]*cli.Command{}`

Example:

* OLD: `var commands = []cli.Command{}`
* NEW: `var commands = []*cli.Command{}`

Compiler messages you might see:

```
cannot convert commands (type []cli.Command) to type cli.CommandsByName
cannot use commands (type []cli.Command) as type []*cli.Command in assignment
```

# Lists of commands should be pointers

If you are building up a list of commands, the individual items should
now be pointers.

* OLD: `cli.Command{`
* NEW: `&cli.Command{`

Compiler messages you might see:

```
cannot use cli.Command literal (type cli.Command) as type *cli.Command in argument to
```

# Appending Commands

Appending to a list of commands needs to be changed since the list is
now pointers.

* OLD: `commands = append(commands, *c)`
* NEW: `commands = append(commands, c)`

Compiler messages you might see:

```
cannot use c (type *cli.Command) as type cli.Command in append
```

# Everything else

Compile the code and work through any errors. Most should
relate to issues listed above.

Once it compiles, test the command. Review the output of `-h` or any
help messages to verify they match the intended flags and subcommands.
Then test the program itself.

If you find any issues not covered by this document please let us know
by submitting a comment on
[Issue 921](https://github.com/urfave/cli/issues/921)
so that others can benefit.
