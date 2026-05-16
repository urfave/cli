---
tags:
  - v3
search:
  boost: 2
---

Being a programmer can be a lonely job. Thankfully by the power of automation
that is not the case! Let's create a greeter app to fend off our demons of
loneliness!

Start by creating a directory named `greet`, and within it, add a file,
`greet.go` with the following code in it:

<!-- {
  "output": "Hello friend!"
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
		Name:  "greet",
		Usage: "fight the loneliness!",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("Hello friend!")
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Install our command to the `$GOPATH/bin` directory:

```sh-session
$ go install
```

Finally run our new command:

```sh-session
$ greet
Hello friend!
```

cli also generates neat help text:

```sh-session
$ greet help
NAME:
   greet - fight the loneliness!

USAGE:
   greet [global options]

GLOBAL OPTIONS:
   --help, -h  show help
```

In general a full help with flags and subcommands would give something like this
```
NAME:
    greet - fight the loneliness!

USAGE:
    greet [global options] command [command options] [arguments...]

COMMANDS:
    help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS
    --help, -h  show help (default: false)
```
