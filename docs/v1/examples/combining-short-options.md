---
tags:
  - v1
---

Traditional use of options using their shortnames look like this:

```
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
``` go
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()
  app.UseShortOptionHandling = true
  app.Commands = []cli.Command{
    {
      Name:  "short",
      Usage: "complete a task on the list",
      Flags: []cli.Flag{
        cli.BoolFlag{Name: "serve, s"},
        cli.BoolFlag{Name: "option, o"},
        cli.StringFlag{Name: "message, m"},
      },
      Action: func(c *cli.Context) error {
        fmt.Println("serve:", c.Bool("serve"))
        fmt.Println("option:", c.Bool("option"))
        fmt.Println("message:", c.String("message"))
        return nil
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

If your program has any number of bool flags such as `serve` and `option`, and
optionally one non-bool flag `message`, with the short options of `-s`, `-o`,
and `-m` respectively, setting `UseShortOptionHandling` will also support the
following syntax:

```
$ cmd -som "Some message"
```

If you enable `UseShortOptionHandling`, then you must not use any flags that
have a single leading `-` or this will result in failures. For example,
`-option` can no longer be used. Flags with two leading dashes (such as
`--options`) are still valid.
