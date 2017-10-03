package sigctx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var ctx context.Context

func init() {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
}

// New signal-bound context.  Context terminates when either SIGINT or SIGTERM
// is caught.
func New() context.Context {
	return ctx
}
