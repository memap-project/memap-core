package shmap

import (
	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedMap is a partitioned map storing key-value pairs across multiple shards.
type ShardedMap struct {
	shardCount uint8
	shards     []*shard.Shard[*vals.Item]
}

// NewShardedMap creates a new ShardedMap with the given shard count.
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
func (shm *ShardedMap) getShard(key string) *shard.Shard[*vals.Item] {
	return shm.shards[shard.ShardIndex(key, shm.shardCount)]
}

// Get retrieves the value of the key from the sharded map.
// Returns empty string and false if the key does not exist or is expired.
func (shm *ShardedMap) Get(key string) (string, bool) {
	sh := shm.getShard(key)
	item, ok := sh.Get(key)
	if !ok {
		return "", false
	}
	if item.IsExpired() {
		sh.Delete(key)
		return "", false
	}
	return item.GetValue(), true
}

// Set sets or updates the value and optional TTL for the key in the sharded map.
func (shm *ShardedMap) Set(key, value string, ttl int64) {
	i := vals.NewItem(value)
	if ttl > 0 {
		i.Expire(ttl)
	}
	shm.getShard(key).Set(key, i)
}

// Delete removes the key from the sharded map.
func (shm *ShardedMap) Delete(key string) {
	shm.getShard(key).Delete(key)
}

// Expire sets the expiration time for the key.
// Returns true if the key exists and expiration was set.
// Returns false if the key does not exist or is already expired.
func (shm *ShardedMap) Expire(key string, ttl int64) bool {
	ok := shm.getShard(key).Update(key, func(i *vals.Item) bool {
		if i.IsExpired() {
			return false
		}
		i.Expire(ttl)
		return true
	})
	return ok
}

// TTL returns the time-to-live of the key in seconds.
// Returns -1 if the key exists and has no expiration time.
// Returns -2 if the key does not exist or is expired.
func (shm *ShardedMap) TTL(key string) int64 {
	sh := shm.getShard(key)
	item, ok := sh.Get(key)
	if !ok {
		return -2
	}
	if item.IsExpired() {
		sh.Delete(key)
		return -2
	}
	if item.TTL() == 0 {
		return -1
	}
	return item.TTL()
}

// CleanExpired removes all expired items across all shards.
func (shm *ShardedMap) CleanExpired() {
	for _, sh := range shm.shards {
		sh.Clean(func(key string, item *vals.Item) bool {
			return item.IsExpired()
		})
	}
}

// Flush removes all items across all shards.
func (shm *ShardedMap) Flush() {
	for _, sh := range shm.shards {
		sh.Flush()
	}
}
