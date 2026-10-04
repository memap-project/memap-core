package vals

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash_NewHash(t *testing.T) {
	h := NewHash()
	require.NotNil(t, h)

	assert.Equal(t, int64(0), h.Len())
	assert.False(t, h.IsExpired())
	assert.Equal(t, int64(0), h.TTL())
	assert.Empty(t, h.Keys())
	assert.Empty(t, h.Values())
	assert.Empty(t, h.GetCopy())
}

func TestHash_Set_and_Get(t *testing.T) {
	t.Run("get missing field returns false", func(t *testing.T) {
		h := NewHash()
		val, ok := h.Get("name")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("set and get field", func(t *testing.T) {
		h := NewHash()
		h.Set("name", "Alice")
		val, ok := h.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("overwrite existing field", func(t *testing.T) {
		h := NewHash()
		h.Set("name", "Alice")
		h.Set("name", "Bob")
		val, ok := h.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Bob", val)
	})

	t.Run("set multiple fields", func(t *testing.T) {
		h := NewHash()
		h.Set("name", "Bob")
		h.Set("age", "30")

		val, ok := h.Get("age")
		assert.True(t, ok)
		assert.Equal(t, "30", val)
		assert.Equal(t, int64(2), h.Len())
	})
}

func TestHash_Delete(t *testing.T) {
	t.Run("delete existing field", func(t *testing.T) {
		h := NewHash()
		h.Set("f1", "v1")
		h.Set("f2", "v2")

		h.Delete("f1")
		assert.Equal(t, int64(1), h.Len())

		_, ok := h.Get("f1")
		assert.False(t, ok)

		val, ok := h.Get("f2")
		assert.True(t, ok)
		assert.Equal(t, "v2", val)
	})

	t.Run("delete non-existent field is a no-op", func(t *testing.T) {
		h := NewHash()
		h.Set("f1", "v1")
		h.Delete("missing")
		assert.Equal(t, int64(1), h.Len())
	})
}

func TestHash_Keys_Values_and_GetCopy(t *testing.T) {
	t.Run("keys and values return all fields", func(t *testing.T) {
		h := NewHash()
		h.Set("first", "John")
		h.Set("last", "Doe")
		h.Set("city", "Kyiv")

		assert.ElementsMatch(t, []string{"first", "last", "city"}, h.Keys())
		assert.ElementsMatch(t, []string{"John", "Doe", "Kyiv"}, h.Values())
	})

	t.Run("get copy returns independent map", func(t *testing.T) {
		h := NewHash()
		h.Set("first", "John")
		h.Set("last", "Doe")

		cp := h.GetCopy()
		assert.Equal(t, map[string]string{
			"first": "John",
			"last":  "Doe",
		}, cp)

		cp["extra"] = "value"
		assert.Equal(t, int64(2), h.Len())
		_, ok := h.Get("extra")
		assert.False(t, ok)
	})
}

func TestHash_Expiration(t *testing.T) {
	t.Run("initial state without expiration", func(t *testing.T) {
		h := NewHash()
		assert.False(t, h.IsExpired())
		assert.Equal(t, int64(0), h.TTL())
	})

	t.Run("negative TTL rejected", func(t *testing.T) {
		h := NewHash()
		assert.False(t, h.Expire(-10))
		assert.False(t, h.IsExpired())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		h := NewHash()
		assert.True(t, h.Expire(100))
		assert.False(t, h.IsExpired())
		assert.Greater(t, h.TTL(), int64(0))
		assert.LessOrEqual(t, h.TTL(), int64(100))
	})

	t.Run("detects expired state", func(t *testing.T) {
		h := NewHash()
		h.expiresAt.Store(time.Now().Unix() - 1)
		assert.True(t, h.IsExpired())
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		h := NewHash()
		h.expiresAt.Store(time.Now().Unix() - 1)
		assert.False(t, h.Expire(50))
	})
}

func TestHash_Concurrency(t *testing.T) {
	t.Run("concurrent reads and writes", func(t *testing.T) {
		h := NewHash()
		const goroutines = 30
		const ops = 100

		var wg sync.WaitGroup
		wg.Add(goroutines * 3)

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					h.Set("key", "val")
				}
			}()
		}

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					h.Get("key")
					h.Len()
					h.Keys()
				}
			}()
		}

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					h.GetCopy()
				}
			}()
		}

		wg.Wait()
	})
}
