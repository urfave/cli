---
tags:
  - v3
search:
  boost: 2
---

Lets add some arguments to our greeter app. This allows you to change the behaviour of
the app depending on what argument has been passed. You can lookup arguments by calling 
the `Args` function on `cli.Command`, e.g.:

<!-- {
  "args" : ["Friend"],
  "output": "Hello \"Friend\""
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

Running this program with an argument gives the following output

```sh-session
$ greet friend
Hello "Friend"
```

Any number of arguments can be passed to the greeter app. We can get the number of arguments
and each argument using the `Args`

<!-- {
  "args" : ["Friend", "1", "bar", "2.0"],
  "output": "Number of args : 4\nHello Friend 1 bar 2.0"
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
			fmt.Printf("Number of args : %d\n", cmd.Args().Len())
			var out string
			for i := 0; i < cmd.Args().Len(); i++ {
				out = out + fmt.Sprintf(" %v", cmd.Args().Get(i))
			}
			fmt.Printf("Hello%v", out)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Running this program with an argument gives the following output

```sh-session
$ greet Friend 1 bar 2.0
Number of args : 4
Hello Friend 1 bar 2.0
```
