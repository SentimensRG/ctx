package ctx

var heartbeat = struct{}{}

// Doner can be done
type Doner interface {
	Done() <-chan struct{}
}

// Tick returns a <-chan whose range ends when the underlying context cancels
func Tick(d Doner) <-chan struct{} {
	cq := make(chan struct{})
	go func() {
		for {
			select {
			case <-d.Done():
				close(cq)
				return
			case cq <- heartbeat:
			}
		}
	}()
	return cq
}

// Defer guarantees that a function will be called after a context has cancelled
func Defer(d Doner, cb func()) {
	go func() {
		<-d.Done()
		cb()
	}()
}
