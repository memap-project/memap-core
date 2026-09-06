package shcounter

import (
	"hash/fnv"

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
func (s *ShardedCounter) getShard(key string) *shard.Shard[*vals.Counter] {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() & uint32(s.shardCount-1)
	return s.shards[idx]
}

// Init initializes a new counter with the given key, limit, and optional TTL.
// Returns true if the counter was created.
// Returns false if a counter already exists for the given key.
func (s *ShardedCounter) Init(key string, limit int64, ttl int64) bool {
	c, ok := s.getShard(key).GetOrInit(key, vals.NewCounter)
	if ok {
		return false
	}
	c.Expire(ttl)
	c.SetLimit(limit)
	return true
}

// SetLimit sets or updates the upper limit for the counter of the given key.
// Returns true if the limit was set.
// Returns false if the counter does not exist or is expired.
func (s *ShardedCounter) SetLimit(key string, limit int64) bool {
	shard := s.getShard(key)
	c, ok := shard.Get(key)
	if !ok {
		return false
	}
	if c.IsExpired() {
		shard.Delete(key)
		return false
	}
	c.SetLimit(limit)
	return true
}

// GetLimit returns the upper limit of the counter for the given key.
// Returns 0 and false if the counter does not exist or is expired.
func (s *ShardedCounter) GetLimit(key string) (int64, bool) {
	shard := s.getShard(key)
	c, ok := shard.Get(key)
	if !ok {
		return 0, false
	}
	if c.IsExpired() {
		shard.Delete(key)
		return 0, false
	}
	return c.GetLimit(), true
}

// Get retrieves the value of the counter for the given key.
// Returns 0 and false if the counter does not exist or is expired.
func (s *ShardedCounter) Get(key string) (int64, bool) {
	shard := s.getShard(key)
	c, ok := shard.Get(key)
	if !ok {
		return 0, false
	}
	if c.IsExpired() {
		shard.Delete(key)
		return 0, false
	}
	return c.GetValue(), true
}

// Delete removes the counter for the given key.
func (s *ShardedCounter) Delete(key string) {
	s.getShard(key).Delete(key)
}

// Expire sets the expiration time for the counter of the given key.
// Returns true if the counter exists and expiration was set.
// Returns false if the counter does not exist or is already expired.
func (s *ShardedCounter) Expire(key string, ttl int64) bool {
	return s.getShard(key).Update(key, func(c *vals.Counter) bool {
		if c.IsExpired() {
			return false
		}
		c.Expire(ttl)
		return true
	})
}

// TTL returns the remaining time-to-live of the counter for the given key in seconds.
// Returns -1 and true if the counter exists and has no expiration time.
// Returns -2 and false if the counter does not exist or is expired.
func (s *ShardedCounter) TTL(key string) (int64, bool) {
	shard := s.getShard(key)
	c, ok := shard.Get(key)
	if !ok || c.IsExpired() {
		return -2, false
	}
	if c.TTL() == 0 {
		return -1, true
	}
	return c.TTL(), true
}

// IncrBy increments the counter for the given key by alpha.
// Returns the new value and StatusSuccess if the increment succeeded.
// Returns 0 and a failure Status if the counter does not exist, is expired, or if the increment exceeds limit.
func (s *ShardedCounter) IncrBy(key string, alpha int64) (int64, shard.Status) {
	sh := s.getShard(key)
	c, ok := sh.Get(key)
	if !ok {
		return 0, shard.StatusNotFound
	}
	if c.IsExpired() {
		sh.Delete(key)
		return 0, shard.StatusExpired
	}
	ok = c.IncrBy(alpha)
	if !ok {
		return 0, shard.StatusLimitExceeded
	}
	return c.GetValue(), shard.StatusSuccess
}

// DecrBy decrements the counter for the given key by alpha.
// Returns the new value and StatusSuccess if the decrement succeeded.
// Returns 0 and a failure Status if the counter does not exist, is expired, or if the decrement would result in a negative value.
func (s *ShardedCounter) DecrBy(key string, alpha int64) (int64, shard.Status) {
	sh := s.getShard(key)
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
func (s *ShardedCounter) Reset(key string) bool {
	shard := s.getShard(key)
	c, ok := shard.Get(key)
	if !ok {
		return false
	}
	if c.IsExpired() {
		shard.Delete(key)
		return false
	}
	c.Reset()
	return true
}

// CleanExpired removes all expired counters across all shards.
func (s *ShardedCounter) CleanExpired() {
	for _, shard := range s.shards {
		shard.Clean(func(key string, counter *vals.Counter) bool {
			return counter.IsExpired()
		})
	}
}

// Flush removes all counters across all shards.
func (s *ShardedCounter) Flush() {
	for _, shard := range s.shards {
		shard.Flush()
	}
}
