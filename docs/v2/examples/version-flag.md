---
tags:
  - v2
search:
  boost: 2
---

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
```go
package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

	app := &cli.App{
		Name:    "partay",
		Version: "v19.99.0",
	}
	app.Run(os.Args)
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

	"github.com/urfave/cli/v2"
)

var (
	Revision = "fafafaf"
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("version=%s revision=%s\n", cCtx.App.Version, Revision)
	}

	app := &cli.App{
		Name:    "partay",
		Version: "v19.99.0",
	}
	app.Run(os.Args)
}
```
