---
tags:
  - v3
search:
  boost: 2
---

Instead of the user having to look through all the arguments one by one we can also specify the argument types and destination
fields so that the value can be directly retrieved by the user. Lets go back to the greeter app and specifying the types for arguments

<!-- {
  "args" : ["friend","1","bar","2.0"],
  "output": "friend-1-bar-2.0"
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
	var name, soap string
	var count int64
	var version float64
	cmd := &cli.Command{
		Arguments: []Argument{
			&StringArg{
				Name: "name",
				Min: 1,
				Max: 1,
				Destination: &name,
			},
			&IntArg{
				Name: "count",
				Min: 1,
				Max: 1,
				Destination: &count,
			},
			&StringArg{
				Name: "soap",
				Min: 1,
				Max: 1,
				Destination: &soap,
			},
			&FloatArg{
				Name: "version",
				Min: 1,
				Max: 1,
				Destination: &version,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("%s-%d-%s-%f", name, count, soap, version)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

