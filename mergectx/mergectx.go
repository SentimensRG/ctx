// Package mergectx provides a utility for merging two context.Context objects,
// creating a child context that contains the union of both context's values,
// and returns the earliest deadline
package mergectx

import (
	"context"
	"sync"
	"time"
)

type cx struct {
	sync.Mutex
	c0, c1 context.Context
	cq     chan struct{}
	err    error
}

// Merge returns new context which is the child of child of two parents.
//
// Done() channel is closed when one of parents contexts is done.
//
// Deadline() returns earliest deadline between parent contexts.
//
// Err() returns error from first done parent context.
//
// Value(key) looks for key in parent contexts. First found is returned.
func Merge(c0, c1 context.Context) (context.Context, context.CancelFunc) {
	c := &cx{c0: c0, c1: c1, cq: make(chan struct{})}
	go c.merge()
	return c, c.cancel
}

func (c *cx) Deadline() (deadline time.Time, ok bool) {

	if d1, ok1 := c.c0.Deadline(); !ok1 {
		deadline, ok = c.c1.Deadline()
	} else if d2, ok2 := c.c1.Deadline(); !ok2 {
		deadline, ok = d1, true
	} else if d2.Before(d1) {
		deadline, ok = d2, true
	} else {
		deadline, ok = d2, true
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

func (c *cx) merge() {
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
	if c.err != nil {
		c.Unlock()
		return
	}
	c.err = dc.Err()
	c.Unlock()
	close(c.cq)
}

func (c *cx) cancel() {
	c.Lock()
	defer c.Unlock()

	if c.err == nil {
		c.err = context.Canceled
	}

	close(c.cq)
}
