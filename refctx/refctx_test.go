package refctx

import (
	"context"
	"testing"
)

func TestRefCtx(t *testing.T) {
	t.Run("IsInitializedToZero", func(t *testing.T) {
		c, rc := WithRefCount(context.Background())

		rc.Incr()
		rc.Decr()

		select {
		case <-c.Done():
		default:
			t.Error("Incr followed by Decr did not release Context")
		}
	})

	t.Run("Add", func(t *testing.T) {
		var batchSize int32 = 10

		c, rc := WithRefCount(context.Background())
		rc.Add(batchSize)
		rc.Add(-batchSize)

		select {
		case <-c.Done():
		default:
			t.Error("Batch incr/decr did not release Context")
		}
	})
}

// t.Run("", func(t *testing.T) {
// })
