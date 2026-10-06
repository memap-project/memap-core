package shard

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HashKey(t *testing.T) {
	t.Run("hash for empty string is offset32", func(t *testing.T) {
		assert.Equal(t, offset32, HashKey(""))
	})

	t.Run("deterministic hash for same string", func(t *testing.T) {
		h1 := HashKey("user:12345")
		h2 := HashKey("user:12345")
		assert.Equal(t, h1, h2)
		assert.NotEqual(t, offset32, h1)
	})

	t.Run("different keys produce different hashes", func(t *testing.T) {
		assert.NotEqual(t, HashKey("key1"), HashKey("key2"))
		assert.NotEqual(t, HashKey("abc"), HashKey("cba"))
	})
}

func Test_ShardIndex(t *testing.T) {
	t.Run("shardCount 0 or 1 returns 0", func(t *testing.T) {
		assert.Equal(t, uint32(0), ShardIndex("key", 0))
		assert.Equal(t, uint32(0), ShardIndex("key", 1))
	})

	t.Run("shardCount power of two distributes within bounds", func(t *testing.T) {
		counts := []uint8{2, 4, 8, 16, 32, 64}
		for _, sc := range counts {
			for _, k := range []string{"alpha", "beta", "gamma", "delta", "epsilon"} {
				idx := ShardIndex(k, sc)
				assert.Less(t, idx, uint32(sc))
			}
		}
	})

	t.Run("deterministic shard index for same key", func(t *testing.T) {
		idx1 := ShardIndex("my-key", 8)
		idx2 := ShardIndex("my-key", 8)
		assert.Equal(t, idx1, idx2)
	})
}

