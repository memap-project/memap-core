package shmap

import (
	"context"
	"hash/maphash"
	"math"
	"time"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// kvEntry stores a KV value and monotonic expiration inline by value inside the shard map,
// avoiding per-key heap allocations and pointer chasing.
type kvEntry struct {
	value     string
	expiresAt int64 // monotonic ns since process start; 0 = no expiration
}

func (e kvEntry) isExpiredAt(nowNs int64) bool {
	return e.expiresAt != 0 && nowNs >= e.expiresAt
}

func (e kvEntry) isExpired() bool {
	return e.isExpiredAt(vals.MonotonicNow())
}

func (e kvEntry) ttlAt(nowNs int64) int64 {
	if e.expiresAt == 0 {
		return -1
	}
	remNs := e.expiresAt - nowNs
	if remNs <= 0 {
		return -2
	}
	sec := remNs / int64(time.Second)
	if remNs%int64(time.Second) > 0 {
		sec++
	}
	return sec
}

// ShardedMap is a partitioned map storing key-value pairs across multiple shards.
type ShardedMap struct {
	shardCount uint8
	seed       maphash.Seed
	mask       uint64
	shards     []*shard.Shard[kvEntry]
}

// NewShardedMap creates a new ShardedMap with the given shard count and optional per-shard key limit.
// shardCount is normalized to a power of two in [1, 128].
func NewShardedMap(shardCount uint8, maxKeysPerShard ...int) *ShardedMap {
	shardCount = shard.NormalizeShardCount(shardCount)
	var limit int
	if len(maxKeysPerShard) > 0 && maxKeysPerShard[0] > 0 {
		limit = maxKeysPerShard[0]
	}
	shards := make([]*shard.Shard[kvEntry], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[kvEntry](limit)
	}
	return &ShardedMap{
		shardCount: shardCount,
		seed:       maphash.MakeSeed(),
		mask:       uint64(shardCount - 1),
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key using a per-instance seeded hash.
func (shm *ShardedMap) getShard(key string) *shard.Shard[kvEntry] {
	return shm.shards[shard.SeededShardIndex(shm.seed, key, shm.mask)]
}

// Get retrieves the value of the key from the sharded map.
// Returns empty string and false if the key does not exist or is expired.
func (shm *ShardedMap) Get(key string) (string, bool) {
	sh := shm.getShard(key)
	entry, ok := sh.Get(key)
	if !ok {
		return "", false
	}
	if entry.isExpired() {
		sh.DeleteIf(key, func(cur kvEntry) bool { return cur == entry })
		return "", false
	}
	return entry.value, true
}

// Set sets or updates the value and optional TTL for the key in the sharded map.
func (shm *ShardedMap) Set(key, value string, ttl int64) {
	var exp int64
	if ttl > 0 {
		now := vals.MonotonicNow()
		if ttl <= (math.MaxInt64-now)/int64(time.Second) {
			exp = now + ttl*int64(time.Second)
		}
	}
	entry := kvEntry{value: value, expiresAt: exp}
	shm.getShard(key).SetWithEvictCheck(key, entry, kvEntry.isExpired)
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
	if ttl > (math.MaxInt64-now)/int64(time.Second) {
		return shard.StatusInvalidTTL
	}
	newExp := now + ttl*int64(time.Second)
	return shm.getShard(key).ModifyValue(key, func(e *kvEntry) (bool, bool, shard.Status) {
		if e.isExpiredAt(now) {
			return false, true, shard.StatusExpired
		}
		e.expiresAt = newExp
		return true, false, shard.StatusSuccess
	})
}

// TTL returns the time-to-live of the key in seconds.
// Returns -1 if the key exists and has no expiration time.
// Returns -2 if the key does not exist or is expired.
func (shm *ShardedMap) TTL(key string) int64 {
	sh := shm.getShard(key)
	entry, ok := sh.Get(key)
	if !ok {
		return -2
	}
	ttl := entry.ttlAt(vals.MonotonicNow())
	if ttl == -2 {
		sh.DeleteIf(key, func(cur kvEntry) bool { return cur == entry })
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
		sh.CleanCtx(ctx, func(_ string, entry kvEntry) bool {
			return entry.isExpiredAt(now)
		})
	}
}

// Flush removes all items across all shards.
func (shm *ShardedMap) Flush() {
	for _, sh := range shm.shards {
		sh.Flush()
	}
}
