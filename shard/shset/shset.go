package shset

import (
	"context"
	"hash/maphash"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedSet is a partitioned set storing members across multiple shards.
type ShardedSet struct {
	shardCount uint8
	seed       maphash.Seed
	mask       uint64
	shards     []*shard.Shard[*vals.Set]
}

// NewShardedSet creates a new ShardedSet with the given number of shards and optional per-shard key limit.
// shardCount is normalized to a power of two in [1, 128].
func NewShardedSet(shardCount uint8, maxKeysPerShard ...int) *ShardedSet {
	shardCount = shard.NormalizeShardCount(shardCount)
	var limit int
	if len(maxKeysPerShard) > 0 && maxKeysPerShard[0] > 0 {
		limit = maxKeysPerShard[0]
	}
	shards := make([]*shard.Shard[*vals.Set], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Set](limit)
	}
	return &ShardedSet{
		shardCount: shardCount,
		seed:       maphash.MakeSeed(),
		mask:       uint64(shardCount - 1),
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key using a per-instance seeded hash.
func (s *ShardedSet) getShard(key string) *shard.Shard[*vals.Set] {
	return s.shards[shard.SeededShardIndex(s.seed, key, s.mask)]
}

// Add adds a member to the Set for the given key.
// Atomically creates or replaces the Set if it does not exist or is expired.
func (s *ShardedSet) Add(key string, member string, ttl int64) {
	sh := s.getShard(key)
	set, existed := sh.GetOrReplace(key, (*vals.Set).IsExpired, func() *vals.Set {
		ns := vals.NewSet()
		if ttl > 0 {
			ns.Expire(ttl)
		}
		return ns
	})
	if existed && ttl > 0 {
		set.Expire(ttl)
	}
	set.Add(member)
}

// Remove removes a member from the Set for the given key.
// Lazily deletes the set entry if it is already expired.
func (s *ShardedSet) Remove(key string, member string) {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return
	}
	if set.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Set) bool { return cur == set })
		return
	}
	set.Remove(member)
}

// IsMember returns true if the member is a member of the Set for the given key.
func (s *ShardedSet) IsMember(key string, member string) bool {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return false
	}
	if set.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Set) bool { return cur == set })
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
		sh.DeleteIf(key, func(cur *vals.Set) bool { return cur == set })
		return 0, false
	}
	return set.Card(), true
}

// Members returns the members of the Set for the given key.
// Returns nil and false if the Set does not exist or is expired.
func (s *ShardedSet) Members(key string) ([]string, bool) {
	sh := s.getShard(key)
	set, ok := sh.Get(key)
	if !ok {
		return nil, false
	}
	if set.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Set) bool { return cur == set })
		return nil, false
	}
	return set.Members(), true
}

// Expire sets the expiration time for the Set of the given key.
// Returns StatusSuccess if expiration was set, StatusInvalidTTL if ttl is invalid or overflows,
// or StatusNotFound/StatusExpired if the Set does not exist or is already expired.
func (s *ShardedSet) Expire(key string, ttl int64) shard.Status {
	if !vals.IsValidTTL(ttl) {
		return shard.StatusInvalidTTL
	}
	now := vals.MonotonicNow()
	return s.getShard(key).ExpireInPlace(
		key,
		func(set *vals.Set) bool { return set.IsExpiredAt(now) },
		func(set *vals.Set) bool { return set.ExpireAt(now, ttl) },
	)
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
	ttl := set.TTL()
	if ttl == -2 {
		sh.DeleteIf(key, func(cur *vals.Set) bool { return cur == set })
		return -2
	}
	return ttl
}

// CleanExpired removes all expired sets across all shards.
func (s *ShardedSet) CleanExpired() {
	s.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes all expired sets across all shards while respecting ctx cancellation.
func (s *ShardedSet) CleanExpiredCtx(ctx context.Context) {
	for _, sh := range s.shards {
		if ctx.Err() != nil {
			return
		}
		now := vals.MonotonicNow()
		sh.CleanCtx(ctx, func(_ string, set *vals.Set) bool {
			return set.IsExpiredAt(now)
		})
	}
}

// Flush removes all sets across all shards.
func (s *ShardedSet) Flush() {
	for _, sh := range s.shards {
		sh.Flush()
	}
}
