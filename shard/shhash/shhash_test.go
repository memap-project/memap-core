package shhash

import (
	"sync"
	"testing"
	"time"

	"github.com/memap-project/memap-core/shard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedHash_NewShardedHash(t *testing.T) {
	shh := NewShardedHash(4)
	require.NotNil(t, shh)
	assert.Equal(t, uint8(4), shh.shardCount)
	assert.Len(t, shh.shards, 4)

	m, ok := shh.Get("any")
	assert.False(t, ok)
	assert.Empty(t, m)
}

func TestShardedHash_SetField_and_GetField(t *testing.T) {
	t.Run("auto-inits hash and sets/gets fields", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("user:1", "name", "Alice")
		shh.SetField("user:1", "email", "alice@example.com")

		val, status := shh.GetField("user:1", "name")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "Alice", val)

		val, status = shh.GetField("user:1", "email")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "alice@example.com", val)
	})

	t.Run("overwrites existing field", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("user:1", "role", "guest")
		shh.SetField("user:1", "role", "admin")

		val, status := shh.GetField("user:1", "role")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "admin", val)
	})

	t.Run("get field on missing hash returns StatusNotFound", func(t *testing.T) {
		shh := NewShardedHash(4)
		val, status := shh.GetField("missing", "field")
		assert.Equal(t, shard.StatusNotFound, status)
		assert.Empty(t, val)
	})

	t.Run("get missing field on existing hash returns StatusFieldNotFound", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "f1", "v1")

		val, status := shh.GetField("h1", "f2")
		assert.Equal(t, shard.StatusFieldNotFound, status)
		assert.Empty(t, val)
	})

	t.Run("get field on expired hash returns StatusExpired", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_hash", "f1", "v1")
		require.True(t, shh.Expire("exp_hash", 1))

		time.Sleep(2100 * time.Millisecond)

		val, status := shh.GetField("exp_hash", "f1")
		assert.Equal(t, shard.StatusExpired, status)
		assert.Empty(t, val)

		assert.False(t, shh.Exists("exp_hash"))
	})

	t.Run("set field on expired hash re-initializes hash", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_hash2", "old", "val")
		require.True(t, shh.Expire("exp_hash2", 1))

		time.Sleep(2100 * time.Millisecond)

		shh.SetField("exp_hash2", "new", "val2")

		val, status := shh.GetField("exp_hash2", "new")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "val2", val)

		_, statusOld := shh.GetField("exp_hash2", "old")
		assert.Equal(t, shard.StatusFieldNotFound, statusOld)
	})
}

func TestShardedHash_Get(t *testing.T) {
	t.Run("retrieves copy of all field-value pairs", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "a", "1")
		shh.SetField("h1", "b", "2")

		m, ok := shh.Get("h1")
		assert.True(t, ok)
		assert.Equal(t, map[string]string{"a": "1", "b": "2"}, m)
	})

	t.Run("get missing hash returns empty map and false", func(t *testing.T) {
		shh := NewShardedHash(4)
		m, ok := shh.Get("missing")
		assert.False(t, ok)
		assert.Empty(t, m)
	})

	t.Run("get expired hash returns empty map and false", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_h", "a", "1")
		require.True(t, shh.Expire("exp_h", 1))

		time.Sleep(2100 * time.Millisecond)

		m, ok := shh.Get("exp_h")
		assert.False(t, ok)
		assert.Empty(t, m)
		assert.False(t, shh.Exists("exp_h"))
	})
}

func TestShardedHash_Set(t *testing.T) {
	t.Run("initializes empty hash with TTL", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.Set("empty_h", 100)

		assert.True(t, shh.Exists("empty_h"))
		assert.Greater(t, shh.TTL("empty_h"), int64(0))

		l, ok := shh.Len("empty_h")
		assert.True(t, ok)
		assert.Equal(t, int64(0), l)
	})

	t.Run("set on expired hash re-initializes hash", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_h", "k", "v")
		require.True(t, shh.Expire("exp_h", 1))

		time.Sleep(2100 * time.Millisecond)

		shh.Set("exp_h", 50)
		assert.True(t, shh.Exists("exp_h"))
		l, ok := shh.Len("exp_h")
		assert.True(t, ok)
		assert.Equal(t, int64(0), l)
	})
}

func TestShardedHash_DeleteField(t *testing.T) {
	t.Run("deletes field from existing hash", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "f1", "v1")
		shh.SetField("h1", "f2", "v2")

		shh.DeleteField("h1", "f1")

		_, status := shh.GetField("h1", "f1")
		assert.Equal(t, shard.StatusFieldNotFound, status)

		val, status := shh.GetField("h1", "f2")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "v2", val)
	})

	t.Run("delete field from non-existent hash is no-op", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.DeleteField("missing", "f1")
		assert.False(t, shh.Exists("missing"))
	})
}

func TestShardedHash_Delete(t *testing.T) {
	t.Run("deletes existing hash", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "f1", "v1")
		shh.Delete("h1")

		assert.False(t, shh.Exists("h1"))
	})

	t.Run("delete missing hash is no-op", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.Delete("missing")
		assert.False(t, shh.Exists("missing"))
	})
}

