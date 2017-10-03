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

// Ref increments the refcount
func (r *RefCtr) Ref() { atomic.AddInt32(&r.refcnt, 1) }

// Free decrements the refcount
func (r *RefCtr) Free() {
	if v := atomic.AddInt32(&r.refcnt, -1); v > 0 {
		r.cancel()
	}
}

// WithRefCount derives a context that will be cancelled when all references are
// freed.
func WithRefCount(c context.Context) (context.Context, *RefCtr) {
	c, cancel := context.WithCancel(c)
	return c, &RefCtr{cancel: cancel}
}
