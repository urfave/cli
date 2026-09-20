---
tags:
  - v3
search:
  boost: 2
---

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
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a task to the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("added task: ", cmd.Args().First())
					return nil
				},
			},
			{
				Name:    "complete",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("completed task: ", cmd.Args().First())
					return nil
				},
			},
			{
				Name:    "template",
				Aliases: []string{"t"},
				Usage:   "options for task templates",
				Commands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add a new template",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("new task template: ", cmd.Args().First())
							return nil
						},
					},
					{
						Name:  "remove",
						Usage: "remove an existing template",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("removed task template: ", cmd.Args().First())
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

#### Deprecated Commands

A command can be marked as *deprecated* by setting the `Deprecated` field to a
message explaining what to use instead. When the command is invoked, a warning is
written to the root command's `ErrWriter` and the command continues to run.

A deprecated command is still listed in help output. Set `Hidden: true` as well to
hide it.

<!-- {
  "args": ["rm", "task"],
  "output": "removed task: task"
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
		Commands: []*cli.Command{
			{
				Name:  "remove",
				Usage: "remove a task from the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("removed task:", cmd.Args().First())
					return nil
				},
			},
			{
				Name:       "rm",
				Usage:      "remove a task from the list",
				Deprecated: "use \"remove\" instead",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("removed task:", cmd.Args().First())
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

Running `rm task` prints the following warning before the output

```
Command "rm" is deprecated, use "remove" instead
```