func TestShard_NewShard(t *testing.T) {
	s := NewShard[string]()
	require.NotNil(t, s)

	val, ok := s.Get("any")
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestShard_Set_and_Get(t *testing.T) {
	t.Run("get missing key returns zero value and false", func(t *testing.T) {
		s := NewShard[int]()
		val, ok := s.Get("missing")
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("set and get value", func(t *testing.T) {
		s := NewShard[string]()
		s.Set("name", "Alice")

		val, ok := s.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("overwrite existing value", func(t *testing.T) {
		s := NewShard[string]()
		s.Set("name", "Alice")
		s.Set("name", "Bob")

		val, ok := s.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Bob", val)
	})

	t.Run("stores zero value correctly", func(t *testing.T) {
		s := NewShard[int]()
		s.Set("counter", 0)

		val, ok := s.Get("counter")
		assert.True(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestShard_GetOrInit(t *testing.T) {
	t.Run("initializes and stores new value when key does not exist", func(t *testing.T) {
		s := NewShard[string]()
		initCalled := false

		val, existed := s.GetOrInit("greeting", func() string {
			initCalled = true
			return "hello"
		})

		assert.False(t, existed)
		assert.True(t, initCalled)
		assert.Equal(t, "hello", val)

		got, ok := s.Get("greeting")
		assert.True(t, ok)
		assert.Equal(t, "hello", got)
	})

	t.Run("returns existing value without calling initFn when key exists", func(t *testing.T) {
		s := NewShard[string]()
		s.Set("greeting", "existing")

		initCalled := false
		val, existed := s.GetOrInit("greeting", func() string {
			initCalled = true
			return "new_hello"
		})

		assert.True(t, existed)
		assert.False(t, initCalled)
		assert.Equal(t, "existing", val)
	})

	t.Run("handles race between read unlock and write lock", func(t *testing.T) {
		s := NewShard[int]()
		const goroutines = 30
		const iterations = 100

		for range iterations {
			var wg sync.WaitGroup
			wg.Add(goroutines)
			start := make(chan struct{})

			for range goroutines {
				go func() {
					defer wg.Done()
					<-start
					val, _ := s.GetOrInit("key", func() int {
						return 42
					})
					assert.Equal(t, 42, val)
				}()
			}

			close(start)
			wg.Wait()
			s.Delete("key")
		}
	})
}

func TestShard_Delete(t *testing.T) {
	t.Run("delete existing key", func(t *testing.T) {
		s := NewShard[int]()
		s.Set("key1", 100)
		s.Delete("key1")

		val, ok := s.Get("key1")
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("delete non-existent key is a no-op", func(t *testing.T) {
		s := NewShard[int]()
		s.Delete("non_existent")

		val, ok := s.Get("non_existent")
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestShard_Update(t *testing.T) {
	t.Run("returns false when key does not exist", func(t *testing.T) {
		s := NewShard[int]()
		called := false

		updated := s.Update("missing", func(val int) bool {
			called = true
			return true
		})

		assert.False(t, updated)
		assert.False(t, called)
	})

	t.Run("updates value when predicate returns true", func(t *testing.T) {
		type user struct {
			score int
		}
		s := NewShard[*user]()
		s.Set("user1", &user{score: 10})

		updated := s.Update("user1", func(u *user) bool {
			u.score = 25
			return true
		})

		assert.True(t, updated)
		got, ok := s.Get("user1")
		assert.True(t, ok)
		assert.Equal(t, 25, got.score)
	})

	t.Run("does not update when predicate returns false", func(t *testing.T) {
		type user struct {
			score int
		}
		s := NewShard[*user]()
		s.Set("user1", &user{score: 10})

		updated := s.Update("user1", func(u *user) bool {
			return false
		})

		assert.False(t, updated)
		got, ok := s.Get("user1")
		assert.True(t, ok)
		assert.Equal(t, 10, got.score)
	})
}

func TestShard_Clean(t *testing.T) {
	t.Run("removes only matching items", func(t *testing.T) {
		s := NewShard[int]()
		s.Set("active:1", 1)
		s.Set("active:2", 2)
		s.Set("expired:1", -1)
		s.Set("expired:2", -2)

		s.Clean(func(key string, val int) bool {
			return val < 0
		})

		_, ok1 := s.Get("active:1")
		assert.True(t, ok1)
		_, ok2 := s.Get("active:2")
		assert.True(t, ok2)
		_, ok3 := s.Get("expired:1")
		assert.False(t, ok3)
		_, ok4 := s.Get("expired:2")
		assert.False(t, ok4)
	})

	t.Run("clean with no matching items keeps all", func(t *testing.T) {
		s := NewShard[int]()
		s.Set("a", 1)
		s.Clean(func(key string, val int) bool {
			return false
		})

		val, ok := s.Get("a")
		assert.True(t, ok)
		assert.Equal(t, 1, val)
	})
}

func TestShard_Flush(t *testing.T) {
	t.Run("removes all items from shard", func(t *testing.T) {
		s := NewShard[int]()
		s.Set("a", 1)
		s.Set("b", 2)
		s.Set("c", 3)

		s.Flush()

		_, okA := s.Get("a")
		assert.False(t, okA)
		_, okB := s.Get("b")
		assert.False(t, okB)
		_, okC := s.Get("c")
		assert.False(t, okC)
	})
}

func TestShard_Concurrency(t *testing.T) {
	t.Run("concurrent reads, writes, updates, and deletes", func(t *testing.T) {
		s := NewShard[int]()
		const goroutines = 20
		const ops = 100

		var wg sync.WaitGroup
		wg.Add(goroutines * 4)

		// Writers
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for i := range ops {
					s.Set(string(rune('a'+(base%26))), i)
				}
			}(g)
		}

		// Readers
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for range ops {
					s.Get(string(rune('a' + (base % 26))))
				}
			}(g)
		}

		// Updaters
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for i := range ops {
					s.Update(string(rune('a'+(base%26))), func(val int) bool {
						return val%2 == i%2
					})
				}
			}(g)
		}

		// Deleters
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for range ops {
					s.Delete(string(rune('a' + (base % 26))))
				}
			}(g)
		}

		wg.Wait()
	})

	t.Run("concurrent GetOrInit on same key initializes once", func(t *testing.T) {
		s := NewShard[int]()
		const goroutines = 50

		var wg sync.WaitGroup
		wg.Add(goroutines)

		var initCalls atomic.Int64
		var createdCount atomic.Int64

		for range goroutines {
			go func() {
				defer wg.Done()
				val, existed := s.GetOrInit("singleton", func() int {
					initCalls.Add(1)
					return 42
				})
				if !existed {
					createdCount.Add(1)
				}
				assert.Equal(t, 42, val)
			}()
		}

		wg.Wait()

		// Exactly 1 goroutine created it
		assert.Equal(t, int64(1), initCalls.Load())
		assert.Equal(t, int64(1), createdCount.Load())

		val, ok := s.Get("singleton")
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})
}
