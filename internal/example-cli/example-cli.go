// minimal example CLI used for binary size checking

package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	_ = (&cli.Command{}).Run(context.Background(), []string{""})
}