func TestShardedHash_Exists(t *testing.T) {
	t.Run("missing hash returns false", func(t *testing.T) {
		shh := NewShardedHash(4)
		assert.False(t, shh.Exists("missing"))
	})

	t.Run("existing unexpired hash returns true", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "f1", "v1")
		assert.True(t, shh.Exists("h1"))
	})

	t.Run("expired hash returns false and deletes key", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_h", "f1", "v1")
		require.True(t, shh.Expire("exp_h", 1))

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shh.Exists("exp_h"))
	})
}

func TestShardedHash_Len_Keys_Values(t *testing.T) {
	t.Run("returns len, keys, and values for populated hash", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "k1", "v1")
		shh.SetField("h1", "k2", "v2")

		l, ok := shh.Len("h1")
		assert.True(t, ok)
		assert.Equal(t, int64(2), l)

		keys, ok := shh.Keys("h1")
		assert.True(t, ok)
		assert.ElementsMatch(t, []string{"k1", "k2"}, keys)

		values, ok := shh.Values("h1")
		assert.True(t, ok)
		assert.ElementsMatch(t, []string{"v1", "v2"}, values)
	})

	t.Run("missing hash returns 0/empty and false", func(t *testing.T) {
		shh := NewShardedHash(4)
		l, ok := shh.Len("missing")
		assert.False(t, ok)
		assert.Equal(t, int64(0), l)

		keys, ok := shh.Keys("missing")
		assert.False(t, ok)
		assert.Empty(t, keys)

		values, ok := shh.Values("missing")
		assert.False(t, ok)
		assert.Empty(t, values)
	})

	t.Run("expired hash returns 0/empty and false and deletes key", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_h", "k1", "v1")
		require.True(t, shh.Expire("exp_h", 1))

		time.Sleep(2100 * time.Millisecond)

		l, ok := shh.Len("exp_h")
		assert.False(t, ok)
		assert.Equal(t, int64(0), l)

		// Create another expired hash for Keys/Values checks
		shh.SetField("exp_h2", "k1", "v1")
		require.True(t, shh.Expire("exp_h2", 1))
		time.Sleep(2100 * time.Millisecond)

		keys, ok := shh.Keys("exp_h2")
		assert.False(t, ok)
		assert.Empty(t, keys)

		shh.SetField("exp_h3", "k1", "v1")
		require.True(t, shh.Expire("exp_h3", 1))
		time.Sleep(2100 * time.Millisecond)

		values, ok := shh.Values("exp_h3")
		assert.False(t, ok)
		assert.Empty(t, values)
	})
}

func TestShardedHash_Expire_and_TTL(t *testing.T) {
	t.Run("missing hash TTL returns -2 and Expire returns false", func(t *testing.T) {
		shh := NewShardedHash(4)
		assert.Equal(t, int64(-2), shh.TTL("missing"))
		assert.False(t, shh.Expire("missing", 10))
	})

	t.Run("hash without expiration returns -1", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "f1", "v1")
		assert.Equal(t, int64(-1), shh.TTL("h1"))
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("h1", "f1", "v1")

		assert.True(t, shh.Expire("h1", 100))
		ttl := shh.TTL("h1")
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(100))
	})

	t.Run("expired hash returns -2 and Expire returns false", func(t *testing.T) {
		shh := NewShardedHash(4)
		shh.SetField("exp_h", "f1", "v1")
		require.True(t, shh.Expire("exp_h", 1))

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shh.Expire("exp_h", 10))
		assert.Equal(t, int64(-2), shh.TTL("exp_h"))
	})
}

func TestShardedHash_CleanExpired(t *testing.T) {
	shh := NewShardedHash(4)
	shh.SetField("active", "k", "v")
	shh.SetField("exp1", "k", "v")
	shh.SetField("exp2", "k", "v")

	require.True(t, shh.Expire("exp1", 1))
	require.True(t, shh.Expire("exp2", 1))

	time.Sleep(2100 * time.Millisecond)

	shh.CleanExpired()

	assert.True(t, shh.Exists("active"))
	assert.False(t, shh.Exists("exp1"))
	assert.False(t, shh.Exists("exp2"))
}

func TestShardedHash_Flush(t *testing.T) {
	shh := NewShardedHash(4)
	shh.SetField("h1", "k", "v")
	shh.SetField("h2", "k", "v")

	shh.Flush()

	assert.False(t, shh.Exists("h1"))
	assert.False(t, shh.Exists("h2"))
}

func TestShardedHash_Concurrency(t *testing.T) {
	shh := NewShardedHash(8)
	const goroutines = 20
	const ops = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Writers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				shh.SetField(string(rune('a'+(id%8))), string(rune('0'+(i%10))), "val")
			}
		}(g)
	}

	// Readers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				shh.GetField(string(rune('a'+(id%8))), string(rune('0'+(i%10))))
				shh.Exists(string(rune('a' + (id % 8))))
				shh.Len(string(rune('a' + (id % 8))))
			}
		}(g)
	}

	// Deleters
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				shh.DeleteField(string(rune('a'+(id%8))), string(rune('0'+(i%10))))
			}
		}(g)
	}

	wg.Wait()
}
