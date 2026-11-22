package shmap

import (
	"context"
	"hash/maphash"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedMap is a partitioned map storing key-value pairs across multiple shards.
type ShardedMap struct {
	shardCount uint8
	seed       maphash.Seed
	mask       uint64
	shards     []*shard.Shard[*vals.Item]
}

// NewShardedMap creates a new ShardedMap with the given shard count and optional per-shard key limit.
// shardCount is normalized to a power of two in [1, 128].
func NewShardedMap(shardCount uint8, maxKeysPerShard ...int) *ShardedMap {
	shardCount = shard.NormalizeShardCount(shardCount)
	var limit int
	if len(maxKeysPerShard) > 0 && maxKeysPerShard[0] > 0 {
		limit = maxKeysPerShard[0]
	}
	shards := make([]*shard.Shard[*vals.Item], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Item](limit)
	}
	return &ShardedMap{
		shardCount: shardCount,
		seed:       maphash.MakeSeed(),
		mask:       uint64(shardCount - 1),
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key using a per-instance seeded hash.
func (shm *ShardedMap) getShard(key string) *shard.Shard[*vals.Item] {
	return shm.shards[shard.SeededShardIndex(shm.seed, key, shm.mask)]
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
		sh.DeleteIf(key, func(cur *vals.Item) bool { return cur == item })
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
	shm.getShard(key).SetWithEvictCheck(key, i, (*vals.Item).IsExpired)
}

// Delete removes the key from the sharded map.
func (shm *ShardedMap) Delete(key string) {
	shm.getShard(key).Delete(key)
}

// Expire sets the expiration time for the key.
// Returns StatusSuccess if expiration was set, StatusInvalidTTL if ttl is invalid or overflows,
// or StatusNotFound/StatusExpired if the key does not exist or is already expired.
func (shm *ShardedMap) Expire(key string, ttl int64) shard.Status {
	if !vals.IsValidTTL(ttl) {
		return shard.StatusInvalidTTL
	}
	now := vals.MonotonicNow()
	return shm.getShard(key).ExpireInPlace(
		key,
		func(i *vals.Item) bool { return i.IsExpiredAt(now) },
		func(i *vals.Item) bool { return i.ExpireAt(now, ttl) },
	)
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
	ttl := item.TTL()
	if ttl == -2 {
		sh.DeleteIf(key, func(cur *vals.Item) bool { return cur == item })
		return -2
	}
	return ttl
}

// CleanExpired removes all expired items across all shards.
func (shm *ShardedMap) CleanExpired() {
	shm.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes all expired items across all shards while respecting ctx cancellation.
func (shm *ShardedMap) CleanExpiredCtx(ctx context.Context) {
	for _, sh := range shm.shards {
		if ctx.Err() != nil {
			return
		}
		now := vals.MonotonicNow()
		sh.CleanCtx(ctx, func(_ string, item *vals.Item) bool {
			return item.IsExpiredAt(now)
		})
	}
}

// Flush removes all items across all shards.
func (shm *ShardedMap) Flush() {
	for _, sh := range shm.shards {
		sh.Flush()
	}
}
