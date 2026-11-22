package shcounter

import (
	"sync"
	"testing"
	"time"

	"github.com/memap-project/memap-core/shard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedCounter_NewShardedCounter(t *testing.T) {
	shc := NewShardedCounter(4)
	require.NotNil(t, shc)
	assert.Equal(t, uint8(4), shc.shardCount)
	assert.Len(t, shc.shards, 4)

	val, ok := shc.Get("any")
	assert.False(t, ok)
	assert.Equal(t, int64(0), val)
}

func TestShardedCounter_IncrBy(t *testing.T) {
	t.Run("auto-inits and increments", func(t *testing.T) {
		shc := NewShardedCounter(4)

		val, status := shc.IncrBy("cnt", 5)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, int64(5), val)

		val, status = shc.IncrBy("cnt", 10)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, int64(15), val)

		got, ok := shc.Get("cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(15), got)
	})

	t.Run("exceeds limit returns StatusLimitExceeded", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.SetLimit("cnt", 20)

		val, status := shc.IncrBy("cnt", 15)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, int64(15), val)

		val, status = shc.IncrBy("cnt", 10)
		assert.Equal(t, shard.StatusLimitExceeded, status)
		assert.Equal(t, int64(0), val)

		got, ok := shc.Get("cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(15), got)
	})

	t.Run("re-inits expired counter deterministically with fresh limit", func(t *testing.T) {
		shc := NewShardedCounter(4)
		require.Equal(t, shard.StatusSuccess, shc.SetLimit("exp_cnt", 50))
		shc.IncrBy("exp_cnt", 20)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt", 1))

		time.Sleep(1100 * time.Millisecond)

		val, status := shc.IncrBy("exp_cnt", 10)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, int64(10), val)

		// Deterministic behavior: expired counter behaves identically to a deleted counter
		lim, ok := shc.GetLimit("exp_cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(0), lim)
	})

	t.Run("rejects negative alpha and negative limit without creating counter", func(t *testing.T) {
		shc := NewShardedCounter(4)
		assert.Equal(t, shard.StatusInvalidArg, shc.SetLimit("neg_lim", -5))
		_, exists := shc.Get("neg_lim")
		assert.False(t, exists)

		_, status := shc.IncrBy("neg_alpha", -1)
		assert.Equal(t, shard.StatusInvalidArg, status)
		_, exists = shc.Get("neg_alpha")
		assert.False(t, exists)

		_, status = shc.DecrBy("neg_alpha", -1)
		assert.Equal(t, shard.StatusInvalidArg, status)
	})
}

func TestShardedCounter_DecrBy(t *testing.T) {
	t.Run("decrements existing counter", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("cnt", 20)

		val, status := shc.DecrBy("cnt", 5)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, int64(15), val)

		val, status = shc.DecrBy("cnt", 15)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, int64(0), val)
	})

	t.Run("returns StatusNotFound for non-existent counter", func(t *testing.T) {
		shc := NewShardedCounter(4)

		val, status := shc.DecrBy("missing", 5)
		assert.Equal(t, shard.StatusNotFound, status)
		assert.Equal(t, int64(0), val)
	})

	t.Run("returns StatusLimitExceeded when decrement would be negative", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("cnt", 10)

		val, status := shc.DecrBy("cnt", 15)
		assert.Equal(t, shard.StatusLimitExceeded, status)
		assert.Equal(t, int64(0), val)

		got, ok := shc.Get("cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(10), got)
	})

	t.Run("returns StatusExpired for expired counter", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("exp_cnt", 10)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt", 1))

		time.Sleep(2100 * time.Millisecond)

		val, status := shc.DecrBy("exp_cnt", 5)
		assert.Equal(t, shard.StatusExpired, status)
		assert.Equal(t, int64(0), val)

		_, ok := shc.Get("exp_cnt")
		assert.False(t, ok)
	})
}

func TestShardedCounter_Get_and_SetLimit_and_GetLimit(t *testing.T) {
	t.Run("get on missing counter returns 0 and false", func(t *testing.T) {
		shc := NewShardedCounter(4)
		val, ok := shc.Get("missing")
		assert.False(t, ok)
		assert.Equal(t, int64(0), val)

		lim, ok := shc.GetLimit("missing")
		assert.False(t, ok)
		assert.Equal(t, int64(0), lim)
	})

	t.Run("set and get limit on fresh counter", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.SetLimit("cnt", 100)

		lim, ok := shc.GetLimit("cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(100), lim)

		val, ok := shc.Get("cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(0), val)
	})

	t.Run("get and get limit on expired counter returns false and cleans key", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.SetLimit("exp_cnt", 50)
		shc.IncrBy("exp_cnt", 25)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt", 1))

		time.Sleep(2100 * time.Millisecond)

		val, ok := shc.Get("exp_cnt")
		assert.False(t, ok)
		assert.Equal(t, int64(0), val)

		// Set again, expire again, and test GetLimit
		shc.SetLimit("exp_cnt2", 50)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt2", 1))
		time.Sleep(2100 * time.Millisecond)

		lim, ok := shc.GetLimit("exp_cnt2")
		assert.False(t, ok)
		assert.Equal(t, int64(0), lim)
	})

	t.Run("set limit on expired counter re-initializes counter", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("exp_cnt3", 25)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt3", 1))

		time.Sleep(2100 * time.Millisecond)

		shc.SetLimit("exp_cnt3", 200)

		lim, ok := shc.GetLimit("exp_cnt3")
		assert.True(t, ok)
		assert.Equal(t, int64(200), lim)

		val, ok := shc.Get("exp_cnt3")
		assert.True(t, ok)
		assert.Equal(t, int64(0), val)
	})
}

