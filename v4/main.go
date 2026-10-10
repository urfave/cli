package cli

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// InterruptedCode is the exit code after SIGINT or SIGTERM, following the
// shell convention of 128 plus the signal number for SIGINT.
const InterruptedCode = 130

// Main runs cmd with os.Args and exits the process. The context passed to the
// action is cancelled on the first SIGINT or SIGTERM; a second one kills the
// process with the default handler.
func Main(ctx context.Context, cmd *Command) {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCtx.Done()
		stop()
	}()

	err := cmd.Run(sigCtx, os.Args)
	interrupted := sigCtx.Err() != nil && ctx.Err() == nil
	stop()

	if interrupted {
		if err != nil && !errors.Is(err, context.Canceled) {
			Handle(cmd, err)
		}
		os.Exit(InterruptedCode)
	}
	os.Exit(Handle(cmd, err))
}
