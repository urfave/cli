---
tags:
  - v3
search:
  boost: 2
---

For additional organization in apps that have many subcommands, you can
associate a category for each command to group them together in the help
output, e.g.:

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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Will include:

```
COMMANDS:
  noop

  template:
    add
    remove
```
