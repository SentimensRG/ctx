// Package mergectx provides a utility for merging two context.Context objects,
// creating a child context that contains the union of both context's values.
package mergectx

import (
	"context"
	"sync"
	"time"

	"github.com/SentimensRG/ctx"
)

type cx struct {
	sync.Mutex
	c0, c1 context.Context
	cq     chan struct{}
	err    error
	dlFunc func() (time.Time, bool)
}

func newCtx(c0, c1 context.Context) *cx {
	return &cx{c0: c0, c1: c1, cq: make(chan struct{})}
}

// Link returns new context which is the child of child of two parents.  It is
// analogous to ctx.Link
//
// Done() channel is closed when one of parents contexts is done.
//
// Deadline() returns earliest deadline between parent contexts.
//
// Err() returns error from first done parent context.
//
// Value(key) looks for key in parent contexts. First found is returned.
func Link(c0, c1 context.Context) context.Context {
	c := newCtx(c0, c1)
	c.dlFunc = c.first
	go c.link()
	return c
}

// Join returns a new context which is the child of two parents.  It behaves
// analogously to ctx.Join.Deadline
//
// Done() channel is closed when both parent contexts are done.
//
// Deadline() returns the latest deadline of the two parent contexts.
//
// Err() returns the first error it finds. Note that this is not reliable.
//
// Value(key) looks for key in parent contexts.  First found is returned.
func Join(c0, c1 context.Context) context.Context {
	c := newCtx(c0, c1)
	c.dlFunc = c.last
	c.join()
	return c
}

func (c *cx) Deadline() (deadline time.Time, ok bool) { return c.dlFunc() }

func (c *cx) first() (deadline time.Time, ok bool) {
	if d1, ok1 := c.c0.Deadline(); !ok1 {
		deadline, ok = c.c1.Deadline()
	} else if d2, ok2 := c.c1.Deadline(); !ok2 {
		deadline, ok = d1, true
	} else if d2.Before(d1) {
		deadline, ok = d2, true
	} else {
		deadline, ok = d1, true
	}

	return
}

func (c *cx) last() (deadline time.Time, ok bool) {
	if d1, ok1 := c.c0.Deadline(); !ok1 {
		deadline, ok = c.c1.Deadline()
	} else if d2, ok2 := c.c1.Deadline(); !ok2 {
		deadline, ok = d1, true
	} else if d2.After(d1) {
		deadline, ok = d2, true
	} else {
		deadline, ok = d1, true
	}

	return
}

func (c *cx) Done() <-chan struct{} { return c.cq }

func (c *cx) Err() error {
	c.Lock()
	defer c.Unlock()
	return c.err
}

func (c *cx) Value(key interface{}) (v interface{}) {
	if v = c.c0.Value(key); v == nil {
		v = c.c1.Value(key)
	}
	return
}

func (c *cx) link() {
	var dc context.Context
	select {
	case <-c.c0.Done():
		dc = c.c0
	case <-c.c1.Done():
		dc = c.c1
	case <-c.cq:
		return
	}

	c.Lock()
	if c.err == nil {
		c.err = dc.Err()
		close(c.cq)
	}
	c.Unlock()
}

func (c *cx) join() {
	ctx.Defer(ctx.Join(c.c0, c.c1), func() {
		c.Lock()
		defer c.Unlock()

		if c.err == nil {
			if c.c0.Err() != nil {
				c.err = c.c0.Err()
			} else {
				c.err = c.c1.Err()
			}
		}
	})
}
