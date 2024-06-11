---
tags:
  - v1
---

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

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "lang",
      Value: "english",
      Usage: "language for the greeting",
    },
  }

  app.Action = func(c *cli.Context) error {
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

  "github.com/urfave/cli"
)

func main() {
  var language string

  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name:        "lang",
      Value:       "english",
      Usage:       "language for the greeting",
      Destination: &language,
    },
  }

  app.Action = func(c *cli.Context) error {
    name := "someone"
    if c.NArg() > 0 {
      name = c.Args()[0]
    }
    if language == "spanish" {
      fmt.Println("Hola", name)
    } else {
      fmt.Println("Hello", name)
    }
    return nil
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

See full list of flags at https://pkg.go.dev/github.com/urfave/cli

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

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag{
    cli.StringFlag{
      Name:  "config, c",
      Usage: "Load configuration from `FILE`",
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

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "lang, l",
      Value: "english",
      Usage: "language for the greeting",
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

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "lang, l",
      Value: "english",
      Usage: "Language for the greeting",
    },
    cli.StringFlag{
      Name: "config, c",
      Usage: "Load configuration from `FILE`",
    },
  }

  app.Commands = []cli.Command{
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

You can also have the default value set from the environment via `EnvVar`.  e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "language for the greeting.*APP_LANG"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "lang, l",
      Value: "english",
      Usage: "language for the greeting",
      EnvVar: "APP_LANG",
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

The `EnvVar` may also be given as a comma-delimited "cascade", where the first
environment variable that resolves is used as the default.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "language for the greeting.*LEGACY_COMPAT_LANG.*APP_LANG.*LANG"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "lang, l",
      Value: "english",
      Usage: "language for the greeting",
      EnvVar: "LEGACY_COMPAT_LANG,APP_LANG,LANG",
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

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "password, p",
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
  altsrc.NewIntFlag(cli.IntFlag{Name: "test"})
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
package notmain

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli"
  "github.com/urfave/cli/altsrc"
)

func main() {
  app := cli.NewApp()

  flags := []cli.Flag{
    altsrc.NewIntFlag(cli.IntFlag{Name: "test"}),
    cli.StringFlag{Name: "load"},
  }

  app.Action = func(c *cli.Context) error {
    fmt.Println("yaml ist rad")
    return nil
  }

  app.Before = altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("load"))
  app.Flags = flags

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

#### Precedence

The precedence for flag value sources is as follows (highest to lowest):

0. Command line flag value from user
0. Environment variable (if specified)
0. Configuration file (if specified)
0. Default defined on the flag
