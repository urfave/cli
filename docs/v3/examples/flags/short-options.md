---
tags:
  - v3
search:
  boost: 2
---

Traditional use of options using their shortnames look like this:

```sh-session
$ cmd -s -o -m "Some message"
```

Suppose you want users to be able to combine options with their shortnames. This
can be done using the `UseShortOptionHandling` bool in your app configuration,
or for individual commands by attaching it to the command configuration. For
example:

<!-- {
  "args": ["short", "&#45;som", "Some message"],
  "output": "serve: true\noption: true\nmessage: Some message\n"
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
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:  "short",
				Usage: "complete a task on the list",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "serve", Aliases: []string{"s"}},
					&cli.BoolFlag{Name: "option", Aliases: []string{"o"}},
					&cli.StringFlag{Name: "message", Aliases: []string{"m"}},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("serve:", cmd.Bool("serve"))
					fmt.Println("option:", cmd.Bool("option"))
					fmt.Println("message:", cmd.String("message"))
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

If your program has any number of bool flags such as `serve` and `option`, and
optionally one non-bool flag `message`, with the short options of `-s`, `-o`,
and `-m` respectively, setting `UseShortOptionHandling` will also support the
following syntax:

```sh-session
$ cmd -som "Some message"
```

If you enable `UseShortOptionHandling`, then you must not use any flags that
have a single leading `-` or this will result in failures. For example,
`-option` can no longer be used. Flags with two leading dashes (such as
`--options`) are still valid.
