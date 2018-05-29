package refctx

import (
	"context"
	"sync/atomic"

	"github.com/SentimensRG/ctx"
)

// RefCtr cancels a context when no references are held
type RefCtr struct {
	cancel func()
	refcnt uint32
}

// Incr increments the refcount
func (r *RefCtr) Incr() { r.Add(1) }

// Add i refcounts
func (r *RefCtr) Add(i uint32) {
	if v := atomic.AddUint32(&r.refcnt, i); v == 0 {
		r.cancel()
	}
}

// Decr decrements the refcount
func (r *RefCtr) Decr() { atomic.AddUint32(&r.refcnt, ^uint32(0)) }

// WithRefCount derives a ctx.C that will be cancelled when all references are
// freed
func WithRefCount(d ctx.Doner) (ctx.C, *RefCtr) {
	ch, cancel := ctx.WithCancel(d)
	return ch, &RefCtr{cancel: cancel}
}

// ContextWithRefCount derives a context that will be cancelled when all
// references are freed.
func ContextWithRefCount(c context.Context) (context.Context, *RefCtr) {
	c, cancel := context.WithCancel(c)
	return c, &RefCtr{cancel: cancel}
}
