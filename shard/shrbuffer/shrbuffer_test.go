package shrbuffer

import (
	"sync"
	"testing"
	"time"

	"github.com/memap-project/memap-core/shard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedRingBuffer_NewShardedRingBuffer(t *testing.T) {
	shrb := NewShardedRingBuffer(4)
	require.NotNil(t, shrb)
	assert.Equal(t, uint8(4), shrb.shardCount)
	assert.Len(t, shrb.shards, 4)

	val, status := shrb.Pop("missing")
	assert.Equal(t, shard.StatusNotFound, status)
	assert.Empty(t, val)
}

func TestShardedRingBuffer_Init(t *testing.T) {
	t.Run("initializes new buffer", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		assert.Equal(t, shard.StatusSuccess, shrb.Init("rb1", 5, 0))

		c, ok := shrb.Cap("rb1")
		assert.True(t, ok)
		assert.Equal(t, int64(5), c)

		l, ok := shrb.Len("rb1")
		assert.True(t, ok)
		assert.Equal(t, int64(0), l)
	})

	t.Run("duplicate initialization returns StatusAlreadyExists", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		assert.Equal(t, shard.StatusSuccess, shrb.Init("rb1", 5, 0))
		assert.Equal(t, shard.StatusAlreadyExists, shrb.Init("rb1", 10, 0))

		c, ok := shrb.Cap("rb1")
		assert.True(t, ok)
		assert.Equal(t, int64(5), c)
	})

	t.Run("invalid capacities return StatusInvalidCapacity without panic", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4, 16)
		assert.Equal(t, shard.StatusInvalidCapacity, shrb.Init("neg", -1, 0))
		assert.Equal(t, shard.StatusInvalidCapacity, shrb.Init("zero", 0, 0))
		assert.Equal(t, shard.StatusInvalidCapacity, shrb.Init("too_big", 17, 0))
		assert.Equal(t, shard.StatusSuccess, shrb.Init("exact_max", 16, 0))

		_, ok := shrb.Cap("zero")
		assert.False(t, ok)
	})

	t.Run("initializes buffer with TTL", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		assert.Equal(t, shard.StatusSuccess, shrb.Init("rb_ttl", 5, 100))
		assert.Greater(t, shrb.TTL("rb_ttl"), int64(0))
	})
}

func TestShardedRingBuffer_Push_and_Pop(t *testing.T) {
	t.Run("push and pop in FIFO order", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 3, 0))

		assert.True(t, shrb.Push("rb", "item1"))
		assert.True(t, shrb.Push("rb", "item2"))

		val, status := shrb.Pop("rb")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "item1", val)

		val, status = shrb.Pop("rb")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "item2", val)
	})

	t.Run("pop from empty buffer returns StatusBufferEmpty", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 3, 0))

		val, status := shrb.Pop("rb")
		assert.Equal(t, shard.StatusBufferEmpty, status)
		assert.Empty(t, val)
	})

	t.Run("push and pop on missing buffer return false/StatusNotFound", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		assert.False(t, shrb.Push("missing", "val"))

		val, status := shrb.Pop("missing")
		assert.Equal(t, shard.StatusNotFound, status)
		assert.Empty(t, val)
	})

	t.Run("push and pop on expired buffer return false/StatusExpired and delete key", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb", 3, 1))

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shrb.Push("exp_rb", "val"))
		assert.Equal(t, int64(-2), shrb.TTL("exp_rb"))

		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb2", 3, 1))
		time.Sleep(2100 * time.Millisecond)

		val, status := shrb.Pop("exp_rb2")
		assert.Equal(t, shard.StatusExpired, status)
		assert.Empty(t, val)
		assert.Equal(t, int64(-2), shrb.TTL("exp_rb2"))
	})
}

