package shmap

import (
	"hash/fnv"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedMap is a partitioned map storing string key-value pairs across multiple shards.
type ShardedMap struct {
	shardCount uint8
	shards     []*shard.Shard[*vals.Item]
}

// NewShardedMap creates a new ShardedMap with default shard count (8).
func NewShardedMap(shardCount uint8) *ShardedMap {
	shards := make([]*shard.Shard[*vals.Item], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Item]()
	}
	return &ShardedMap{
		shardCount: shardCount,
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key based on FNV-1a hash.
func (s *ShardedMap) getShard(key string) *shard.Shard[*vals.Item] {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() & uint32(s.shardCount-1)
	return s.shards[idx]
}

// Get retrieves the value of the key from the sharded map.
// Returns empty string and false if the key does not exist or is expired.
func (s *ShardedMap) Get(key string) (string, bool) {
	shard := s.getShard(key)
	i, ok := shard.Get(key)
	if !ok {
		return "", false
	}
	if i.IsExpired() {
		shard.Delete(key)
		return "", false
	}
	return i.GetValue(), true
}

// Set sets or updates the value and optional TTL for the key in the sharded map.
func (s *ShardedMap) Set(key, value string, ttl int64) bool {
	i := vals.NewItem(value)
	if ttl > 0 {
		i.Expire(ttl)
	}
	s.getShard(key).Set(key, i)
	return true
}

// Delete removes the key from the sharded map.
func (s *ShardedMap) Delete(key string) {
	s.getShard(key).Delete(key)
}

// Expire sets the expiration time for the key.
// Returns true if the key exists and expiration was set.
// Returns false if the key does not exist or is already expired.
func (s *ShardedMap) Expire(key string, ttl int64) bool {
	return s.getShard(key).Update(key, func(i *vals.Item) bool {
		if i.IsExpired() {
			return false
		}
		i.Expire(ttl)
		return true
	})
}

// TTL returns the time-to-live of the key in seconds.
// Returns -1 if the key exists and has no expiration time.
// Returns -2 if the key does not exist or is expired.
func (s *ShardedMap) TTL(key string) int64 {
	i, ok := s.getShard(key).Get(key)
	if !ok || i.IsExpired() {
		return -2
	}
	if i.TTL() == 0 {
		return -1
	}
	return i.TTL()
}

// CleanExpired removes all expired items across all shards.
func (s *ShardedMap) CleanExpired() {
	for _, shard := range s.shards {
		shard.Clean(func(key string, item *vals.Item) bool {
			return item.IsExpired()
		})
	}
}

// Flush removes all items across all shards.
func (s *ShardedMap) Flush() {
	for _, shard := range s.shards {
		shard.Flush()
	}
}
