package refctx

import (
	"context"
	"sync/atomic"
)

// RefCtr cancels a context when no references are held
type RefCtr struct {
	cancel context.CancelFunc
	refcnt int32
}

// Incr increments the refcount
func (r *RefCtr) Incr() { r.Add(1) }

// Add i refcounts
func (r *RefCtr) Add(i int32) {
	if v := atomic.AddInt32(&r.refcnt, i); v <= 0 {
		r.cancel()
	}
}

// Decr decrements the refcount
func (r *RefCtr) Decr() { r.Add(-1) }

// WithRefCount derives a context that will be cancelled when all references are
// freed.
func WithRefCount(c context.Context) (context.Context, *RefCtr) {
	c, cancel := context.WithCancel(c)
	return c, &RefCtr{cancel: cancel}
}
