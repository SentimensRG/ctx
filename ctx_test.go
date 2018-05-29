package ctx

import (
	"context"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

func TestBindFunc(t *testing.T) {
	f := BindFunc(func(d Doner) { panic("called") })
	assert.Panic(t, func() { f.Bind(nil) }, "called")
}

func TestC(t *testing.T) {
	ch := make(chan struct{})
	close(ch)

	select {
	case <-C(ch).Done():
	default:
		t.Error("doner did not reflect closed state of channel")
	}
}

func TestDefer(t *testing.T) {
	ch := make(chan struct{})
	close(ch)

	chT := make(chan struct{})
	Defer(C(ch), func() { close(chT) })

	select {
	case <-chT:
	case <-time.After(time.Millisecond * 100):
		t.Error("deferred function was not called")
	}
}

func TestLink(t *testing.T) {
	ch := make(chan struct{})
	close(ch)

	select {
	case <-Link(C(ch), C(nil)):
	case <-time.After(time.Millisecond * 100):
		t.Error("link did not fire despite a Doner having fired")
	}
}

func TestJoin(t *testing.T) {
	ch := make(chan struct{})
	close(ch)

	c, cancel := context.WithCancel(context.Background())

	d := Join(C(ch), c)

	select {
	case <-d.Done():
		t.Error("premature firing of join-Doner")
	default:
	}

	cancel()

	select {
	case <-d.Done():
	case <-time.After(time.Millisecond * 100):
		t.Error("join-Doner did not fire despite all constituent Doners having fired")
	}
}