func TestShardedRingBuffer_At_and_Slice(t *testing.T) {
	t.Run("access by index and slice", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 3, 0))
		shrb.Push("rb", "a")
		shrb.Push("rb", "b")

		val, status := shrb.At("rb", 0)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "a", val)

		val, status = shrb.At("rb", 1)
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "b", val)

		val, status = shrb.At("rb", 2)
		assert.Equal(t, shard.StatusIndexOutOfBounds, status)
		assert.Empty(t, val)

		s, ok := shrb.Slice("rb")
		assert.True(t, ok)
		assert.Equal(t, []string{"a", "b"}, s)
	})

	t.Run("at and slice on missing buffer", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		val, status := shrb.At("missing", 0)
		assert.Equal(t, shard.StatusNotFound, status)
		assert.Empty(t, val)

		s, ok := shrb.Slice("missing")
		assert.False(t, ok)
		assert.Empty(t, s)
	})

	t.Run("at and slice on expired buffer return StatusExpired/false and delete key", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb", 3, 1))
		shrb.Push("exp_rb", "a")

		time.Sleep(2100 * time.Millisecond)

		val, status := shrb.At("exp_rb", 0)
		assert.Equal(t, shard.StatusExpired, status)
		assert.Empty(t, val)

		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb2", 3, 1))
		time.Sleep(2100 * time.Millisecond)

		s, ok := shrb.Slice("exp_rb2")
		assert.False(t, ok)
		assert.Empty(t, s)
	})
}

func TestShardedRingBuffer_Peek_and_Back(t *testing.T) {
	t.Run("peek and back return head and tail without removing", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 5, 0))
		shrb.Push("rb", "first")
		shrb.Push("rb", "second")
		shrb.Push("rb", "third")

		peekVal, status := shrb.Peek("rb")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "first", peekVal)

		backVal, status := shrb.Back("rb")
		assert.Equal(t, shard.StatusSuccess, status)
		assert.Equal(t, "third", backVal)

		l, ok := shrb.Len("rb")
		assert.True(t, ok)
		assert.Equal(t, int64(3), l)
	})

	t.Run("peek and back on empty buffer return StatusBufferEmpty", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 5, 0))

		val, status := shrb.Peek("rb")
		assert.Equal(t, shard.StatusBufferEmpty, status)
		assert.Empty(t, val)

		val, status = shrb.Back("rb")
		assert.Equal(t, shard.StatusBufferEmpty, status)
		assert.Empty(t, val)
	})

	t.Run("peek and back on missing buffer return StatusNotFound", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)

		val, status := shrb.Peek("missing")
		assert.Equal(t, shard.StatusNotFound, status)
		assert.Empty(t, val)

		val, status = shrb.Back("missing")
		assert.Equal(t, shard.StatusNotFound, status)
		assert.Empty(t, val)
	})

	t.Run("peek and back on expired buffer return StatusExpired", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb", 5, 1))
		shrb.Push("exp_rb", "a")

		time.Sleep(2100 * time.Millisecond)

		val, status := shrb.Peek("exp_rb")
		assert.Equal(t, shard.StatusExpired, status)
		assert.Empty(t, val)

		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb2", 5, 1))
		shrb.Push("exp_rb2", "a")
		time.Sleep(2100 * time.Millisecond)

		val, status = shrb.Back("exp_rb2")
		assert.Equal(t, shard.StatusExpired, status)
		assert.Empty(t, val)
	})
}

func TestShardedRingBuffer_Cap_Len_Reset(t *testing.T) {
	t.Run("capacity, length, and reset on valid buffer", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 5, 0))
		shrb.Push("rb", "1")
		shrb.Push("rb", "2")

		c, ok := shrb.Cap("rb")
		assert.True(t, ok)
		assert.Equal(t, int64(5), c)

		l, ok := shrb.Len("rb")
		assert.True(t, ok)
		assert.Equal(t, int64(2), l)

		assert.True(t, shrb.Reset("rb"))

		l, ok = shrb.Len("rb")
		assert.True(t, ok)
		assert.Equal(t, int64(0), l)
	})

	t.Run("cap, len, and reset on missing buffer", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)

		c, ok := shrb.Cap("missing")
		assert.False(t, ok)
		assert.Equal(t, int64(0), c)

		l, ok := shrb.Len("missing")
		assert.False(t, ok)
		assert.Equal(t, int64(0), l)

		assert.False(t, shrb.Reset("missing"))
	})

	t.Run("cap, len, and reset on expired buffer", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb", 5, 1))

		time.Sleep(2100 * time.Millisecond)

		c, ok := shrb.Cap("exp_rb")
		assert.False(t, ok)
		assert.Equal(t, int64(0), c)

		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb2", 5, 1))
		time.Sleep(2100 * time.Millisecond)

		l, ok := shrb.Len("exp_rb2")
		assert.False(t, ok)
		assert.Equal(t, int64(0), l)

		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb3", 5, 1))
		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shrb.Reset("exp_rb3"))
	})
}

