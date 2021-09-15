cli v2 manual
===

<!-- toc -->

- [Migrating From Older Releases](#migrating-from-older-releases)
- [Getting Started](#getting-started)
- [Examples](#examples)
  * [Arguments](#arguments)
  * [Flags](#flags)
    + [Placeholder Values](#placeholder-values)
    + [Alternate Names](#alternate-names)
    + [Ordering](#ordering)
    + [Values from the Environment](#values-from-the-environment)
    + [Values from files](#values-from-files)
    + [Values from alternate input sources (YAML, TOML, and others)](#values-from-alternate-input-sources-yaml-toml-and-others)
    + [Required Flags](#required-flags)
    + [Default Values for help output](#default-values-for-help-output)
    + [Precedence](#precedence)
  * [Subcommands](#subcommands)
  * [Subcommands categories](#subcommands-categories)
  * [Exit code](#exit-code)
  * [Combining short options](#combining-short-options)
  * [Bash Completion](#bash-completion)
    + [Default auto-completion](#default-auto-completion)
    + [Custom auto-completion](#custom-auto-completion)
    + [Enabling](#enabling)
    + [Distribution and Persistent Autocompletion](#distribution-and-persistent-autocompletion)
    + [Customization](#customization)
    + [ZSH Support](#zsh-support)
    + [ZSH default auto-complete example](#zsh-default-auto-complete-example)
    + [ZSH custom auto-complete example](#zsh-custom-auto-complete-example)
    + [PowerShell Support](#powershell-support)
  * [Generated Help Text](#generated-help-text)
    + [Customization](#customization-1)
  * [Version Flag](#version-flag)
    + [Customization](#customization-2)
  * [Timestamp Flag](#timestamp-flag)
  * [Full API Example](#full-api-example)

<!-- tocstop -->

## Migrating From Older Releases

There are a small set of breaking changes between v1 and v2.
Converting is relatively straightforward and typically takes less than
an hour. Specific steps are included in
[Migration Guide: v1 to v2](../migrate-v1-to-v2.md). Also see the [pkg.go.dev docs](https://pkg.go.dev/github.com/urfave/cli/v2) for v2 API documentation.

## Getting Started

One of the philosophies behind cli is that an API should be playful and full of
discovery. So a cli app can be as little as one line of code in `main()`.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "A new cli application"
} -->
``` go
package main

import (
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  (&cli.App{}).Run(os.Args)
}
```

This app will run and show help text, but is not very useful. Let's give an
action to execute and some help documentation:

<!-- {
  "output": "boom! I say!"
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Name: "boom",
    Usage: "make an explosive entrance",
    Action: func(c *cli.Context) error {
      fmt.Println("boom! I say!")
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Running this already gives you a ton of functionality, plus support for things
like subcommands and flags, which are covered below.

## Examples

Being a programmer can be a lonely job. Thankfully by the power of automation
that is not the case! Let's create a greeter app to fend off our demons of
loneliness!

Start by creating a directory named `greet`, and within it, add a file,
`greet.go` with the following code in it:

<!-- {
  "output": "Hello friend!"
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Name: "greet",
    Usage: "fight the loneliness!",
    Action: func(c *cli.Context) error {
      fmt.Println("Hello friend!")
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Install our command to the `$GOPATH/bin` directory:

```
$ go install
```

Finally run our new command:

```
$ greet
Hello friend!
```

cli also generates neat help text:

```
$ greet help
NAME:
    greet - fight the loneliness!

USAGE:
    greet [global options] command [command options] [arguments...]

COMMANDS:
    help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS
    --help, -h  show help (default: false)
```

### Arguments

You can lookup arguments by calling the `Args` function on `cli.Context`, e.g.:

<!-- {
  "output": "Hello \""
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Action: func(c *cli.Context) error {
      fmt.Printf("Hello %q", c.Args().Get(0))
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

### Flags

Setting and querying flags is simple.

<!-- {
  "output": "Hello Nefertiti"
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag {
      &cli.StringFlag{
        Name: "lang",
        Value: "english",
        Usage: "language for the greeting",
      },
    },
    Action: func(c *cli.Context) error {
      name := "Nefertiti"
      if c.NArg() > 0 {
        name = c.Args().Get(0)
      }
      if c.String("lang") == "spanish" {
        fmt.Println("Hola", name)
      } else {
        fmt.Println("Hello", name)
      }
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

You can also set a destination variable for a flag, to which the content will be
scanned.

<!-- {
  "output": "Hello someone"
} -->
``` go
package main

import (
  "log"
  "os"
  "fmt"

  "github.com/urfave/cli/v2"
)

func main() {
  var language string

  app := &cli.App{
    Flags: []cli.Flag {
      &cli.StringFlag{
        Name:        "lang",
        Value:       "english",
        Usage:       "language for the greeting",
        Destination: &language,
      },
    },
    Action: func(c *cli.Context) error {
      name := "someone"
      if c.NArg() > 0 {
        name = c.Args().Get(0)
      }
      if language == "spanish" {
        fmt.Println("Hola", name)
      } else {
        fmt.Println("Hello", name)
      }
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

See full list of flags at https://pkg.go.dev/github.com/urfave/cli/v2

#### Placeholder Values

Sometimes it's useful to specify a flag's value within the usage string itself.
Such placeholders are indicated with back quotes.

For example this:

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "&#45;&#45;config FILE, &#45;c FILE"
} -->
```go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name:    "config",
        Aliases: []string{"c"},
        Usage:   "Load configuration from `FILE`",
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Will result in help output like:

```
--config FILE, -c FILE   Load configuration from FILE
```

Note that only the first placeholder is used. Subsequent back-quoted words will
be left as-is.

#### Alternate Names

You can set alternate (or short) names for flags by providing a comma-delimited
list for the `Name`. e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "&#45;&#45;lang value, &#45;l value.*language for the greeting.*default: \"english\""
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag {
      &cli.StringFlag{
        Name:    "lang",
        Aliases: []string{"l"},
        Value:   "english",
        Usage:   "language for the greeting",
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

That flag can then be set with `--lang spanish` or `-l spanish`. Note that
giving two different forms of the same flag in the same command invocation is an
error.

#### Ordering

Flags for the application and commands are shown in the order they are defined.
However, it's possible to sort them from outside this library by using `FlagsByName`
or `CommandsByName` with `sort`.

For example this:

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "add a task to the list\n.*complete a task on the list\n.*\n\n.*\n.*Load configuration from FILE\n.*Language for the greeting.*"
} -->
``` go
package main

import (
  "log"
  "os"
  "sort"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "lang",
            Aliases: []string{"l"},
            Value:   "english",
            Usage:   "Language for the greeting",
        },
        &cli.StringFlag{
            Name:    "config",
            Aliases: []string{"c"},
            Usage:   "Load configuration from `FILE`",
        },
    },
    Commands: []*cli.Command{
      {
        Name:    "complete",
        Aliases: []string{"c"},
        Usage:   "complete a task on the list",
        Action:  func(c *cli.Context) error {
          return nil
        },
      },
      {
        Name:    "add",
        Aliases: []string{"a"},
        Usage:   "add a task to the list",
        Action:  func(c *cli.Context) error {
          return nil
        },
      },
    },
  }

  sort.Sort(cli.FlagsByName(app.Flags))
  sort.Sort(cli.CommandsByName(app.Commands))

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Will result in help output like:

```
--config FILE, -c FILE  Load configuration from FILE
--lang value, -l value  Language for the greeting (default: "english")
```

#### Values from the Environment

You can also have the default value set from the environment via `EnvVars`.  e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "language for the greeting.*APP_LANG"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag {
      &cli.StringFlag{
        Name:    "lang",
        Aliases: []string{"l"},
        Value:   "english",
        Usage:   "language for the greeting",
        EnvVars: []string{"APP_LANG"},
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

If `EnvVars` contains more than one string, the first environment variable that
resolves is used.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "language for the greeting.*LEGACY_COMPAT_LANG.*APP_LANG.*LANG"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name:    "lang",
        Aliases: []string{"l"},
        Value:   "english",
        Usage:   "language for the greeting",
        EnvVars: []string{"LEGACY_COMPAT_LANG", "APP_LANG", "LANG"},
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

#### Values from files

You can also have the default value set from file via `FilePath`.  e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "password for the mysql database"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    &cli.StringFlag{
      Name: "password",
      Aliases: []string{"p"},
      Usage: "password for the mysql database",
      FilePath: "/etc/mysql/password",
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Note that default values set from file (e.g. `FilePath`) take precedence over
default values set from the environment (e.g. `EnvVar`).

#### Values from alternate input sources (YAML, TOML, and others)

There is a separate package altsrc that adds support for getting flag values
from other file input sources.

Currently supported input source formats:
* YAML
* JSON
* TOML

In order to get values for a flag from an alternate input source the following
code would be added to wrap an existing cli.Flag like below:

``` go
  altsrc.NewIntFlag(&cli.IntFlag{Name: "test"})
```

Initialization must also occur for these flags. Below is an example initializing
getting data from a yaml file below.

``` go
  command.Before = altsrc.InitInputSourceWithContext(command.Flags, NewYamlSourceFromFlagFunc("load"))
```

The code above will use the "load" string as a flag name to get the file name of
a yaml file from the cli.Context.  It will then use that file name to initialize
the yaml input source for any flags that are defined on that command.  As a note
the "load" flag used would also have to be defined on the command flags in order
for this code snippet to work.

Currently only YAML, JSON, and TOML files are supported but developers can add support
for other input sources by implementing the altsrc.InputSourceContext for their
given sources.

Here is a more complete sample of a command using YAML support:

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "&#45&#45;test value.*default: 0"
} -->
``` go
package main

import (
  "fmt"
  "os"

  "github.com/urfave/cli/v2"
  "github.com/urfave/cli/v2/altsrc"
)

func main() {
  flags := []cli.Flag{
    altsrc.NewIntFlag(&cli.IntFlag{Name: "test"}),
    &cli.StringFlag{Name: "load"},
  }

  app := &cli.App{
    Action: func(c *cli.Context) error {
      fmt.Println("--test value.*default: 0")
      return nil
    },
    Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("load")),
    Flags: flags,
  }

  app.Run(os.Args)
}
```

#### Required Flags

You can make a flag required by setting the `Required` field to `true`. If a user
does not provide a required flag, they will be shown an error message.

Take for example this app that requires the `lang` flag:

<!-- {
  "error": "Required flag \"lang\" not set"
} -->
```go
package main

import (
  "log"
  "os"
  "github.com/urfave/cli/v2"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    &cli.StringFlag{
      Name: "lang",
      Value: "english",
      Usage: "language for the greeting",
      Required: true,
    },
  }

  app.Action = func(c *cli.Context) error {
    var output string
    if c.String("lang") == "spanish" {
      output = "Hola"
    } else {
      output = "Hello"
    }
    fmt.Println(output)
    return nil
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

If the app is run without the `lang` flag, the user will see the following message

```
Required flag "lang" not set
```

#### Default Values for help output

Sometimes it's useful to specify a flag's default help-text value within the flag declaration. This can be useful if the default value for a flag is a computed value. The default value can be set via the `DefaultText` struct field.

For example this:

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "&#45;&#45;port value"
} -->
```go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.IntFlag{
        Name:    "port",
        Usage:   "Use a randomized port",
        Value: 0,
        DefaultText: "random",
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Will result in help output like:

```
--port value  Use a randomized port (default: random)
```

#### Precedence

The precedence for flag value sources is as follows (highest to lowest):

0. Command line flag value from user
0. Environment variable (if specified)
0. Configuration file (if specified)
0. Default defined on the flag

### Subcommands

Subcommands can be defined for a more git-like command line app.

<!-- {
  "args": ["template", "add"],
  "output": "new task template: .+"
} -->
```go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Commands: []*cli.Command{
      {
        Name:    "add",
        Aliases: []string{"a"},
        Usage:   "add a task to the list",
        Action:  func(c *cli.Context) error {
          fmt.Println("added task: ", c.Args().First())
          return nil
        },
      },
      {
        Name:    "complete",
        Aliases: []string{"c"},
        Usage:   "complete a task on the list",
        Action:  func(c *cli.Context) error {
          fmt.Println("completed task: ", c.Args().First())
          return nil
        },
      },
      {
        Name:        "template",
        Aliases:     []string{"t"},
        Usage:       "options for task templates",
        Subcommands: []*cli.Command{
          {
            Name:  "add",
            Usage: "add a new template",
            Action: func(c *cli.Context) error {
              fmt.Println("new task template: ", c.Args().First())
              return nil
            },
          },
          {
            Name:  "remove",
            Usage: "remove an existing template",
            Action: func(c *cli.Context) error {
              fmt.Println("removed task template: ", c.Args().First())
              return nil
            },
          },
        },
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

### Subcommands categories

For additional organization in apps that have many subcommands, you can
associate a category for each command to group them together in the help
output.

E.g.

```go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Commands: []*cli.Command{
      {
        Name: "noop",
      },
      {
        Name:     "add",
        Category: "template",
      },
      {
        Name:     "remove",
        Category: "template",
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Will include:

```
COMMANDS:
  noop

  Template actions:
    add
    remove
```

### Exit code

Calling `App.Run` will not automatically call `os.Exit`, which means that by
default the exit code will "fall through" to being `0`.  An explicit exit code
may be set by returning a non-nil error that fulfills `cli.ExitCoder`, *or* a
`cli.MultiError` that includes an error that fulfills `cli.ExitCoder`, e.g.:
<!-- {
  "error": "Ginger croutons are not in the soup"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.BoolFlag{
        Name:  "ginger-crouton",
        Usage: "is it in the soup?",
      },
    },
    Action: func(ctx *cli.Context) error {
      if !ctx.Bool("ginger-crouton") {
        return cli.Exit("Ginger croutons are not in the soup", 86)
      }
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

### Combining short options

Traditional use of options using their shortnames look like this:

```
$ cmd -s -o -m "Some message"
```

Suppose you want users to be able to combine options with their shortnames. This
can be done using the `UseShortOptionHandling` bool in your app configuration,
or for individual commands by attaching it to the command configuration. For
example:

<!-- {
  "args": ["short", "&#45;som", "Some message"],
  "output": "serve: true\noption: true\nmessage: Some message\n"
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{}
  app.UseShortOptionHandling = true
  app.Commands = []*cli.Command{
    {
      Name:  "short",
      Usage: "complete a task on the list",
      Flags: []cli.Flag{
        &cli.BoolFlag{Name: "serve", Aliases: []string{"s"}},
        &cli.BoolFlag{Name: "option", Aliases: []string{"o"}},
        &cli.StringFlag{Name: "message", Aliases: []string{"m"}},
      },
      Action: func(c *cli.Context) error {
        fmt.Println("serve:", c.Bool("serve"))
        fmt.Println("option:", c.Bool("option"))
        fmt.Println("message:", c.String("message"))
        return nil
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

If your program has any number of bool flags such as `serve` and `option`, and
optionally one non-bool flag `message`, with the short options of `-s`, `-o`,
and `-m` respectively, setting `UseShortOptionHandling` will also support the
following syntax:

```
$ cmd -som "Some message"
```

If you enable `UseShortOptionHandling`, then you must not use any flags that
have a single leading `-` or this will result in failures. For example,
`-option` can no longer be used. Flags with two leading dashes (such as
`--options`) are still valid.

### Bash Completion

You can enable completion commands by setting the `EnableBashCompletion`
flag on the `App` object to `true`.  By default, this setting will allow auto-completion 
for an app's subcommands, but you can write your own completion methods for
the App or its subcommands as well.

#### Default auto-completion

```go
package main
import (
	"fmt"
	"log"
	"os"
	"github.com/urfave/cli/v2"
)
func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a task to the list",
			Action: func(c *cli.Context) error {
				fmt.Println("added task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task on the list",
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:    "template",
			Aliases: []string{"t"},
			Usage:   "options for task templates",
			Subcommands: []*cli.Command{
				{
					Name:  "add",
					Usage: "add a new template",
					Action: func(c *cli.Context) error {
						fmt.Println("new task template: ", c.Args().First())
						return nil
					},
				},
				{
					Name:  "remove",
					Usage: "remove an existing template",
					Action: func(c *cli.Context) error {
						fmt.Println("removed task template: ", c.Args().First())
						return nil
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
```
![](/docs/v2/images/default-bash-autocomplete.gif)

#### Custom auto-completion
<!-- {
  "args": ["complete", "&#45;&#45;generate&#45;bash&#45;completion"],
  "output": "laundry"
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  tasks := []string{"cook", "clean", "laundry", "eat", "sleep", "code"}

  app := &cli.App{
    EnableBashCompletion: true,
    Commands: []*cli.Command{
      {
        Name:    "complete",
        Aliases: []string{"c"},
        Usage:   "complete a task on the list",
        Action: func(c *cli.Context) error {
           fmt.Println("completed task: ", c.Args().First())
           return nil
        },
        BashComplete: func(c *cli.Context) {
          // This will complete if no args are passed
          if c.NArg() > 0 {
            return
          }
          for _, t := range tasks {
            fmt.Println(t)
          }
        },
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```
![](/docs/v2/images/custom-bash-autocomplete.gif)

#### Enabling

To enable auto-completion for the current shell session, a bash script,
`autocomplete/bash_autocomplete` is included in this repo.

To use `autocomplete/bash_autocomplete` set an environment variable named `PROG` to 
the name of your program and then `source` the `autocomplete/bash_autocomplete` file.

For example, if your cli program is called `myprogram`:

`PROG=myprogram source path/to/cli/autocomplete/bash_autocomplete`

Auto-completion is now enabled for the current shell, but will not persist into a new shell.

#### Distribution and Persistent Autocompletion

Copy `autocomplete/bash_autocomplete` into `/etc/bash_completion.d/` and rename
it to the name of the program you wish to add autocomplete support for (or
automatically install it there if you are distributing a package). Don't forget
to source the file or restart your shell to activate the auto-completion.

```
sudo cp path/to/autocomplete/bash_autocomplete /etc/bash_completion.d/<myprogram>
source /etc/bash_completion.d/<myprogram>
```

Alternatively, you can just document that users should `source` the generic
`autocomplete/bash_autocomplete` and set `$PROG` within their bash configuration 
file, adding these lines:

```
PROG=<myprogram>
source path/to/cli/autocomplete/bash_autocomplete
```
Keep in mind that if they are enabling auto-completion for more than one program, 
they will need to set `PROG` and source `autocomplete/bash_autocomplete` for each 
program, like so:

```
PROG=<program1>
source path/to/cli/autocomplete/bash_autocomplete
PROG=<program2>
source path/to/cli/autocomplete/bash_autocomplete
```

#### Customization

The default shell completion flag (`--generate-bash-completion`) is defined as
`cli.EnableBashCompletion`, and may be redefined if desired, e.g.:

<!-- {
  "args": ["&#45;&#45;generate&#45;bash&#45;completion"],
  "output": "wat\nhelp\nh"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    EnableBashCompletion: true,
    Commands: []*cli.Command{
      {
        Name: "wat",
      },
    },
  }
  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

#### ZSH Support
Auto-completion for ZSH is also supported using the `autocomplete/zsh_autocomplete` 
file included in this repo. Two environment variables are used, `PROG` and `_CLI_ZSH_AUTOCOMPLETE_HACK`. 
Set `PROG` to the program name as before, set `_CLI_ZSH_AUTOCOMPLETE_HACK` to `1`, and 
then `source path/to/autocomplete/zsh_autocomplete`. Adding the following lines to your ZSH 
configuration file (usually `.zshrc`) will allow the auto-completion to persist across new shells:

```
PROG=<myprogram>
_CLI_ZSH_AUTOCOMPLETE_HACK=1
source  path/to/autocomplete/zsh_autocomplete
```
#### ZSH default auto-complete example
![](/docs/v2/images/default-zsh-autocomplete.gif)
#### ZSH custom auto-complete example
![](/docs/v2/images/custom-zsh-autocomplete.gif)

#### PowerShell Support
Auto-completion for PowerShell is also supported using the `autocomplete/powershell_autocomplete.ps1` 
file included in this repo. 

Rename the script to `<my program>.ps1` and move it anywhere in your file system.
The location of script does not matter, only the file name of the script has to match
the your program's binary name. 

To activate it, enter `& path/to/autocomplete/<my program>.ps1`

To persist across new shells, open the PowerShell profile (with `code $profile` or `notepad $profile`)
and add the line:
```
& path/to/autocomplete/<my program>.ps1
```


### Generated Help Text

The default help flag (`-h/--help`) is defined as `cli.HelpFlag` and is checked
by the cli internals in order to print generated help text for the app, command,
or subcommand, and break execution.

#### Customization

All of the help text generation may be customized, and at multiple levels.  The
templates are exposed as variables `AppHelpTemplate`, `CommandHelpTemplate`, and
`SubcommandHelpTemplate` which may be reassigned or augmented, and full override
is possible by assigning a compatible func to the `cli.HelpPrinter` variable,
e.g.:

<!-- {
  "output": "Ha HA.  I pwnd the help!!1"
} -->
``` go
package main

import (
  "fmt"
  "io"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  // EXAMPLE: Append to an existing template
  cli.AppHelpTemplate = fmt.Sprintf(`%s

WEBSITE: http://awesometown.example.com

SUPPORT: support@awesometown.example.com

`, cli.AppHelpTemplate)

  // EXAMPLE: Override a template
  cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`

  // EXAMPLE: Replace the `HelpPrinter` func
  cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
    fmt.Println("Ha HA.  I pwnd the help!!1")
  }

  (&cli.App{}).Run(os.Args)
}
```

The default flag may be customized to something other than `-h/--help` by
setting `cli.HelpFlag`, e.g.:

<!-- {
  "args": ["&#45;&#45halp"],
  "output": "haaaaalp.*HALP"
} -->
``` go
package main

import (
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  cli.HelpFlag = &cli.BoolFlag{
    Name: "haaaaalp",
    Aliases: []string{"halp"},
    Usage: "HALP",
    EnvVars: []string{"SHOW_HALP", "HALPPLZ"},
  }

  (&cli.App{}).Run(os.Args)
}
```

### Version Flag

The default version flag (`-v/--version`) is defined as `cli.VersionFlag`, which
is checked by the cli internals in order to print the `App.Version` via
`cli.VersionPrinter` and break execution.

#### Customization

The default flag may be customized to something other than `-v/--version` by
setting `cli.VersionFlag`, e.g.:

<!-- {
  "args": ["&#45;&#45print-version"],
  "output": "partay version v19\\.99\\.0"
} -->
``` go
package main

import (
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  cli.VersionFlag = &cli.BoolFlag{
    Name: "print-version",
    Aliases: []string{"V"},
    Usage: "print only the version",
  }

  app := &cli.App{
    Name: "partay",
    Version: "v19.99.0",
  }
  app.Run(os.Args)
}
```

Alternatively, the version printer at `cli.VersionPrinter` may be overridden, e.g.:

<!-- {
  "args": ["&#45;&#45version"],
  "output": "version=v19\\.99\\.0 revision=fafafaf"
} -->
``` go
package main

import (
  "fmt"
  "os"

  "github.com/urfave/cli/v2"
)

var (
  Revision = "fafafaf"
)

func main() {
  cli.VersionPrinter = func(c *cli.Context) {
    fmt.Printf("version=%s revision=%s\n", c.App.Version, Revision)
  }

  app := &cli.App{
    Name: "partay",
    Version: "v19.99.0",
  }
  app.Run(os.Args)
}
```

### Timestamp Flag

Using the timestamp flag is simple. Please refer to [`time.Parse`](https://golang.org/pkg/time/#example_Parse) to get possible formats.

<!-- {
  "args": ["&#45;&#45;meeting", "2019-08-12T15:04:05"],
  "output": "2019\\-08\\-12 15\\:04\\:05 \\+0000 UTC"
} -->
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag {
      &cli.TimestampFlag{Name: "meeting", Layout: "2006-01-02T15:04:05"},
    },
    Action: func(c *cli.Context) error {
      fmt.Printf("%s", c.Timestamp("meeting").String())
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

In this example the flag could be used like this : 

`myapp --meeting 2019-08-12T15:04:05`

Side note: quotes may be necessary around the date depending on your layout (if you have spaces for instance)

### Full API Example

**Notice**: This is a contrived (functioning) example meant strictly for API
demonstration purposes.  Use of one's imagination is encouraged.

<!-- {
  "output": "made it!\nPhew!"
} -->
``` go
package main

import (
  "errors"
  "flag"
  "fmt"
  "io"
  "io/ioutil"
  "os"
  "time"

  "github.com/urfave/cli/v2"
)

func init() {
  cli.AppHelpTemplate += "\nCUSTOMIZED: you bet ur muffins\n"
  cli.CommandHelpTemplate += "\nYMMV\n"
  cli.SubcommandHelpTemplate += "\nor something\n"

  cli.HelpFlag = &cli.BoolFlag{Name: "halp"}
  cli.VersionFlag = &cli.BoolFlag{Name: "print-version", Aliases: []string{"V"}}

  cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
    fmt.Fprintf(w, "best of luck to you\n")
  }
  cli.VersionPrinter = func(c *cli.Context) {
    fmt.Fprintf(c.App.Writer, "version=%s\n", c.App.Version)
  }
  cli.OsExiter = func(c int) {
    fmt.Fprintf(cli.ErrWriter, "refusing to exit %d\n", c)
  }
  cli.ErrWriter = ioutil.Discard
  cli.FlagStringer = func(fl cli.Flag) string {
    return fmt.Sprintf("\t\t%s", fl.Names()[0])
  }
}

type hexWriter struct{}

func (w *hexWriter) Write(p []byte) (int, error) {
  for _, b := range p {
    fmt.Printf("%x", b)
  }
  fmt.Printf("\n")

  return len(p), nil
}

type genericType struct {
  s string
}

func (g *genericType) Set(value string) error {
  g.s = value
  return nil
}

func (g *genericType) String() string {
  return g.s
}

func main() {
  app := &cli.App{
    Name: "kənˈtrīv",
    Version: "v19.99.0",
    Compiled: time.Now(),
    Authors: []*cli.Author{
      &cli.Author{
        Name:  "Example Human",
        Email: "human@example.com",
      },
    },
    Copyright: "(c) 1999 Serious Enterprise",
    HelpName: "contrive",
    Usage: "demonstrate available API",
    UsageText: "contrive - demonstrating the available API",
    ArgsUsage: "[args and such]",
    Commands: []*cli.Command{
      &cli.Command{
        Name:        "doo",
        Aliases:     []string{"do"},
        Category:    "motion",
        Usage:       "do the doo",
        UsageText:   "doo - does the dooing",
        Description: "no really, there is a lot of dooing to be done",
        ArgsUsage:   "[arrgh]",
        Flags: []cli.Flag{
          &cli.BoolFlag{Name: "forever", Aliases: []string{"forevvarr"}},
        },
        Subcommands: []*cli.Command{
          &cli.Command{
            Name:   "wop",
            Action: wopAction,
          },
        },
        SkipFlagParsing: false,
        HideHelp:        false,
        Hidden:          false,
        HelpName:        "doo!",
        BashComplete: func(c *cli.Context) {
          fmt.Fprintf(c.App.Writer, "--better\n")
        },
        Before: func(c *cli.Context) error {
          fmt.Fprintf(c.App.Writer, "brace for impact\n")
          return nil
        },
        After: func(c *cli.Context) error {
          fmt.Fprintf(c.App.Writer, "did we lose anyone?\n")
          return nil
        },
        Action: func(c *cli.Context) error {
          c.Command.FullName()
          c.Command.HasName("wop")
          c.Command.Names()
          c.Command.VisibleFlags()
          fmt.Fprintf(c.App.Writer, "dodododododoodododddooooododododooo\n")
          if c.Bool("forever") {
            c.Command.Run(c)
          }
          return nil
        },
        OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
          fmt.Fprintf(c.App.Writer, "for shame\n")
          return err
        },
      },
    },
    Flags: []cli.Flag{
      &cli.BoolFlag{Name: "fancy"},
      &cli.BoolFlag{Value: true, Name: "fancier"},
      &cli.DurationFlag{Name: "howlong", Aliases: []string{"H"}, Value: time.Second * 3},
      &cli.Float64Flag{Name: "howmuch"},
      &cli.GenericFlag{Name: "wat", Value: &genericType{}},
      &cli.Int64Flag{Name: "longdistance"},
      &cli.Int64SliceFlag{Name: "intervals"},
      &cli.IntFlag{Name: "distance"},
      &cli.IntSliceFlag{Name: "times"},
      &cli.StringFlag{Name: "dance-move", Aliases: []string{"d"}},
      &cli.StringSliceFlag{Name: "names", Aliases: []string{"N"}},
      &cli.UintFlag{Name: "age"},
      &cli.Uint64Flag{Name: "bigage"},
    },
    EnableBashCompletion: true,
    HideHelp: false,
    HideVersion: false,
    BashComplete: func(c *cli.Context) {
      fmt.Fprintf(c.App.Writer, "lipstick\nkiss\nme\nlipstick\nringo\n")
    },
    Before: func(c *cli.Context) error {
      fmt.Fprintf(c.App.Writer, "HEEEERE GOES\n")
      return nil
    },
    After: func(c *cli.Context) error {
      fmt.Fprintf(c.App.Writer, "Phew!\n")
      return nil
    },
    CommandNotFound: func(c *cli.Context, command string) {
      fmt.Fprintf(c.App.Writer, "Thar be no %q here.\n", command)
    },
    OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
      if isSubcommand {
        return err
      }

      fmt.Fprintf(c.App.Writer, "WRONG: %#v\n", err)
      return nil
    },
    Action: func(c *cli.Context) error {
      cli.DefaultAppComplete(c)
      cli.HandleExitCoder(errors.New("not an exit coder, though"))
      cli.ShowAppHelp(c)
      cli.ShowCommandCompletions(c, "nope")
      cli.ShowCommandHelp(c, "also-nope")
      cli.ShowCompletions(c)
      cli.ShowSubcommandHelp(c)
      cli.ShowVersion(c)

      fmt.Printf("%#v\n", c.App.Command("doo"))
      if c.Bool("infinite") {
      	c.App.Run([]string{"app", "doo", "wop"})
      }

      if c.Bool("forevar") {
      	c.App.RunAsSubcommand(c)
      }
      c.App.Setup()
      fmt.Printf("%#v\n", c.App.VisibleCategories())
      fmt.Printf("%#v\n", c.App.VisibleCommands())
      fmt.Printf("%#v\n", c.App.VisibleFlags())

      fmt.Printf("%#v\n", c.Args().First())
      if c.Args().Len() > 0 {
        fmt.Printf("%#v\n", c.Args().Get(1))
      }
      fmt.Printf("%#v\n", c.Args().Present())
      fmt.Printf("%#v\n", c.Args().Tail())

      set := flag.NewFlagSet("contrive", 0)
      nc := cli.NewContext(c.App, set, c)

      fmt.Printf("%#v\n", nc.Args())
      fmt.Printf("%#v\n", nc.Bool("nope"))
      fmt.Printf("%#v\n", !nc.Bool("nerp"))
      fmt.Printf("%#v\n", nc.Duration("howlong"))
      fmt.Printf("%#v\n", nc.Float64("hay"))
      fmt.Printf("%#v\n", nc.Generic("bloop"))
      fmt.Printf("%#v\n", nc.Int64("bonk"))
      fmt.Printf("%#v\n", nc.Int64Slice("burnks"))
      fmt.Printf("%#v\n", nc.Int("bips"))
      fmt.Printf("%#v\n", nc.IntSlice("blups"))
      fmt.Printf("%#v\n", nc.String("snurt"))
      fmt.Printf("%#v\n", nc.StringSlice("snurkles"))
      fmt.Printf("%#v\n", nc.Uint("flub"))
      fmt.Printf("%#v\n", nc.Uint64("florb"))

      fmt.Printf("%#v\n", nc.FlagNames())
      fmt.Printf("%#v\n", nc.IsSet("wat"))
      fmt.Printf("%#v\n", nc.Set("wat", "nope"))
      fmt.Printf("%#v\n", nc.NArg())
      fmt.Printf("%#v\n", nc.NumFlags())
      fmt.Printf("%#v\n", nc.Lineage()[1])
      nc.Set("wat", "also-nope")

      ec := cli.Exit("ohwell", 86)
      fmt.Fprintf(c.App.Writer, "%d", ec.ExitCode())
      fmt.Printf("made it!\n")
      return ec
    },
    Metadata: map[string]interface{}{
      "layers":          "many",
      "explicable":      false,
      "whatever-values": 19.99,
    },
  }

  if os.Getenv("HEXY") != "" {
    app.Writer = &hexWriter{}
    app.ErrWriter = &hexWriter{}
  }

  app.Run(os.Args)
}

func wopAction(c *cli.Context) error {
  fmt.Fprintf(c.App.Writer, ":wave: over here, eh\n")
  return nil
}
```
