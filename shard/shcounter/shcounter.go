package shcounter

import (
	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedCounter is a partitioned map storing counters across multiple shards.
type ShardedCounter struct {
	shardCount uint8
	shards     []*shard.Shard[*vals.Counter]
}

// NewShardedCounter creates a new ShardedCounter with the given shard count.
func NewShardedCounter(shardCount uint8) *ShardedCounter {
	shards := make([]*shard.Shard[*vals.Counter], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Counter]()
	}
	return &ShardedCounter{
		shardCount: shardCount,
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key based on FNV-1a hash.
func (shc *ShardedCounter) getShard(key string) *shard.Shard[*vals.Counter] {
	return shc.shards[shard.ShardIndex(key, shc.shardCount)]
}


// SetLimit sets or updates the upper limit for the counter of the given key.
// Initializes the counter if it does not exist or is expired.
func (shc *ShardedCounter) SetLimit(key string, limit int64) {
	sh := shc.getShard(key)
	c, ok := sh.GetOrInit(key, vals.NewCounter)
	if ok && c.IsExpired() {
		c = vals.NewCounter()
		sh.Set(key, c)
	}
	c.SetLimit(limit)
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
		sh.Delete(key)
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
		sh.Delete(key)
		return 0, false
	}
	return c.GetValue(), true
}

// Delete removes the counter for the given key.
func (shc *ShardedCounter) Delete(key string) {
	shc.getShard(key).Delete(key)
}

// Expire sets the expiration time for the counter of the given key.
// Returns true if the counter exists and expiration was set.
// Returns false if the counter does not exist or is already expired.
func (shc *ShardedCounter) Expire(key string, ttl int64) bool {
	ok := shc.getShard(key).Update(key, func(c *vals.Counter) bool {
		if c.IsExpired() {
			return false
		}
		c.Expire(ttl)
		return true
	})
	return ok
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
	if counter.IsExpired() {
		sh.Delete(key)
		return -2
	}
	if counter.TTL() == 0 {
		return -1
	}
	return counter.TTL()
}

// IncrBy increments the counter for the given key by alpha.
// Initializes the counter if it does not exist or is expired.
// Returns the new value and StatusSuccess if the increment succeeded.
// Returns 0 and StatusLimitExceeded if the increment exceeds limit.
func (shc *ShardedCounter) IncrBy(key string, alpha int64) (int64, shard.Status) {
	sh := shc.getShard(key)
	c, ok := sh.GetOrInit(key, vals.NewCounter)
	if ok && c.IsExpired() {
		lim := c.GetLimit()
		c = vals.NewCounter()
		c.SetLimit(lim)
		sh.Set(key, c)
	}
	if !c.IncrBy(alpha) {
		return 0, shard.StatusLimitExceeded
	}
	return c.GetValue(), shard.StatusSuccess
}

// DecrBy decrements the counter for the given key by alpha.
// Returns the new value and StatusSuccess if the decrement succeeded.
// Returns 0 and a failure Status if the counter does not exist, is expired, or if the decrement would result in a negative value.
func (shc *ShardedCounter) DecrBy(key string, alpha int64) (int64, shard.Status) {
	sh := shc.getShard(key)
	c, exist := sh.Get(key)
	if !exist {
		return 0, shard.StatusNotFound
	}
	if c.IsExpired() {
		sh.Delete(key)
		return 0, shard.StatusExpired
	}
	ok := c.DecrBy(alpha)
	if !ok {
		return 0, shard.StatusLimitExceeded
	}
	return c.GetValue(), shard.StatusSuccess
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
		sh.Delete(key)
		return false
	}
	c.Reset()
	return true
}

// CleanExpired removes all expired counters across all shards.
func (shc *ShardedCounter) CleanExpired() {
	for _, sh := range shc.shards {
		sh.Clean(func(key string, counter *vals.Counter) bool {
			return counter.IsExpired()
		})
	}
}

// Flush removes all counters across all shards.
func (shc *ShardedCounter) Flush() {
	for _, sh := range shc.shards {
		sh.Flush()
	}
}
