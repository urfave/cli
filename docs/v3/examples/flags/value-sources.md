---
tags:
  - v3
search:
  boost: 2
---

Flags can have their default values set from different sources. The following sources are
provided by default with `urfave/cli`
 - Environment
 - Text Files

The library also provides a framework for users to plugin their own implementation of value sources
to be fetched via other mechanisms(http and so on). 

In addition there is a `urfave/cli-altsrc` repo which hosts some common value sources to read 
 - YAML
 - JSON
 - TOML
from files or via http/https. 

#### Values from the Environment

To set a value from the environment use `cli.EnvVars`.  e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "language for the greeting.*APP_LANG"
} -->
```go
package main

import (
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Value:   "english",
				Usage:   "language for the greeting",
				Sources: cli.EnvVars("APP_LANG"),
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

If `cli.EnvVars` contains more than one string, the first environment variable that
resolves is used.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "language for the greeting.*LEGACY_COMPAT_LANG.*APP_LANG.*LANG"
} -->
```go
package main

import (
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Value:   "english",
				Usage:   "language for the greeting",
				Sources: cli.EnvVars("LEGACY_COMPAT_LANG", "APP_LANG", "LANG"),
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

#### Values from files

You can also have the default value set from file via `cli.File`.  e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "password for the mysql database"
} -->
```go
package main

import (
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"p"},
				Usage:    "password for the mysql database",
				Sources: cli.Files("/etc/mysql/password"),
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Note that default values are set in the same order as they are defined in the
`Sources` param. This allows the user to choose order of priority

#### Values from alternate input sources (YAML, TOML, and others)

There is a separate package altsrc that adds support for getting flag values
from other file input sources.

Currently supported input source formats:

- YAML
- JSON
- TOML

In order to get values for a flag from an alternate input source the following
code would be added to wrap an existing cli.Flag like below:

```go
  // --- >8 ---
  altsrc.NewIntFlag(&cli.IntFlag{Name: "test"})
```

Initialization must also occur for these flags. Below is an example initializing
getting data from a yaml file below.

```go
  // --- >8 ---
  command.Before = func(ctx context.Context, cmd *Command) (context.Context, error) {
	return ctx, altsrc.InitInputSourceWithContext(command.Flags, NewYamlSourceFromFlagFunc("load"))
  }
```

The code above will use the "load" string as a flag name to get the file name of
a yaml file from the cli.Context.  It will then use that file name to initialize
the yaml input source for any flags that are defined on that command.  As a note
the "load" flag used would also have to be defined on the command flags in order
for this code snippet to work.

Currently only YAML, JSON, and TOML files are supported but developers can add
support for other input sources by implementing the altsrc.InputSourceContext
for their given sources.

Here is a more complete sample of a command using YAML support:

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "&#45&#45;test int.*default: 0"
} -->
```go
package main

import (
	"context"
	"fmt"
	"os"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
)

func main() {
	flags := []cli.Flag{
		&cli.IntFlag{
			Name:    "test",
			Sources: altsrc.YAML("key", "/path/to/file"),
		},
		&cli.StringFlag{Name: "load"},
	}

	cmd := &cli.Command{
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("--test value.*default: 0")
			return nil
		},
		Flags: flags,
	}

	cmd.Run(context.Background(), os.Args)
}
```