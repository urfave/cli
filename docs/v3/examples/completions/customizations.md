---
tags:
  - v3
search:
  boost: 2
---

If default completion isn't sufficient additional customizations are available 

- custom auto-completion
- customizing completion command

#### Custom auto-completion
<!-- {
  "args": ["complete", "&#45;&#45;generate&#45;shell&#45;completion"],
  "output": "laundry"
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
	tasks := []string{"cook", "clean", "laundry", "eat", "sleep", "code"}

	cmd := &cli.Command{
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "complete",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("completed task: ", cmd.Args().First())
					return nil
				},
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					// This will complete if no args are passed
					if cmd.NArg() > 0 {
						return
					}
					for _, t := range tasks {
						fmt.Println(t)
					}
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```
![](../../images/custom-bash-autocomplete.gif)

#### Customize a completion command

By default, a completion command is hidden, meaning the command isn't included in the help message.
You can customize it by setting root Command's `ConfigureShellCompletionCommand`.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name: "greet",
		// EnableShellCompletion is unnecessary
		ConfigureShellCompletionCommand: func(cmd *cli.Command) { // cmd is a completion command
			cmd.Hidden = false // Make a completion command public
			cmd.Usage = "..." // Customize Usage
			cmd.Description = "..." // Customize Description
		},
		Commands: []*cli.Command{
			{
				Name:  "hello",
				Usage: "Say hello",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Hello")
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

#### Customization

The default shell completion flag (`--generate-shell-completion`) is defined as
`cli.EnableShellCompletion`, and may be redefined if desired, e.g.:

<!-- {
  "args": ["&#45;&#45;generate&#45;shell&#45;completion"],
  "output": "wat\nhelp\n"
} -->
```go
package main

import (
	"log"
	"os"
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name: "wat",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
```
