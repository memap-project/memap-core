package vals

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounter_NewCounter(t *testing.T) {
	c := NewCounter()
	require.NotNil(t, c)

	assert.Equal(t, int64(0), c.GetValue())
	assert.Equal(t, int64(0), c.GetLimit())
	assert.Equal(t, int64(0), c.TTL())
	assert.False(t, c.IsExpired())
}

func TestCounter_Limit(t *testing.T) {
	t.Run("set limit", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(100)
		assert.Equal(t, int64(100), c.GetLimit())
	})

	t.Run("update limit", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(100)
		c.SetLimit(250)
		assert.Equal(t, int64(250), c.GetLimit())
	})
}

func TestCounter_IncrBy(t *testing.T) {
	t.Run("without limit", func(t *testing.T) {
		c := NewCounter()

		assert.True(t, c.IncrBy(5))
		assert.Equal(t, int64(5), c.GetValue())

		assert.True(t, c.IncrBy(10))
		assert.Equal(t, int64(15), c.GetValue())
	})

	t.Run("with limit up to capacity", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(20)

		assert.True(t, c.IncrBy(15))
		assert.Equal(t, int64(15), c.GetValue())

		assert.True(t, c.IncrBy(5))
		assert.Equal(t, int64(20), c.GetValue())
	})

	t.Run("fails when exceeding limit and preserves value", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(20)
		c.IncrBy(20)

		assert.False(t, c.IncrBy(1))
		assert.Equal(t, int64(20), c.GetValue())
	})
}

func TestCounter_DecrBy(t *testing.T) {
	t.Run("valid decrement", func(t *testing.T) {
		c := NewCounter()
		c.IncrBy(20)

		assert.True(t, c.DecrBy(5))
		assert.Equal(t, int64(15), c.GetValue())
	})

	t.Run("decrement down to zero", func(t *testing.T) {
		c := NewCounter()
		c.IncrBy(15)

		assert.True(t, c.DecrBy(15))
		assert.Equal(t, int64(0), c.GetValue())
	})

	t.Run("fails when decrementing below zero and preserves value", func(t *testing.T) {
		c := NewCounter()

		assert.False(t, c.DecrBy(1))
		assert.Equal(t, int64(0), c.GetValue())
	})
}

func TestCounter_Reset(t *testing.T) {
	t.Run("resets value to zero", func(t *testing.T) {
		c := NewCounter()
		c.IncrBy(42)
		assert.Equal(t, int64(42), c.GetValue())

		c.Reset()
		assert.Equal(t, int64(0), c.GetValue())
	})
}

func TestCounter_Expiration(t *testing.T) {
	t.Run("initial state without expiration", func(t *testing.T) {
		c := NewCounter()
		assert.False(t, c.IsExpired())
		assert.Equal(t, int64(0), c.TTL())
	})

	t.Run("negative TTL rejected", func(t *testing.T) {
		c := NewCounter()
		assert.False(t, c.Expire(-5))
		assert.False(t, c.IsExpired())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		c := NewCounter()
		assert.True(t, c.Expire(60))
		assert.False(t, c.IsExpired())
		assert.Greater(t, c.TTL(), int64(0))
		assert.LessOrEqual(t, c.TTL(), int64(60))
	})

	t.Run("detects expired state", func(t *testing.T) {
		c := NewCounter()
		c.expiresAt.Store(time.Now().Unix() - 1)
		assert.True(t, c.IsExpired())
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		c := NewCounter()
		c.expiresAt.Store(time.Now().Unix() - 1)
		assert.False(t, c.Expire(30))
	})
}

func TestCounter_Concurrency(t *testing.T) {
	t.Run("concurrent increments and decrements", func(t *testing.T) {
		c := NewCounter()
		const goroutines = 50
		const increments = 100

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()
				for range increments {
					c.IncrBy(1)
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(goroutines*increments), c.GetValue())

		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				for range increments {
					c.DecrBy(1)
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(0), c.GetValue())
	})
}
