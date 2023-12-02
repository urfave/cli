---
tags:
  - v3
search:
  boost: 2
---

The default version flag (`-v/--version`) is defined as `cli.VersionFlag`, which
is checked by the cli internals in order to print the `Command.Version` via
`cli.VersionPrinter` and break execution.

#### Customization

The default flag may be customized to something other than `-v/--version` by
setting `cli.VersionFlag`, e.g.:

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
