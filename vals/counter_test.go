package vals

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounter_NewCounter(t *testing.T) {
	c := NewCounter()
	require.NotNil(t, c)

	assert.Equal(t, int64(0), c.GetValue())
	assert.Equal(t, int64(0), c.GetLimit())
	assert.Equal(t, int64(-1), c.TTL())
	assert.False(t, c.IsExpired())
}

func TestCounter_Limit(t *testing.T) {
	t.Run("set limit", func(t *testing.T) {
		c := NewCounter()
		assert.True(t, c.SetLimit(100))
		assert.Equal(t, int64(100), c.GetLimit())
	})

	t.Run("update limit", func(t *testing.T) {
		c := NewCounter()
		assert.True(t, c.SetLimit(100))
		assert.True(t, c.SetLimit(250))
		assert.Equal(t, int64(250), c.GetLimit())
	})

	t.Run("rejects negative limit", func(t *testing.T) {
		c := NewCounter()
		assert.False(t, c.SetLimit(-1))
		assert.Equal(t, int64(0), c.GetLimit())
	})
}

func TestCounter_IncrBy(t *testing.T) {
	t.Run("without limit", func(t *testing.T) {
		c := NewCounter()

		v, ok := c.IncrBy(5)
		assert.True(t, ok)
		assert.Equal(t, int64(5), v)
		assert.Equal(t, int64(5), c.GetValue())

		v, ok = c.IncrBy(10)
		assert.True(t, ok)
		assert.Equal(t, int64(15), v)
		assert.Equal(t, int64(15), c.GetValue())
	})

	t.Run("with limit up to capacity", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(20)

		v, ok := c.IncrBy(15)
		assert.True(t, ok)
		assert.Equal(t, int64(15), v)

		v, ok = c.IncrBy(5)
		assert.True(t, ok)
		assert.Equal(t, int64(20), v)
	})

	t.Run("fails when exceeding limit and preserves value", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(20)
		_, ok := c.IncrBy(20)
		require.True(t, ok)

		v, ok := c.IncrBy(1)
		assert.False(t, ok)
		assert.Equal(t, int64(0), v)
		assert.Equal(t, int64(20), c.GetValue())
	})

	t.Run("prevents signed int64 overflow and rejects negative alpha", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(100)
		_, ok := c.IncrBy(50)
		require.True(t, ok)

		_, ok = c.IncrBy(math.MaxInt64)
		assert.False(t, ok)
		assert.Equal(t, int64(50), c.GetValue())

		_, ok = c.IncrBy(-10)
		assert.False(t, ok)
		assert.Equal(t, int64(50), c.GetValue())

		// Without limit, int64 overflow must also be rejected
		cUnlimited := NewCounter()
		_, ok = cUnlimited.IncrBy(math.MaxInt64 - 1)
		require.True(t, ok)
		_, ok = cUnlimited.IncrBy(2)
		assert.False(t, ok)
		assert.Equal(t, int64(math.MaxInt64-1), cUnlimited.GetValue())
	})
}

func TestCounter_DecrBy(t *testing.T) {
	t.Run("valid decrement", func(t *testing.T) {
		c := NewCounter()
		c.IncrBy(20)

		v, ok := c.DecrBy(5)
		assert.True(t, ok)
		assert.Equal(t, int64(15), v)
		assert.Equal(t, int64(15), c.GetValue())
	})

	t.Run("decrement down to zero", func(t *testing.T) {
		c := NewCounter()
		c.IncrBy(15)

		v, ok := c.DecrBy(15)
		assert.True(t, ok)
		assert.Equal(t, int64(0), v)
		assert.Equal(t, int64(0), c.GetValue())
	})

	t.Run("fails when decrementing below zero or negative alpha and preserves value", func(t *testing.T) {
		c := NewCounter()
		c.SetLimit(10)
		c.IncrBy(5)

		_, ok := c.DecrBy(6)
		assert.False(t, ok)
		assert.Equal(t, int64(5), c.GetValue())

		// Negative alpha must not bypass limit by increasing value
		_, ok = c.DecrBy(-100)
		assert.False(t, ok)
		assert.Equal(t, int64(5), c.GetValue())
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
		assert.Equal(t, int64(-1), c.TTL())
	})

	t.Run("non-positive TTL rejected", func(t *testing.T) {
		c := NewCounter()
		assert.False(t, c.Expire(-5))
		assert.False(t, c.Expire(0))
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
		c.expiresAt.Store(MonotonicNow() - 1)
		assert.True(t, c.IsExpired())
		assert.Equal(t, int64(-2), c.TTL())
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		c := NewCounter()
		c.expiresAt.Store(MonotonicNow() - 1)
		assert.False(t, c.Expire(30))
	})
}

func TestCounter_Concurrency(t *testing.T) {
	t.Run("concurrent increments return unique sequential values", func(t *testing.T) {
		c := NewCounter()
		const goroutines = 50
		const increments = 100
		const total = goroutines * increments

		seen := make([]bool, total+1)
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()
				for range increments {
					v, ok := c.IncrBy(1)
					assert.True(t, ok)
					mu.Lock()
					assert.False(t, seen[v], "duplicate returned value %d", v)
					seen[v] = true
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(total), c.GetValue())

		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				for range increments {
					_, ok := c.DecrBy(1)
					assert.True(t, ok)
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(0), c.GetValue())
	})
}
