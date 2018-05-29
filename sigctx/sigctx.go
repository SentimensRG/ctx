package sigctx

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/SentimensRG/ctx"
)

var (
	c    ctx.C
	once sync.Once
)

// New signal-bound ctx.C that terminates when either SIGINT or SIGTERM
// is caught.
func New() ctx.C {
	once.Do(func() {
		dc := make(chan struct{})
		c = dc

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			select {
			case <-ch:
				close(dc)
			case <-c.Done():
			}
		}()
	})

	return c
}

// NewContext calls New and wraps the result in a context.Context.  The result
// is a context that fires when either SIGINT or SIGTERM is caught.
func NewContext() context.Context {
	return ctx.AsContext(New())
}
