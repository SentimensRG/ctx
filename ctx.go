package ctx

import "context"

var heartbeat = struct{}{}

// Tick returns a <-chan whose range ends when the underlying context cancels
func Tick(c context.Context) <-chan struct{} {
	cq := make(chan struct{})
	go func() {
		for {
			select {
			case <-c.Done():
				close(cq)
				return
			case cq <- heartbeat:
			}
		}
	}()
	return cq
}

// Defer guarantees that a function will be called after a context has cancelled
func Defer(c context.Context, cb func()) {
	<-c.Done()
	cb()
}
