---
tags:
  - v2
search:
  boost: 2
---

For additional organization in apps that have many subcommands, you can
associate a category for each command to group them together in the help
output, e.g.:

<!-- {
  "output": ".*COMMANDS:\\n.*noop[ ]*\\n.*\\n[ ]*template:\\n[ ]*add[ ]*\\n[ ]*remove.*"
} -->
```go
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
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

	if err := app.Run(os.Args); err != nil {
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