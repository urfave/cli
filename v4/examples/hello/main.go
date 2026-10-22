// Command hello is the smallest program built on the core package. It is the
// baseline for the binary size budget.
package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v4"
)

func main() {
	cli.Main(context.Background(), &cli.Command{
		Name:  "hello",
		Usage: "print a greeting",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Fprintln(cmd.Out(), "Hello")
			return nil
		},
	})
}
