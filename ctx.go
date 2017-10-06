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
