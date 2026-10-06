package tsmap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypedSyncMap_Store_and_Load(t *testing.T) {
	t.Run("load from empty map returns zero value and false", func(t *testing.T) {
		var tm TypedSyncMap[string, int]

		val, ok := tm.Load("missing")
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("stores and loads values", func(t *testing.T) {
		var tm TypedSyncMap[string, string]

		tm.Store("user:1", "Alice")
		tm.Store("user:2", "Bob")

		val1, ok1 := tm.Load("user:1")
		assert.True(t, ok1)
		assert.Equal(t, "Alice", val1)

		val2, ok2 := tm.Load("user:2")
		assert.True(t, ok2)
		assert.Equal(t, "Bob", val2)
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		var tm TypedSyncMap[string, string]

		tm.Store("name", "Alice")
		tm.Store("name", "Bob")

		val, ok := tm.Load("name")
		assert.True(t, ok)
		assert.Equal(t, "Bob", val)
	})

	t.Run("stores zero value correctly", func(t *testing.T) {
		var tm TypedSyncMap[string, int]

		tm.Store("zero", 0)

		val, ok := tm.Load("zero")
		assert.True(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("supports struct and pointer types", func(t *testing.T) {
		type user struct {
			age int
		}
		var tm TypedSyncMap[int, *user]

		u := &user{age: 25}
		tm.Store(1, u)

		val, ok := tm.Load(1)
		assert.True(t, ok)
		assert.Same(t, u, val)
	})
}

func TestTypedSyncMap_LoadOrStore(t *testing.T) {
	t.Run("stores and returns new value when key is absent", func(t *testing.T) {
		var tm TypedSyncMap[string, int]

		val, loaded := tm.LoadOrStore("counter", 10)
		assert.False(t, loaded)
		assert.Equal(t, 10, val)

		got, ok := tm.Load("counter")
		assert.True(t, ok)
		assert.Equal(t, 10, got)
	})

	t.Run("returns existing value when key is present", func(t *testing.T) {
		var tm TypedSyncMap[string, int]

		tm.Store("counter", 10)

		val, loaded := tm.LoadOrStore("counter", 99)
		assert.True(t, loaded)
		assert.Equal(t, 10, val)

		got, ok := tm.Load("counter")
		assert.True(t, ok)
		assert.Equal(t, 10, got)
	})
}

func TestTypedSyncMap_Delete(t *testing.T) {
	t.Run("deletes existing key", func(t *testing.T) {
		var tm TypedSyncMap[string, string]

		tm.Store("k1", "v1")
		tm.Delete("k1")

		val, ok := tm.Load("k1")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("delete non-existent key is a no-op", func(t *testing.T) {
		var tm TypedSyncMap[string, string]

		tm.Delete("non_existent")

		val, ok := tm.Load("non_existent")
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestTypedSyncMap_Range(t *testing.T) {
	t.Run("iterates over empty map without calls", func(t *testing.T) {
		var tm TypedSyncMap[string, int]
		called := false

		tm.Range(func(key string, value int) bool {
			called = true
			return true
		})

		assert.False(t, called)
	})

	t.Run("iterates over all key-value pairs", func(t *testing.T) {
		var tm TypedSyncMap[string, int]
		expected := map[string]int{"a": 1, "b": 2, "c": 3}

		for k, v := range expected {
			tm.Store(k, v)
		}

		result := make(map[string]int)
		tm.Range(func(key string, value int) bool {
			result[key] = value
			return true
		})

		assert.Equal(t, expected, result)
	})

	t.Run("stops iteration when callback returns false", func(t *testing.T) {
		var tm TypedSyncMap[int, int]

		for i := range 10 {
			tm.Store(i, i*10)
		}

		count := 0
		tm.Range(func(key, value int) bool {
			count++
			return false
		})

		assert.Equal(t, 1, count)
	})
}

func TestTypedSyncMap_Concurrency(t *testing.T) {
	t.Run("concurrent reads, writes, and deletes", func(t *testing.T) {
		var tm TypedSyncMap[int, int]
		const goroutines = 30
		const ops = 100

		var wg sync.WaitGroup
		wg.Add(goroutines * 3)

		// Writers
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for i := range ops {
					tm.Store(base*ops+i, i)
				}
			}(g)
		}

		// Readers
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for i := range ops {
					tm.Load(base*ops + i)
				}
			}(g)
		}

		// Deleters
		for g := range goroutines {
			go func(base int) {
				defer wg.Done()
				for i := range ops {
					tm.Delete(base*ops + i)
				}
			}(g)
		}

		wg.Wait()
	})

	t.Run("concurrent LoadOrStore on identical key", func(t *testing.T) {
		var tm TypedSyncMap[string, int]
		const goroutines = 50

		var wg sync.WaitGroup
		wg.Add(goroutines)

		var mu sync.Mutex
		loadedCount := 0

		for g := range goroutines {
			go func(id int) {
				defer wg.Done()
				val, loaded := tm.LoadOrStore("shared_key", id)

				mu.Lock()
				if loaded {
					loadedCount++
				}
				mu.Unlock()

				// All routines must observe the value that was actually stored
				got, ok := tm.Load("shared_key")
				assert.True(t, ok)
				assert.Equal(t, val, got)
			}(g)
		}

		wg.Wait()

		// Exactly 1 goroutine stored the value (loaded == false), the remaining goroutines loaded it
		assert.Equal(t, goroutines-1, loadedCount)
	})
}
