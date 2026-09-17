package shset

import (
	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedHash is a partitioned set storing members across multiple shards.
type ShardedSet struct {
	shardCount uint8
	shards     []*shard.Shard[*vals.Set]
}

// NewShardedSet creates a new ShardedSet with the given number of shards.
func NewShardedSet(shardCount uint8) *ShardedSet {
	shards := make([]*shard.Shard[*vals.Set], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Set]()
	}
	return &ShardedSet{
		shardCount: shardCount,
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key based on FNV-1a hash.
func (s *ShardedSet) getShard(key string) *shard.Shard[*vals.Set] {
	return s.shards[shard.ShardIndex(key, s.shardCount)]
}

// Add adds a member to the Set for the given key.
// If the Set does not exist or is expired, a new Set is created.
func (s *ShardedSet) Add(key string, member string, ttl int64) {
	sh := s.getShard(key)
	set, ok := sh.GetOrInit(key, func() *vals.Set {
		return vals.NewSet()
	})
	if ok {
		ok := set.IsExpired()
		if ok {
			sh.Delete(key)
			sh.Set(key, vals.NewSet())
		}
	}
	set.Add(member)
}

// Remove removes a member from the Set for the given key.
func (s *ShardedSet) Remove(key string, member string) {
	set, ok := s.getShard(key).Get(key)
	if ok {
		set.Remove(member)
	}
}

// IsMember returns true if the member is a member of the Set for the given key.
func (s *ShardedSet) IsMember(key string, member string) bool {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return false
	}
	if set.IsExpired() {
		sh.Delete(key)
		return false
	}
	return set.IsMember(member)
}

// Card returns the number of members in the Set for the given key.
// Returns 0 and false if the Set does not exist or is expired.
func (s *ShardedSet) Card(key string) (int64, bool) {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return 0, false
	}
	if set.IsExpired() {
		sh.Delete(key)
		return 0, false
	}
	return set.Card(), true
}

// Members returns the members of the Set for the given key.
// Returns an empty slice and false if the Set does not exist or is expired.
func (s *ShardedSet) Members(key string) ([]string, bool) {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return []string{}, false
	}
	if set.IsExpired() {
		sh.Delete(key)
		return []string{}, false
	}
	return set.Members(), true
}

// Expire sets the expiration time for the Set of the given key.
// Returns true if the Set exists and expiration was set.
// Returns false if the Set does not exist or is already expired.
func (s *ShardedSet) Expire(key string, ttl int64) bool {
	ok := s.getShard(key).Update(key, func(h *vals.Set) bool {
		if h.IsExpired() {
			return false
		}
		h.Expire(ttl)
		return true
	})
	return ok
}

// TTL returns the time-to-live of the Set for the given key in seconds.
// Returns -1 if the Set exists and has no expiration time.
// Returns -2 if the Set does not exist or is expired.
func (s *ShardedSet) TTL(key string) int64 {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return -2
	}
	if set.IsExpired() {
		sh.Delete(key)
		return -2
	}
	if set.TTL() == 0 {
		return -1
	}
	return set.TTL()
}

// CleanExpired removes all expired sets across all shards.
func (s *ShardedSet) CleanExpired() {
	for _, shard := range s.shards {
		shard.Clean(func(_ string, set *vals.Set) bool {
			return set.IsExpired()
		})
	}
}

// Flush removes all sets across all shards.
func (s *ShardedSet) Flush() {
	for _, shard := range s.shards {
		shard.Flush()
	}
}
