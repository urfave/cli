---
tags:
  - v3
search:
  boost: 2
---

# Path and Walk

`Command.Path()` returns the list of command names from the root to the
current command. `Command.Walk()` visits a command and every subcommand
recursively, calling a function on each.

## Path

The `Path()` method returns `[]string` where each element is a `Command.Name`
starting from the root. `FullName()` is equivalent to
`strings.Join(cmd.Path(), " ")`.

<!-- {
  "output": "top mid bottom"
} -->
```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func main() {
	subSubCmd := &cli.Command{
		Name: "bottom",
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Println(strings.Join(c.Path(), " "))
			return nil
		},
	}
	subCmd := &cli.Command{Name: "mid", Commands: []*cli.Command{subSubCmd}, Action: func(context.Context, *cli.Command) error { return nil }}
	cmd := &cli.Command{
		Name:     "top",
		Commands: []*cli.Command{subCmd},
	}

	cmd.Run(context.Background(), []string{"top", "mid", "bottom"})
}
```

```sh-session
$ go run .
top mid bottom
```

## Walk

`Walk()` traverses the command tree depth-first, visiting the command itself
first, then each subcommand recursively.

<!-- {
  "output": "top\nmid\nbottom"
} -->
```go
package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func main() {
	subSubCmd := &cli.Command{Name: "bottom", Action: func(context.Context, *cli.Command) error { return nil }}
	subCmd := &cli.Command{Name: "mid", Commands: []*cli.Command{subSubCmd}, Action: func(context.Context, *cli.Command) error { return nil }}
	cmd := &cli.Command{
		Name: "top",
		Commands: []*cli.Command{subCmd},
		Action: func(ctx context.Context, c *cli.Command) error { return nil },
	}

	cmd.Walk(func(c *cli.Command) error {
		fmt.Println(c.Name)
		return nil
	})
}
```

```sh-session
$ go run .
top
mid
bottom
```

### Short-circuiting

Return a non-nil error from the walk function to stop traversal early.

<!-- {
  "output": "top\nmid"
} -->
```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"
)

func main() {
	subSubCmd := &cli.Command{Name: "bottom", Action: func(context.Context, *cli.Command) error { return nil }}
	subCmd := &cli.Command{Name: "mid", Commands: []*cli.Command{subSubCmd}, Action: func(context.Context, *cli.Command) error { return nil }}
	cmd := &cli.Command{
		Name: "top",
		Commands: []*cli.Command{subCmd},
		Action: func(ctx context.Context, c *cli.Command) error { return nil },
	}

	err := cmd.Walk(func(c *cli.Command) error {
		fmt.Println(c.Name)
		if c.Name == "mid" {
			return errors.New("stop")
		}
		return nil
	})
	fmt.Println(err)
}
```

```sh-session
$ go run .
top
mid
stop
```

## Relation to Lineage

[`Lineage()`](https://pkg.go.dev/github.com/urfave/cli/v3#Command.Lineage)
returns the command and all its ancestors as `[]*Command` (child first).
`Path()` is similar but returns only the command names as `[]string` (root
first). Use `Lineage()` when you need access to the ancestor `*Command`
values; use `Path()` when you only need the names.
