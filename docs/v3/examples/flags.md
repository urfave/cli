---
tags:
  - v3
search:
  boost: 2
---

Setting and querying flags is simple.

<!-- {
  "output": "Hello Nefertiti"
} -->
```go
package main

import (
	"fmt"
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "lang",
				Value: "english",
				Usage: "language for the greeting",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := "Nefertiti"
			if cmd.NArg() > 0 {
				name = cmd.Args().Get(0)
			}
			if cmd.String("lang") == "spanish" {
				fmt.Println("Hola", name)
			} else {
				fmt.Println("Hello", name)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

You can also set a destination variable for a flag, to which the content will be
scanned. Note that if the `Value` is set for the flag, it will be shown as default,
and destination will be set to this value before parsing flag on the command line.

<!-- {
  "output": "Hello someone"
} -->
```go
package main

import (
	"fmt"
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	var language string

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lang",
				Value:       "english",
				Usage:       "language for the greeting",
				Destination: &language,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := "someone"
			if cmd.NArg() > 0 {
				name = cmd.Args().Get(0)
			}
			if language == "spanish" {
				fmt.Println("Hola", name)
			} else {
				fmt.Println("Hello", name)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

See full list of flags at https://pkg.go.dev/github.com/urfave/cli/v3

For bool flags you can specify the flag multiple times to get a count(e.g -v -v -v or -vvv)

> If you want to support the `-vvv` flag, you need to set `Command.UseShortOptionHandling`.

<!-- {
  "args": ["&#45;&#45;foo", "&#45;&#45;foo", "&#45;fff"],
  "output": "count 5"
} -->
```go
package main

import (
	"fmt"
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	var count int

	cmd := &cli.Command{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "foo",
				Usage:       "foo greeting",
				Aliases:     []string{"f"},
				Config: cli.BoolConfig{
					Count: &count,
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("count", count)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

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
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
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
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

That flag can then be set with `--lang spanish` or `-l spanish`. Note that
giving two different forms of the same flag in the same command invocation is an
error.

#### Multiple Values per Single Flag

Using a slice flag allows you to pass multiple values for a single flag; the values will be provided as a slice:

- `UintSliceFlag`
- `IntSliceFlag`
- `StringSliceFlag`

<!-- {
  "args": ["&#45;&#45;greeting", "Hello", "&#45;&#45;greeting", "Hola"],
  "output": "Hello, Hola"
} -->
```go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "greeting",
				Usage: "Pass multiple greetings",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println(strings.Join(cmd.StringSlice("greeting"), `, `))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Multiple values need to be passed as separate, repeating flags, e.g. `--greeting Hello --greeting Hola`.

#### Ordering

Flags for the application and commands are shown in the order they are defined.
However, it's possible to sort them from outside this library by using `FlagsByName`
or `CommandsByName` with `sort`.

For example this:

<!-- {
  "args": ["&#45;&#45;help"],
  "output": ".*Load configuration from FILE\n.*Language for the greeting.*"
} -->
```go
package main

import (
	"log"
	"os"
	"sort"
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
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			},
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a task to the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(cmd.Flags))

	if err := cmd.Run(context.Background(), os.Args); err != nil {
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

You can also have the default value set from the environment via `cli.EnvVars`.  e.g.

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
  "output": "&#45&#45;test value.*default: 0"
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
	"fmt"
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "lang",
				Value:    "english",
				Usage:    "language for the greeting",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			output := "Hello"
			if cmd.String("lang") == "spanish" {
				output = "Hola"
			}
			fmt.Println(output)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

If the command is run without the `lang` flag, the user will see the following message

```
Required flag "lang" not set
```

#### Default Values for help output

Sometimes it's useful to specify a flag's default help-text value within the
flag declaration. This can be useful if the default value for a flag is a
computed value. The default value can be set via the `DefaultText` struct field.

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
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Usage:       "Use a randomized port",
				Value:       0,
				DefaultText: "random",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Will result in help output like:

```
--port value  Use a randomized port (default: random)
```

#### Flag Actions

Handlers can be registered per flag which are triggered after a flag has been processed. 
This can be used for a variety of purposes, one of which is flag validation

<!-- {
  "args": ["&#45;&#45;port","70000"],
  "error": "Flag port value 70000 out of range[0-65535]"
} -->
```go
package main

import (
	"log"
	"os"
	"fmt"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Usage:       "Use a randomized port",
				Value:       0,
				DefaultText: "random",
				Action: func(ctx context.Context, cmd *cli.Command, v int64) error {
					if v >= 65536 {
						return fmt.Errorf("Flag port value %v out of range[0-65535]", v)
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Will result in help output like:

```
Flag port value 70000 out of range[0-65535]
```
