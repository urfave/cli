---
tags:
  - v1
---

For additional organization in apps that have many subcommands, you can
associate a category for each command to group them together in the help
output.

E.g.

```go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()

  app.Commands = []cli.Command{
    {
      Name: "noop",
    },
    {
      Name:     "add",
      Category: "Template actions",
    },
    {
      Name:     "remove",
      Category: "Template actions",
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

Will include:

```
COMMANDS:
  noop

  Template actions:
    add
    remove
```
