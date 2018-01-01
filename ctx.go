package ctx

import (
	"context"
	"sync"
	"time"
)

// Doner can block until something is done
type Doner interface {
	Done() <-chan struct{}
}

// DoneChan is a basic implementation of Doner
type DoneChan <-chan struct{}

// Done returns a channel that receives when an action is complete
func (dc DoneChan) Done() <-chan struct{} { return dc }

// Lift takes a chan and wraps it in a Doner
func Lift(c <-chan struct{}) DoneChan { return DoneChan(c) }

// AsContext creates a context that fires when the Doner fires
func AsContext(d Doner) context.Context {
	c, cancel := context.WithCancel(context.Background())
	Defer(d, cancel)
	return c
}

// After time time has elapsed, the Doner fires
func After(d time.Duration) DoneChan {
	ch := make(chan struct{})
	go func() {
		<-time.After(d)
		close(ch)
	}()
	return ch
}

// WithCancel returns a new Doner that can be cancelled via the associated
// function
func WithCancel(d Doner) (Doner, func()) {
	cq := make(chan struct{})
	return Lift(cq), func() { close(cq) }
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
			case cq <- struct{}{}:
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

// Link returns a channel that fires if ANY of the constituent Doners have fired
func Link(doners ...Doner) DoneChan {
	c := make(chan struct{})
	cancel := func() { close(c) }

	var once sync.Once
	for _, d := range doners {
		Defer(d, func() { once.Do(cancel) })
	}

	return c
}

// Join returns a channel that receives when all constituent Doners have fired
func Join(doners ...Doner) DoneChan {
	var wg sync.WaitGroup
	wg.Add(len(doners))
	for _, d := range doners {
		Defer(d, wg.Done)
	}

	cq := make(chan struct{})
	go func() {
		wg.Wait()
		close(cq)
	}()
	return cq
}

// FTick calls a function in a loop until the Doner has fired
func FTick(d Doner, f func()) {
	for _ = range Tick(d) {
		f()
	}
}

// FTickInterval calls a function repeatedly at a given internval, until the Doner
// has fired.
func FTickInterval(d Doner, t time.Duration, f func()) {
	timer := time.AfterFunc(t, f)
	<-d.Done()
	timer.Stop()
}
