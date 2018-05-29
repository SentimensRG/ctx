package ctx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBindFunc(t *testing.T) {
	f := BindFunc(func(d Doner) { panic("called") })
	assert.Panics(t, func() { f.Bind(nil) }, "called")
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
	case <-time.After(time.Millisecond):
		t.Error("deferred function was not called")
	}
}

func TestLink(t *testing.T) {
	ch := make(chan struct{})
	close(ch)

	select {
	case <-Link(C(ch), C(nil)):
	case <-time.After(time.Millisecond):
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
	case <-time.After(time.Millisecond):
		t.Error("join-Doner did not fire despite all constituent Doners having fired")
	}
}

func TestCtx(t *testing.T) {
	ch := make(chan struct{})
	c := AsContext(C(ch)) // should not panic

	t.Run("Deadline", func(t *testing.T) {
		d, ok := c.Deadline()
		assert.Zero(t, d)
		assert.False(t, ok)
	})

	t.Run("Value", func(t *testing.T) {
		// should always be nil
		assert.Nil(t, c.Value(struct{}{}))
	})

	t.Run("Err", func(t *testing.T) {
		assert.NoError(t, c.Err())
		close(ch)
		assert.EqualError(t, c.Err(), context.Canceled.Error())
	})
}

func TestWithCancel(t *testing.T) {
	d, cancel := WithCancel(C(make(chan struct{})))

	t.Run("NotCancelled", func(t *testing.T) {
		select {
		case <-d.Done():
			t.Error("Doner expired by default")
		default:
		}
	})

	t.Run("Cancelled", func(t *testing.T) {
		cancel()
		select {
		case <-d.Done():
		default:
			t.Error("not cancelled")
		}
	})

	t.Run("IdempotentCancel", func(t *testing.T) {
		cancel() // subsequent calls to cancel should not panic
	})

	t.Run("CloseUnderlyingDoner", func(t *testing.T) {
		ch := make(chan struct{})
		d, _ := WithCancel(C(ch))

		close(ch)

		select {
		case <-d.Done():
		case <-time.After(time.Millisecond):
			t.Error("not cancelled")
		}
	})
}

func TestTick(t *testing.T) {
	ch := make(chan struct{})
	tc := Tick(C(ch))

	t.Run("RecvWhileOpen", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			select {
			case <-tc:
			case <-time.After(time.Millisecond):
				t.Error("no tick")
			}
		}
	})

	t.Run("BlockWhenClose", func(t *testing.T) {
		close(ch)
		select {
		case <-tc:
			t.Error("failed to stop ticking")
		default:
		}
	})
}
