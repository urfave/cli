// minimal example CLI used for binary size checking

package main

import (
	"github.com/urfave/cli/v3"
)

func main() {
	_ = (&cli.App{}).Run([]string{""})
}
