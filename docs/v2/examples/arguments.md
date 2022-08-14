You can lookup arguments by calling the `Args` function on `cli.Context`, e.g.:

<!-- {
  "output": "Hello \""
} -->
```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("Hello %q", cCtx.Args().Get(0))
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
```
