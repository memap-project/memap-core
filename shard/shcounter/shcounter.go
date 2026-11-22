package shcounter

import (
	"context"
	"hash/maphash"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedCounter is a partitioned map storing counters across multiple shards.
type ShardedCounter struct {
	shardCount uint8
	seed       maphash.Seed
	mask       uint64
	shards     []*shard.Shard[*vals.Counter]
}

// NewShardedCounter creates a new ShardedCounter with the given shard count and optional per-shard key limit.
// shardCount is normalized to a power of two in [1, 128].
func NewShardedCounter(shardCount uint8, maxKeysPerShard ...int) *ShardedCounter {
	shardCount = shard.NormalizeShardCount(shardCount)
	var limit int
	if len(maxKeysPerShard) > 0 && maxKeysPerShard[0] > 0 {
		limit = maxKeysPerShard[0]
	}
	shards := make([]*shard.Shard[*vals.Counter], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Counter](limit)
	}
	return &ShardedCounter{
		shardCount: shardCount,
		seed:       maphash.MakeSeed(),
		mask:       uint64(shardCount - 1),
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key using a per-instance seeded hash.
func (shc *ShardedCounter) getShard(key string) *shard.Shard[*vals.Counter] {
	return shc.shards[shard.SeededShardIndex(shc.seed, key, shc.mask)]
}

// SetLimit sets or updates the upper limit for the counter of the given key.
// Initializes the counter if it does not exist or is expired.
// Returns StatusInvalidArg without creating a counter if limit is negative.
func (shc *ShardedCounter) SetLimit(key string, limit int64) shard.Status {
	if limit < 0 {
		return shard.StatusInvalidArg
	}
	sh := shc.getShard(key)
	c, _ := sh.GetOrReplace(key, (*vals.Counter).IsExpired, func() *vals.Counter {
		nc := vals.NewCounter()
		nc.SetLimit(limit)
		return nc
	})
	c.SetLimit(limit)
	return shard.StatusSuccess
}

// GetLimit returns the upper limit of the counter for the given key.
// Returns 0 and false if the counter does not exist or is expired.
func (shc *ShardedCounter) GetLimit(key string) (int64, bool) {
	sh := shc.getShard(key)
	c, ok := sh.Get(key)
	if !ok {
		return 0, false
	}
	if c.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Counter) bool { return cur == c })
		return 0, false
	}
	return c.GetLimit(), true
}

// Get retrieves the value of the counter for the given key.
// Returns 0 and false if the counter does not exist or is expired.
func (shc *ShardedCounter) Get(key string) (int64, bool) {
	sh := shc.getShard(key)
	c, ok := sh.Get(key)
	if !ok {
		return 0, false
	}
	if c.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Counter) bool { return cur == c })
		return 0, false
	}
	return c.GetValue(), true
}

// Delete removes the counter for the given key.
func (shc *ShardedCounter) Delete(key string) {
	shc.getShard(key).Delete(key)
}

// Expire sets the expiration time for the counter of the given key.
// Returns StatusSuccess if expiration was set, StatusInvalidTTL if ttl is invalid or overflows,
// or StatusNotFound/StatusExpired if the counter does not exist or is already expired.
func (shc *ShardedCounter) Expire(key string, ttl int64) shard.Status {
	if !vals.IsValidTTL(ttl) {
		return shard.StatusInvalidTTL
	}
	now := vals.MonotonicNow()
	return shc.getShard(key).ExpireInPlace(
		key,
		func(c *vals.Counter) bool { return c.IsExpiredAt(now) },
		func(c *vals.Counter) bool { return c.ExpireAt(now, ttl) },
	)
}

// TTL returns the remaining time-to-live of the counter for the given key in seconds.
// Returns -1 if the counter exists and has no expiration time.
// Returns -2 if the counter does not exist or is expired.
func (shc *ShardedCounter) TTL(key string) int64 {
	sh := shc.getShard(key)
	counter, ok := sh.Get(key)
	if !ok {
		return -2
	}
	ttl := counter.TTL()
	if ttl == -2 {
		sh.DeleteIf(key, func(cur *vals.Counter) bool { return cur == counter })
		return -2
	}
	return ttl
}

// IncrBy atomically increments the counter for the given key by alpha (>= 0).
// Atomically initializes a fresh counter if it does not exist or is expired.
// Returns the exact post-increment value and StatusSuccess if the increment succeeded.
// Returns 0 and StatusInvalidArg if alpha < 0.
// Returns 0 and StatusLimitExceeded if the increment would overflow or exceed limit.
func (shc *ShardedCounter) IncrBy(key string, alpha int64) (int64, shard.Status) {
	if alpha < 0 {
		return 0, shard.StatusInvalidArg
	}
	sh := shc.getShard(key)
	c, _ := sh.GetOrReplace(key, (*vals.Counter).IsExpired, vals.NewCounter)
	val, ok := c.IncrBy(alpha)
	if !ok {
		return 0, shard.StatusLimitExceeded
	}
	return val, shard.StatusSuccess
}

// DecrBy atomically decrements the counter for the given key by alpha (>= 0).
// Returns the exact post-decrement value and StatusSuccess if the decrement succeeded.
// Returns 0 and StatusInvalidArg if alpha < 0.
// Returns 0 and StatusNotFound/StatusExpired/StatusLimitExceeded on failure.
func (shc *ShardedCounter) DecrBy(key string, alpha int64) (int64, shard.Status) {
	if alpha < 0 {
		return 0, shard.StatusInvalidArg
	}
	sh := shc.getShard(key)
	c, exist := sh.Get(key)
	if !exist {
		return 0, shard.StatusNotFound
	}
	if c.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Counter) bool { return cur == c })
		return 0, shard.StatusExpired
	}
	val, ok := c.DecrBy(alpha)
	if !ok {
		return 0, shard.StatusLimitExceeded
	}
	return val, shard.StatusSuccess
}

// Reset resets the counter value for the given key to 0.
// Returns true if the counter was reset.
// Returns false if the counter does not exist or is expired.
func (shc *ShardedCounter) Reset(key string) bool {
	sh := shc.getShard(key)
	c, ok := sh.Get(key)
	if !ok {
		return false
	}
	if c.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Counter) bool { return cur == c })
		return false
	}
	c.Reset()
	return true
}

// CleanExpired removes all expired counters across all shards.
func (shc *ShardedCounter) CleanExpired() {
	shc.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes all expired counters across all shards while respecting ctx cancellation.
func (shc *ShardedCounter) CleanExpiredCtx(ctx context.Context) {
	for _, sh := range shc.shards {
		if ctx.Err() != nil {
			return
		}
		now := vals.MonotonicNow()
		sh.CleanCtx(ctx, func(_ string, counter *vals.Counter) bool {
			return counter.IsExpiredAt(now)
		})
	}
}

// Flush removes all counters across all shards.
func (shc *ShardedCounter) Flush() {
	for _, sh := range shc.shards {
		sh.Flush()
	}
}
