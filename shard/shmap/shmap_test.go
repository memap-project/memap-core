package shmap

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedMap_NewShardedMap(t *testing.T) {
	shm := NewShardedMap(4)
	require.NotNil(t, shm)
	assert.Equal(t, uint8(4), shm.shardCount)
	assert.Len(t, shm.shards, 4)

	val, ok := shm.Get("missing")
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestShardedMap_Set_and_Get(t *testing.T) {
	t.Run("set and get without TTL", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("k1", "v1", 0)

		val, ok := shm.Get("k1")
		assert.True(t, ok)
		assert.Equal(t, "v1", val)
	})

	t.Run("set with TTL", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("k2", "v2", 100)

		val, ok := shm.Get("k2")
		assert.True(t, ok)
		assert.Equal(t, "v2", val)
		assert.Greater(t, shm.TTL("k2"), int64(0))
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("k", "old", 0)
		shm.Set("k", "new", 0)

		val, ok := shm.Get("k")
		assert.True(t, ok)
		assert.Equal(t, "new", val)
	})

	t.Run("get missing key returns empty string and false", func(t *testing.T) {
		shm := NewShardedMap(4)
		val, ok := shm.Get("missing")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("get expired key returns empty string and false and deletes key", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("exp_k", "v", 1)

		time.Sleep(2100 * time.Millisecond)

		val, ok := shm.Get("exp_k")
		assert.False(t, ok)
		assert.Empty(t, val)

		// Check TTL is -2 (missing)
		assert.Equal(t, int64(-2), shm.TTL("exp_k"))
	})
}

func TestShardedMap_Delete(t *testing.T) {
	t.Run("deletes existing key", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("k", "v", 0)
		shm.Delete("k")

		_, ok := shm.Get("k")
		assert.False(t, ok)
	})

	t.Run("delete missing key is no-op", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Delete("missing")
		_, ok := shm.Get("missing")
		assert.False(t, ok)
	})
}

func TestShardedMap_Expire_and_TTL(t *testing.T) {
	t.Run("missing key TTL is -2 and Expire returns false", func(t *testing.T) {
		shm := NewShardedMap(4)
		assert.Equal(t, int64(-2), shm.TTL("missing"))
		assert.False(t, shm.Expire("missing", 10))
	})

	t.Run("key without TTL returns -1", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("k", "v", 0)
		assert.Equal(t, int64(-1), shm.TTL("k"))
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("k", "v", 0)

		assert.True(t, shm.Expire("k", 100))
		ttl := shm.TTL("k")
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(100))
	})

	t.Run("expired key returns TTL -2 and Expire returns false", func(t *testing.T) {
		shm := NewShardedMap(4)
		shm.Set("exp_k", "v", 1)

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shm.Expire("exp_k", 10))
		assert.Equal(t, int64(-2), shm.TTL("exp_k"))
	})
}

func TestShardedMap_CleanExpired(t *testing.T) {
	shm := NewShardedMap(4)
	shm.Set("active", "v", 0)
	shm.Set("exp1", "v", 1)
	shm.Set("exp2", "v", 1)

	time.Sleep(2100 * time.Millisecond)

	shm.CleanExpired()

	_, okActive := shm.Get("active")
	assert.True(t, okActive)

	assert.Equal(t, int64(-2), shm.TTL("exp1"))
	assert.Equal(t, int64(-2), shm.TTL("exp2"))
}

func TestShardedMap_Flush(t *testing.T) {
	shm := NewShardedMap(4)
	shm.Set("k1", "v1", 0)
	shm.Set("k2", "v2", 0)

	shm.Flush()

	_, ok1 := shm.Get("k1")
	assert.False(t, ok1)
	_, ok2 := shm.Get("k2")
	assert.False(t, ok2)
}

func TestShardedMap_Concurrency(t *testing.T) {
	shm := NewShardedMap(8)
	const goroutines = 20
	const ops = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Writers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shm.Set(string(rune('a'+(id%8))), "val", 0)
			}
		}(g)
	}

	// Readers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shm.Get(string(rune('a' + (id % 8))))
				shm.TTL(string(rune('a' + (id % 8))))
			}
		}(g)
	}

	// Deleters
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shm.Delete(string(rune('a' + (id % 8))))
			}
		}(g)
	}

	wg.Wait()
}
