---
tags:
  - v3
search:
  boost: 2
---

You can lookup arguments by calling the `Args` function on `cli.Command`, e.g.:

<!-- {
  "output": "Hello \""
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("Hello %q", cmd.Args().Get(0))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```
