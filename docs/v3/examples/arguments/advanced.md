---
tags:
  - v3
search:
  boost: 2
---

The [Basics] showed how to access arguments for a command. They are all retrieved as strings which is fine
but it we need to say get integers or timestamps the user would have to convert from string to desired type. 
To ease the burden on users the `cli` library offers predefined `{Type}Arg` and `{Type}Args` structure to faciliate this
The value of the argument can be retrieved using the `command.{Type}Arg()` function. For e.g

<!-- {
  "args" : ["10"],
  "output": "We got 10"
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
		Arguments: []cli.Argument{
			&cli.IntArg{
				Name: "someint",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("We got %d", cmd.IntArg("someint"))
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
$ greet 10
We got 10
```

Instead of using the `cmd.{Type}Arg()` function to retrieve the argument value a destination for the argument can be set
for e.g

<!-- {
  "args" : ["25"],
  "output": "We got 25"
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
	var ival int
	cmd := &cli.Command{
		Arguments: []cli.Argument{
			&cli.IntArg{
				Name: "someint",
				Destination: &ival,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("We got %d", ival)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Some of the basic types arguments suported are

- `FloatArg`
- `IntArg`
- `Int8Arg`
- `Int16Arg`
- `Int32Arg`
- `Int64Arg`
- `StringArg`
- `UintArg`
- `Uint8Arg`
- `Uint16Arg`
- `Uint32Arg`
- `Uint64Arg`
- `TimestampArg`

This is ok for single value arguments. Any number of these single value arguments can be concatenated in the `Arguments`
slice field of `Command`. 

The library also support multi value arguments for e.g

<!-- {
  "args" : ["10", "20"],
  "output": "We got &#91;10 20&#93;"
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
		Arguments: []cli.Argument{
			&cli.IntArgs{
				Name: "someint",
				Min: 0,
				Max: -1,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("We got ", cmd.IntArgs("someint"))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Some things to note about multi value arguments

1. They are of `{Type}Args` type rather than `{Type}Arg` to differentiate them from single value arguments
2. The `Max` field needs to be defined to a non zero value without which it cannot be parsed
3. `Max` field value needs to be greater than the `Min` field value

As with single value args the destination field can be set

<!-- {
  "args" : ["10", "30"],
  "output": "We got &#91;10 30&#93;"
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
	var ivals []int
	cmd := &cli.Command{
		Arguments: []cli.Argument{
			&cli.IntArgs{
				Name: "someint",
				Min: 0,
				Max: -1,
				Destination: &ivals,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("We got ", ivals)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```

Following multi value arguments are supported

- `FloatArgs`
- `IntArgs`
- `Int8Args`
- `Int16Args`
- `Int32Args`
- `Int64Args`
- `StringArgs`
- `UintArgs`
- `Uint8Args`
- `Uint16Args`
- `Uint32Args`
- `Uint64Args`
- `TimestampArgs`

It goes without saying that the chain of arguments set in the Arguments slice need to be consistent. Generally a glob
argument(`max=-1`) should be set for the argument at the end of the slice. To glob args we arent interested in we coud add
the following to the end of the Arguments slice and retrieve them as a slice

```
&StringArgs{
	Max: -1,
},
```