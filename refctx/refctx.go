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
func (r *RefCtr) Add(i int32) { atomic.AddInt32(&r.refcnt, i) }

// Decr decrements the refcount
func (r *RefCtr) Decr() {
	if v := atomic.AddInt32(&r.refcnt, -1); v <= 0 {
		r.cancel()
	}
}

// WithRefCount derives a context that will be cancelled when all references are
// freed.
func WithRefCount(c context.Context) (context.Context, *RefCtr) {
	c, cancel := context.WithCancel(c)
	return c, &RefCtr{cancel: cancel}
}