func TestShardedRingBuffer_Delete(t *testing.T) {
	t.Run("deletes existing buffer", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 5, 0))

		shrb.Delete("rb")

		_, ok := shrb.Cap("rb")
		assert.False(t, ok)
	})

	t.Run("delete missing buffer is no-op", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		shrb.Delete("missing")
		_, ok := shrb.Cap("missing")
		assert.False(t, ok)
	})
}

func TestShardedRingBuffer_Expire_and_TTL(t *testing.T) {
	t.Run("missing buffer TTL is -2 and Expire returns false", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		assert.Equal(t, int64(-2), shrb.TTL("missing"))
		assert.False(t, shrb.Expire("missing", 10))
	})

	t.Run("buffer without expiration returns -1", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 5, 0))
		assert.Equal(t, int64(-1), shrb.TTL("rb"))
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("rb", 5, 0))

		assert.True(t, shrb.Expire("rb", 100))
		ttl := shrb.TTL("rb")
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(100))
	})

	t.Run("expired buffer returns TTL -2 and Expire returns false", func(t *testing.T) {
		shrb := NewShardedRingBuffer(4)
		require.Equal(t, shard.StatusSuccess, shrb.Init("exp_rb", 5, 1))

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shrb.Expire("exp_rb", 10))
		assert.Equal(t, int64(-2), shrb.TTL("exp_rb"))
	})
}

func TestShardedRingBuffer_CleanExpired(t *testing.T) {
	shrb := NewShardedRingBuffer(4)
	require.Equal(t, shard.StatusSuccess, shrb.Init("active", 5, 0))
	require.Equal(t, shard.StatusSuccess, shrb.Init("exp1", 5, 1))
	require.Equal(t, shard.StatusSuccess, shrb.Init("exp2", 5, 1))

	time.Sleep(2100 * time.Millisecond)

	shrb.CleanExpired()

	_, okActive := shrb.Cap("active")
	assert.True(t, okActive)

	assert.Equal(t, int64(-2), shrb.TTL("exp1"))
	assert.Equal(t, int64(-2), shrb.TTL("exp2"))
}

func TestShardedRingBuffer_Flush(t *testing.T) {
	shrb := NewShardedRingBuffer(4)
	require.Equal(t, shard.StatusSuccess, shrb.Init("rb1", 5, 0))
	require.Equal(t, shard.StatusSuccess, shrb.Init("rb2", 5, 0))

	shrb.Flush()

	_, ok1 := shrb.Cap("rb1")
	assert.False(t, ok1)
	_, ok2 := shrb.Cap("rb2")
	assert.False(t, ok2)
}

func TestShardedRingBuffer_Concurrency(t *testing.T) {
	shrb := NewShardedRingBuffer(8)
	const goroutines = 20
	const ops = 50

	for g := range 8 {
		shrb.Init(string(rune('a'+g)), 100, 0)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Pushers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shrb.Push(string(rune('a'+(id%8))), "val")
			}
		}(g)
	}

	// Poppers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				shrb.Pop(string(rune('a' + (id % 8))))
			}
		}(g)
	}

	// Readers (Peek, Back, Slice, Cap, Len)
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range ops {
				key := string(rune('a' + (id % 8)))
				shrb.Peek(key)
				shrb.Back(key)
				shrb.Slice(key)
				shrb.Len(key)
				shrb.Cap(key)
			}
		}(g)
	}

	wg.Wait()
}
