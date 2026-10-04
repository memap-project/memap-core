package vals

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRingBuffer_NewRingBuffer(t *testing.T) {
	t.Run("initializes empty buffer with given capacity", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		require.NotNil(t, rb)

		assert.Equal(t, int64(5), rb.Cap())
		assert.Equal(t, int64(0), rb.Len())
		assert.True(t, rb.IsEmpty())
		assert.False(t, rb.IsFull())
		assert.False(t, rb.IsExpired())
		assert.Equal(t, int64(0), rb.TTL())
	})

	t.Run("supports int64 generic type", func(t *testing.T) {
		rbInt := NewRingBuffer[int64](10)
		require.NotNil(t, rbInt)
		assert.Equal(t, int64(10), rbInt.Cap())
		assert.True(t, rbInt.IsEmpty())
	})
}

func TestRingBuffer_Push_and_Pop(t *testing.T) {
	t.Run("pop from empty buffer returns false", func(t *testing.T) {
		rb := NewRingBuffer[string](3)
		val, ok := rb.Pop()
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("push and pop in FIFO order", func(t *testing.T) {
		rb := NewRingBuffer[string](3)

		rb.Push("item1")
		assert.Equal(t, int64(1), rb.Len())
		assert.False(t, rb.IsEmpty())

		rb.Push("item2")
		rb.Push("item3")
		assert.Equal(t, int64(3), rb.Len())
		assert.True(t, rb.IsFull())

		val, ok := rb.Pop()
		assert.True(t, ok)
		assert.Equal(t, "item1", val)

		val, ok = rb.Pop()
		assert.True(t, ok)
		assert.Equal(t, "item2", val)

		val, ok = rb.Pop()
		assert.True(t, ok)
		assert.Equal(t, "item3", val)
		assert.True(t, rb.IsEmpty())
	})
}

func TestRingBuffer_Overflow_Overwrite(t *testing.T) {
	t.Run("overwrites oldest element when buffer is full", func(t *testing.T) {
		rb := NewRingBuffer[string](3)

		rb.Push("a")
		rb.Push("b")
		rb.Push("c")
		assert.True(t, rb.IsFull())

		// 4th item overwrites "a"
		rb.Push("d")
		assert.Equal(t, int64(3), rb.Len())
		assert.True(t, rb.IsFull())

		val, ok := rb.Pop()
		assert.True(t, ok)
		assert.Equal(t, "b", val)

		val, ok = rb.Pop()
		assert.True(t, ok)
		assert.Equal(t, "c", val)

		val, ok = rb.Pop()
		assert.True(t, ok)
		assert.Equal(t, "d", val)
	})
}

func TestRingBuffer_At(t *testing.T) {
	t.Run("out of bounds returns false", func(t *testing.T) {
		rb := NewRingBuffer[string](3)
		_, ok := rb.At(0)
		assert.False(t, ok)

		rb.Push("v0")
		_, ok = rb.At(-1)
		assert.False(t, ok)
		_, ok = rb.At(1)
		assert.False(t, ok)
	})

	t.Run("access elements by index before and after wrap-around", func(t *testing.T) {
		rb := NewRingBuffer[string](3)
		rb.Push("v0")
		rb.Push("v1")

		val, ok := rb.At(0)
		assert.True(t, ok)
		assert.Equal(t, "v0", val)

		val, ok = rb.At(1)
		assert.True(t, ok)
		assert.Equal(t, "v1", val)

		// Wrap-around
		rb.Push("v2")
		rb.Push("v3") // overwrites v0, logical order: [v1, v2, v3]

		val, ok = rb.At(0)
		assert.True(t, ok)
		assert.Equal(t, "v1", val)

		val, ok = rb.At(1)
		assert.True(t, ok)
		assert.Equal(t, "v2", val)

		val, ok = rb.At(2)
		assert.True(t, ok)
		assert.Equal(t, "v3", val)
	})
}

func TestRingBuffer_Slice(t *testing.T) {
	t.Run("returns elements in logical order", func(t *testing.T) {
		rb := NewRingBuffer[string](4)
		assert.Empty(t, rb.Slice())

		rb.Push("1")
		rb.Push("2")
		assert.Equal(t, []string{"1", "2"}, rb.Slice())

		rb.Push("3")
		rb.Push("4")
		assert.Equal(t, []string{"1", "2", "3", "4"}, rb.Slice())

		// Wrap-around: overwrites "1" with "5", and "2" with "6"
		rb.Push("5")
		rb.Push("6")
		assert.Equal(t, []string{"3", "4", "5", "6"}, rb.Slice())
	})
}

func TestRingBuffer_Peek_and_Back(t *testing.T) {
	t.Run("returns false on empty buffer", func(t *testing.T) {
		rb := NewRingBuffer[string](3)
		_, ok := rb.Peek()
		assert.False(t, ok)
		_, ok = rb.Back()
		assert.False(t, ok)
	})

	t.Run("returns oldest and newest elements without removal", func(t *testing.T) {
		rb := NewRingBuffer[string](2)

		rb.Push("first")
		peekVal, ok := rb.Peek()
		assert.True(t, ok)
		assert.Equal(t, "first", peekVal)

		backVal, ok := rb.Back()
		assert.True(t, ok)
		assert.Equal(t, "first", backVal)

		rb.Push("second")
		peekVal, ok = rb.Peek()
		assert.True(t, ok)
		assert.Equal(t, "first", peekVal)

		backVal, ok = rb.Back()
		assert.True(t, ok)
		assert.Equal(t, "second", backVal)

		// Overwrite moves head
		rb.Push("third") // buffer full (cap=2), overwrites "first", head moves to "second"
		peekVal, ok = rb.Peek()
		assert.True(t, ok)
		assert.Equal(t, "second", peekVal)

		backVal, ok = rb.Back()
		assert.True(t, ok)
		assert.Equal(t, "third", backVal)
	})
}

func TestRingBuffer_Reset(t *testing.T) {
	t.Run("resets buffer to empty state", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		rb.Push("a")
		rb.Push("b")
		assert.Equal(t, int64(2), rb.Len())

		rb.Reset()
		assert.Equal(t, int64(0), rb.Len())
		assert.True(t, rb.IsEmpty())
		assert.False(t, rb.IsFull())

		_, ok := rb.Pop()
		assert.False(t, ok)
	})
}

func TestRingBuffer_Expiration(t *testing.T) {
	t.Run("initial state without expiration", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		assert.False(t, rb.IsExpired())
		assert.Equal(t, int64(0), rb.TTL())
	})

	t.Run("negative TTL rejected", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		assert.False(t, rb.Expire(-1))
		assert.False(t, rb.IsExpired())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		assert.True(t, rb.Expire(300))
		assert.False(t, rb.IsExpired())
		assert.Greater(t, rb.TTL(), int64(0))
		assert.LessOrEqual(t, rb.TTL(), int64(300))
	})

	t.Run("detects expired state", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		rb.expiresAt.Store(time.Now().Unix() - 1)
		assert.True(t, rb.IsExpired())
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		rb := NewRingBuffer[string](5)
		rb.expiresAt.Store(time.Now().Unix() - 1)
		assert.False(t, rb.Expire(60))
	})
}

func TestRingBuffer_Concurrency(t *testing.T) {
	t.Run("concurrent push and pop operations", func(t *testing.T) {
		rb := NewRingBuffer[int64](100)
		const goroutines = 20
		const ops = 100

		var wg sync.WaitGroup
		wg.Add(goroutines * 2)

		for i := range goroutines {
			go func(id int) {
				defer wg.Done()
				for j := range ops {
					rb.Push(int64(id*ops + j))
				}
			}(i)
		}

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					rb.Pop()
					rb.Peek()
					rb.Back()
					rb.Slice()
				}
			}()
		}

		wg.Wait()
	})
}
