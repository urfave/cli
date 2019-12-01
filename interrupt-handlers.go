package cli

import "os"

var ExitOnInterrupt InterruptHandlerFunc = func(ctx *Context) {
	CancelContextOnInterrupt(ctx)
	<-ctx.Done()
	os.Exit(1)
}

var CancelContextOnInterrupt InterruptHandlerFunc = func(ctx *Context) {
	ctx.Cancel()
}
