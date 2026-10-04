package vals

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_NewSet(t *testing.T) {
	s := NewSet()
	require.NotNil(t, s)

	assert.Equal(t, int64(0), s.Card())
	assert.Empty(t, s.Members())
	assert.False(t, s.IsExpired())
	assert.Equal(t, int64(0), s.TTL())
}

func TestSet_Add_and_IsMember(t *testing.T) {
	t.Run("membership on empty set returns false", func(t *testing.T) {
		s := NewSet()
		assert.False(t, s.IsMember("item1"))
	})

	t.Run("add member and check membership", func(t *testing.T) {
		s := NewSet()
		s.Add("item1")
		assert.True(t, s.IsMember("item1"))
		assert.Equal(t, int64(1), s.Card())

		s.Add("item2")
		assert.True(t, s.IsMember("item2"))
		assert.Equal(t, int64(2), s.Card())
	})

	t.Run("duplicate member does not increase cardinality", func(t *testing.T) {
		s := NewSet()
		s.Add("item1")
		s.Add("item1")
		assert.True(t, s.IsMember("item1"))
		assert.Equal(t, int64(1), s.Card())
	})
}

func TestSet_Remove(t *testing.T) {
	t.Run("remove existing member", func(t *testing.T) {
		s := NewSet()
		s.Add("a")
		s.Add("b")

		s.Remove("a")
		assert.False(t, s.IsMember("a"))
		assert.True(t, s.IsMember("b"))
		assert.Equal(t, int64(1), s.Card())
	})

	t.Run("remove non-existent member is a no-op", func(t *testing.T) {
		s := NewSet()
		s.Add("b")
		s.Remove("missing")
		assert.Equal(t, int64(1), s.Card())
	})
}

func TestSet_Members(t *testing.T) {
	t.Run("returns all members", func(t *testing.T) {
		s := NewSet()
		assert.Empty(t, s.Members())

		s.Add("apple")
		s.Add("banana")
		s.Add("cherry")

		assert.Equal(t, int64(3), s.Card())
		assert.ElementsMatch(t, []string{"apple", "banana", "cherry"}, s.Members())
	})
}

func TestSet_Expiration(t *testing.T) {
	t.Run("initial state without expiration", func(t *testing.T) {
		s := NewSet()
		assert.False(t, s.IsExpired())
		assert.Equal(t, int64(0), s.TTL())
	})

	t.Run("negative TTL rejected", func(t *testing.T) {
		s := NewSet()
		assert.False(t, s.Expire(-1))
		assert.False(t, s.IsExpired())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		s := NewSet()
		assert.True(t, s.Expire(180))
		assert.False(t, s.IsExpired())
		assert.Greater(t, s.TTL(), int64(0))
		assert.LessOrEqual(t, s.TTL(), int64(180))
	})

	t.Run("detects expired state", func(t *testing.T) {
		s := NewSet()
		s.expiresAt.Store(time.Now().Unix() - 1)
		assert.True(t, s.IsExpired())
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		s := NewSet()
		s.expiresAt.Store(time.Now().Unix() - 1)
		assert.False(t, s.Expire(60))
	})
}

func TestSet_Concurrency(t *testing.T) {
	t.Run("concurrent adds, checks, and removes", func(t *testing.T) {
		s := NewSet()
		const goroutines = 30
		const ops = 100

		var wg sync.WaitGroup
		wg.Add(goroutines * 3)

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					s.Add("item")
				}
			}()
		}

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					s.IsMember("item")
					s.Card()
					s.Members()
				}
			}()
		}

		for range goroutines {
			go func() {
				defer wg.Done()
				for range ops {
					s.Remove("item")
				}
			}()
		}

		wg.Wait()
	})
}
