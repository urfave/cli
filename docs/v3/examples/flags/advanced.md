---
tags:
  - v3
search:
  boost: 2
---

#### Alternate Names

You can set alternate (or short) names for flags by providing a list of strings for `Aliases`
e.g.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "&#45;&#45;lang string, &#45;l string.*language for the greeting.*default: \"english\""
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

As noted in the basics for flag, the simple flags allow only one value per flag and only the last
entered value on command line will be returned to user on query. 

`urfave/cli` also supports multi-value flags called slice flags. These flags can take multiple values of same type. 
In addition they can be invoked multiple times on the command line and values will be appended to original value
of the flag and returned to the user as a slice

- `UintSliceFlag`
- `IntSliceFlag`
- `StringSliceFlag`
- `FloatSliceFlag`

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

#### Count for bool flag

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

#### Required Flags

You can mark a flag as *required* by setting the `Required` field to `true`. If a user
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
  "output": "&#45;&#45;port int"
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
