---
tags:
  - v3
search:
  boost: 2
---

One of the philosophies behind cli is that an API should be playful and full of
discovery. So a cli app can be as little as one line of code in `main()`.

<!-- {
  "args": ["&#45;&#45;help"],
  "output": "A new cli application"
} -->
```go
package main

import (
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	(&cli.Command{}).Run(context.Background(), os.Args)
}
```

This app will run and show help text, but is not very useful.

```
$ wl-paste > hello.go
$ go build hello.go
$ ./hello
NAME:
   hello - A new cli application

USAGE:
   hello [global options]

GLOBAL OPTIONS:
   --help, -h  show help
```

Let's add an action to execute and some help documentation:

<!-- {
  "output": "boom! I say!"
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
		Name:  "boom",
		Usage: "make an explosive entrance",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("boom! I say!")
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```
The output of above code is 

```
boom! I say!
```

Running this already gives you a ton of functionality, plus support for things
like subcommands and flags, which are covered in a separate section. 
