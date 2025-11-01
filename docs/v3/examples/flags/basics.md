---
tags:
  - v3
search:
  boost: 2
---

Flags, also called options, can be used to control various behaviour of the app
by turning on/off capabilities or setting some configuration and so on. 
Setting and querying flags is done using the ```cmd.<FlagType>(<flagName>)```
function

Here is an example of using a StringFlag which accepts a string as its option value

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

This very simple program gives a lot of outputs depending on the value of the flag set.
```sh-session
$ greet
Hello Nefertiti
```
Note that the Value for the flag is the default value that will be used when the flag
is not set on the command line. Since in the above invocation no flag was specified the
value of the "lang" flag was default to "english". Now lets change the language

```sh-session
$ greet --lang spanish
Hola Nefertiti
```

Flag values can be provided with a space after the flag name or using the ```=``` sign
```sh-session
$ greet --lang=spanish
Hola Nefertiti
$ greet --lang=spanish my-friend
Hola my-friend
```

While the value of any flag can be retrieved using ```command.<flagType>``` sometimes
it is convenient to have the value of the flag automatically stored in a destination
variable for a flag. If the `Value` is set for the flag, it will be shown as default,
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

Note that most flag can be invoked multiple times but only the last value entered for the flag
will be provided to the user(with some exceptions. See flags-advanced.md)

The following basic flags are supported

- `IntFlag`
- `Int8Flag`
- `Int16Flag`
- `Int32Flag`
- `Int64Flag`
- `UintFlag`
- `Uint8Flag`
- `Uint16Flag`
- `Uint32Flag`
- `Uint64Flag`
- `BoolFlag`
- `DurationFlag`
- `FloatFlag`
- `Float32Flag`
- `Float64Flag`
- `StringFlag`
- `TimestampFlag`

For full list of flags see [`https://pkg.go.dev/github.com/urfave/cli/v3`](https://pkg.go.dev/github.com/urfave/cli/v3)

### Timestamp Flag ###

Using the timestamp flag is similar to other flags but special attention is need 
for the format to be provided to the flag . Please refer to
[`time.Parse`](https://golang.org/pkg/time/#example_Parse) to get possible
formats.

<!-- {
  "args": ["&#45;&#45;meeting", "2019-08-12T15:04:05"],
  "output": "2019\\-08\\-12 15\\:04\\:05 \\+0000 UTC"
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
			&cli.TimestampFlag{
				Name: "meeting", 
				Config: cli.TimestampConfig{
					Layouts: []string{"2006-01-02T15:04:05"},
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("%s", cmd.Timestamp("meeting").String())
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

In this example the flag could be used like this:

```sh-session
$ myapp --meeting 2019-08-12T15:04:05
```

When the layout doesn't contain timezones, timestamp will render with UTC. To
change behavior, a default timezone can be provided with flag definition:

```go
cmd := &cli.Command{
	Flags: []cli.Flag{
		&cli.TimestampFlag{
			Name: "meeting",
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				AvailableLayouts: []string{"2006-01-02T15:04:05"},
			},
		},
	},
}
```

(time.Local contains the system's local time zone.)

Side note: quotes may be necessary around the date depending on your layout (if
you have spaces for instance)

### Version Flags ###

A default version flag (`-v/--version`) is provided as `cli.VersionFlag`, which
is checked by the cli internals in order to print the `Command.Version` via
`cli.VersionPrinter` and break execution.

#### Customization

The default flag may be customized to something other than `-v/--version` by
setting fields of `cli.VersionFlag`, e.g.:

<!-- {
  "args": ["&#45;&#45print-version"],
  "output": "partay version v19\\.99\\.0"
} -->
```go
package main

import (
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

	cmd := &cli.Command{
		Name:    "partay",
		Version: "v19.99.0",
	}
	cmd.Run(context.Background(), os.Args)
}
```

Alternatively, the version printer at `cli.VersionPrinter` may be overridden,
e.g.:

<!-- {
  "args": ["&#45;&#45version"],
  "output": "version=v19\\.99\\.0 revision=fafafaf"
} -->
```go
package main

import (
	"fmt"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

var (
	Revision = "fafafaf"
)

func main() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("version=%s revision=%s\n", cmd.Root().Version, Revision)
	}

	cmd := &cli.Command{
		Name:    "partay",
		Version: "v19.99.0",
	}
	cmd.Run(context.Background(), os.Args)
}
```
