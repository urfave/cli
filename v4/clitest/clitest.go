// Package clitest runs a command the way [cli.Main] would, but captures its
// output and exit code instead of exiting the process.
package clitest

import (
	"bytes"
	"context"

	"github.com/urfave/cli/v4"
)

// Result is what one run of a command produced.
type Result struct {
	Stdout string
	Stderr string
	Code   int
	Err    error
}

// Run runs cmd with args, which exclude the program name. It replaces the
// command's writers, so don't share cmd between parallel tests.
func Run(ctx context.Context, cmd *cli.Command, args ...string) Result {
	var stdout, stderr bytes.Buffer
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(ctx, append([]string{cmd.Name}, args...))
	code := cli.Handle(cmd, err)
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Code: code, Err: err}
}
