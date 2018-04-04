package ctx

import (
	"context"
	"sync"
	"time"
)

// Binder is the interface that wraps the basic Bind method.
// Bind executes logic until the Doner completes.  Implementations of Bind must
// not return until the Doner has completed.
type Binder interface {
	Bind(Doner)
}

// BindFunc is an adapter to allow the use of ordinary functions as Binders.
type BindFunc func(Doner)

// Bind executes logic until the Doner completes.  It satisfies the Binder
// interface.
func (f BindFunc) Bind(d Doner) { f(d) }

// Doner can block until something is done
type Doner interface {
	Done() <-chan struct{}
}

// C is a basic implementation of Doner
type C <-chan struct{}

// Done returns a channel that receives when an action is complete
func (dc C) Done() <-chan struct{} { return dc }

// AsContext creates a context that fires when the Doner fires
func AsContext(d Doner) context.Context {
	c, cancel := context.WithCancel(context.Background())
	Defer(d, cancel)
	return c
}

// After time time has elapsed, the Doner fires
func After(d time.Duration) C {
	ch := make(chan struct{})
	go func() {
		<-time.After(d)
		close(ch)
	}()
	return ch
}

// WithCancel returns a new Doner that can be cancelled via the associated
// function
func WithCancel(d Doner) (C, func()) {
	var closer sync.Once
	cq := make(chan struct{})
	cancel := func() { closer.Do(func() { close(cq) }) }
	return Link(d, C(cq)), cancel
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
func Link(doners ...Doner) C {
	c := make(chan struct{})
	cancel := func() { close(c) }

	var once sync.Once
	for _, d := range doners {
		Defer(d, func() { once.Do(cancel) })
	}

	return c
}

// Join returns a channel that receives when all constituent Doners have fired
func Join(doners ...Doner) C {
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

// FDone returns a doner that fires when the function returns or panics
func FDone(f func()) C {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		f()
	}()
	return ch
}
