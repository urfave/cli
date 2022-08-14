Calling `App.Run` will not automatically call `os.Exit`, which means that by
default the exit code will "fall through" to being `0`.  An explicit exit code
may be set by returning a non-nil error that fulfills `cli.ExitCoder`, *or* a
`cli.MultiError` that includes an error that fulfills `cli.ExitCoder`, e.g.:
<!-- {
  "error": "Ginger croutons are not in the soup"
} -->
``` go
package main

import (
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()
  app.Flags = []cli.Flag{
    cli.BoolFlag{
      Name:  "ginger-crouton",
      Usage: "Add ginger croutons to the soup",
    },
  }
  app.Action = func(ctx *cli.Context) error {
    if !ctx.Bool("ginger-crouton") {
      return cli.NewExitError("Ginger croutons are not in the soup", 86)
    }
    return nil
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```
