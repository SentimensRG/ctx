package refctx

import (
	"testing"
	"time"

	"github.com/SentimensRG/ctx"
	"github.com/stretchr/testify/assert"
)

func TestRefCtx(t *testing.T) {
	t.Run("IsInitializedToZero", func(t *testing.T) {
		c, rc := WithRefCount(ctx.C(make(chan struct{})))

		rc.Incr()
		assert.Equal(t, rc.refcnt, uint32(1))

		rc.Decr()
		assert.Zero(t, rc.refcnt)

		select {
		case <-c.Done():
		case <-time.After(time.Microsecond):
			t.Error("Incr followed by Decr did not release Doner")
		}
	})

	t.Run("Add", func(t *testing.T) {
		var batchSize uint32 = 10

		c, rc := WithRefCount(ctx.C(make(chan struct{})))

		rc.Add(batchSize)
		assert.Equal(t, rc.refcnt, batchSize)

		rc.Add(-batchSize)
		assert.Zero(t, rc.refcnt)

		select {
		case <-c.Done():
		case <-time.After(time.Microsecond):
			t.Error("Batch incr/decr did not release Doner")
		}
	})
}

// t.Run("", func(t *testing.T) {
// })