func TestShardedCounter_Reset(t *testing.T) {
	t.Run("resets existing counter to 0", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("cnt", 42)

		assert.True(t, shc.Reset("cnt"))

		val, ok := shc.Get("cnt")
		assert.True(t, ok)
		assert.Equal(t, int64(0), val)
	})

	t.Run("reset on missing counter returns false", func(t *testing.T) {
		shc := NewShardedCounter(4)
		assert.False(t, shc.Reset("missing"))
	})

	t.Run("reset on expired counter returns false", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("exp_cnt", 10)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt", 1))

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shc.Reset("exp_cnt"))
	})
}

func TestShardedCounter_Delete(t *testing.T) {
	t.Run("deletes existing counter", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("cnt", 10)

		shc.Delete("cnt")

		_, ok := shc.Get("cnt")
		assert.False(t, ok)
	})

	t.Run("delete non-existent counter is a no-op", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.Delete("missing")
		_, ok := shc.Get("missing")
		assert.False(t, ok)
	})
}

func TestShardedCounter_Expire_and_TTL(t *testing.T) {
	t.Run("missing counter TTL returns -2 and expire returns StatusNotFound", func(t *testing.T) {
		shc := NewShardedCounter(4)
		assert.Equal(t, int64(-2), shc.TTL("missing"))
		assert.Equal(t, shard.StatusNotFound, shc.Expire("missing", 10))
	})

	t.Run("counter without TTL returns -1", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("cnt", 1)
		assert.Equal(t, int64(-1), shc.TTL("cnt"))
	})

	t.Run("sets valid TTL and rejects invalid TTL", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("cnt", 1)

		assert.Equal(t, shard.StatusInvalidTTL, shc.Expire("cnt", 0))
		assert.Equal(t, shard.StatusInvalidTTL, shc.Expire("cnt", -1))
		assert.Equal(t, shard.StatusSuccess, shc.Expire("cnt", 100))
		ttl := shc.TTL("cnt")
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(100))
	})

	t.Run("expired counter returns TTL -2 and expire returns StatusExpired", func(t *testing.T) {
		shc := NewShardedCounter(4)
		shc.IncrBy("exp_cnt", 1)
		require.Equal(t, shard.StatusSuccess, shc.Expire("exp_cnt", 1))

		time.Sleep(1100 * time.Millisecond)

		assert.Equal(t, shard.StatusExpired, shc.Expire("exp_cnt", 10))
		assert.Equal(t, int64(-2), shc.TTL("exp_cnt"))
	})
}

func TestShardedCounter_CleanExpired(t *testing.T) {
	shc := NewShardedCounter(4)
	shc.IncrBy("active", 1)
	shc.IncrBy("exp1", 1)
	shc.IncrBy("exp2", 1)

	require.Equal(t, shard.StatusSuccess, shc.Expire("exp1", 1))
	require.Equal(t, shard.StatusSuccess, shc.Expire("exp2", 1))

	time.Sleep(2100 * time.Millisecond)

	shc.CleanExpired()

	_, okActive := shc.Get("active")
	assert.True(t, okActive)

	assert.Equal(t, int64(-2), shc.TTL("exp1"))
	assert.Equal(t, int64(-2), shc.TTL("exp2"))
}

func TestShardedCounter_Flush(t *testing.T) {
	shc := NewShardedCounter(4)
	shc.IncrBy("cnt1", 10)
	shc.IncrBy("cnt2", 20)

	shc.Flush()

	_, ok1 := shc.Get("cnt1")
	assert.False(t, ok1)
	_, ok2 := shc.Get("cnt2")
	assert.False(t, ok2)
}

func TestShardedCounter_Concurrency(t *testing.T) {
	shc := NewShardedCounter(8)
	const goroutines = 20
	const ops = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Incrementers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shc.IncrBy(string(rune('a'+(id%8))), 1)
			}
		}(g)
	}

	// Readers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shc.Get(string(rune('a' + (id % 8))))
				shc.TTL(string(rune('a' + (id % 8))))
			}
		}(g)
	}

	// Decrementers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shc.DecrBy(string(rune('a'+(id%8))), 1)
			}
		}(g)
	}

	wg.Wait()
}
