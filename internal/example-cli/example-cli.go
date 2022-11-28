// minimal example CLI used for binary size checking

package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	(&cli.App{}).RunContext(context.Background(), []string{""})
}
