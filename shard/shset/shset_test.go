package shset

import (
	"sync"
	"testing"
	"time"

	"github.com/memap-project/memap-core/shard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedSet_NewShardedSet(t *testing.T) {
	shs := NewShardedSet(4)
	require.NotNil(t, shs)
	assert.Equal(t, uint8(4), shs.shardCount)
	assert.Len(t, shs.shards, 4)

	assert.False(t, shs.IsMember("missing", "mem"))
}

func TestShardedSet_Add_and_IsMember(t *testing.T) {
	t.Run("adds members and checks membership", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("set1", "m1", 0)
		shs.Add("set1", "m2", 0)

		assert.True(t, shs.IsMember("set1", "m1"))
		assert.True(t, shs.IsMember("set1", "m2"))
		assert.False(t, shs.IsMember("set1", "m3"))
	})

	t.Run("membership on missing set returns false", func(t *testing.T) {
		shs := NewShardedSet(4)
		assert.False(t, shs.IsMember("missing", "m1"))
	})

	t.Run("adds member with TTL", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("set_ttl", "m1", 100)

		assert.True(t, shs.IsMember("set_ttl", "m1"))
		assert.Greater(t, shs.TTL("set_ttl"), int64(0))
	})

	t.Run("add on expired set re-initializes set", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("exp_set", "old_mem", 1)

		time.Sleep(2100 * time.Millisecond)

		shs.Add("exp_set", "new_mem", 0)

		assert.True(t, shs.IsMember("exp_set", "new_mem"))
		assert.False(t, shs.IsMember("exp_set", "old_mem"))
	})

	t.Run("isMember on expired set returns false and deletes key", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("exp_set2", "mem", 1)

		time.Sleep(2100 * time.Millisecond)

		assert.False(t, shs.IsMember("exp_set2", "mem"))
		assert.Equal(t, int64(-2), shs.TTL("exp_set2"))
	})
}

func TestShardedSet_Remove(t *testing.T) {
	t.Run("removes existing member", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("set1", "m1", 0)
		shs.Add("set1", "m2", 0)

		shs.Remove("set1", "m1")

		assert.False(t, shs.IsMember("set1", "m1"))
		assert.True(t, shs.IsMember("set1", "m2"))
	})

	t.Run("remove non-existent member is no-op", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("set1", "m1", 0)

		shs.Remove("set1", "missing_mem")
		assert.True(t, shs.IsMember("set1", "m1"))
	})

	t.Run("remove on missing set is no-op", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Remove("missing", "m1")
		assert.False(t, shs.IsMember("missing", "m1"))
	})
}

func TestShardedSet_Card_and_Members(t *testing.T) {
	t.Run("cardinality and members on populated set", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("set1", "a", 0)
		shs.Add("set1", "b", 0)

		c, ok := shs.Card("set1")
		assert.True(t, ok)
		assert.Equal(t, int64(2), c)

		mems, ok := shs.Members("set1")
		assert.True(t, ok)
		assert.ElementsMatch(t, []string{"a", "b"}, mems)
	})

	t.Run("card and members on missing set", func(t *testing.T) {
		shs := NewShardedSet(4)

		c, ok := shs.Card("missing")
		assert.False(t, ok)
		assert.Equal(t, int64(0), c)

		mems, ok := shs.Members("missing")
		assert.False(t, ok)
		assert.Empty(t, mems)
	})

	t.Run("card and members on expired set return false and delete key", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("exp_set", "a", 1)

		time.Sleep(2100 * time.Millisecond)

		c, ok := shs.Card("exp_set")
		assert.False(t, ok)
		assert.Equal(t, int64(0), c)

		shs.Add("exp_set2", "a", 1)
		time.Sleep(2100 * time.Millisecond)

		mems, ok := shs.Members("exp_set2")
		assert.False(t, ok)
		assert.Empty(t, mems)
	})
}

func TestShardedSet_Expire_and_TTL(t *testing.T) {
	t.Run("missing set TTL returns -2 and Expire returns StatusNotFound", func(t *testing.T) {
		shs := NewShardedSet(4)
		assert.Equal(t, int64(-2), shs.TTL("missing"))
		assert.Equal(t, shard.StatusNotFound, shs.Expire("missing", 10))
	})

	t.Run("set without TTL returns -1", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("s1", "m", 0)
		assert.Equal(t, int64(-1), shs.TTL("s1"))
	})

	t.Run("sets valid TTL and rejects invalid TTL", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("s1", "m", 0)

		assert.Equal(t, shard.StatusInvalidTTL, shs.Expire("s1", 0))
		assert.Equal(t, shard.StatusInvalidTTL, shs.Expire("s1", -1))
		assert.Equal(t, shard.StatusSuccess, shs.Expire("s1", 100))
		ttl := shs.TTL("s1")
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(100))
	})

	t.Run("expired set returns TTL -2 and Expire returns StatusExpired", func(t *testing.T) {
		shs := NewShardedSet(4)
		shs.Add("exp_s", "m", 1)

		time.Sleep(1100 * time.Millisecond)

		assert.Equal(t, shard.StatusExpired, shs.Expire("exp_s", 10))
		assert.Equal(t, int64(-2), shs.TTL("exp_s"))
	})
}

func TestShardedSet_CleanExpired(t *testing.T) {
	shs := NewShardedSet(4)
	shs.Add("active", "m", 0)
	shs.Add("exp1", "m", 1)
	shs.Add("exp2", "m", 1)

	time.Sleep(2100 * time.Millisecond)

	shs.CleanExpired()

	assert.True(t, shs.IsMember("active", "m"))
	assert.Equal(t, int64(-2), shs.TTL("exp1"))
	assert.Equal(t, int64(-2), shs.TTL("exp2"))
}

func TestShardedSet_Flush(t *testing.T) {
	shs := NewShardedSet(4)
	shs.Add("s1", "m1", 0)
	shs.Add("s2", "m2", 0)

	shs.Flush()

	assert.False(t, shs.IsMember("s1", "m1"))
	assert.False(t, shs.IsMember("s2", "m2"))
}

func TestShardedSet_Concurrency(t *testing.T) {
	shs := NewShardedSet(8)
	const goroutines = 20
	const ops = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Adders
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				shs.Add(string(rune('a'+(id%8))), string(rune('0'+(i%10))), 0)
			}
		}(g)
	}

	// Checkers (IsMember, Card, Members)
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				key := string(rune('a' + (id % 8)))
				shs.IsMember(key, string(rune('0'+(i%10))))
				shs.Card(key)
				shs.Members(key)
			}
		}(g)
	}

	// Removers
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				shs.Remove(string(rune('a'+(id%8))), string(rune('0'+(i%10))))
			}
		}(g)
	}

	wg.Wait()
}
