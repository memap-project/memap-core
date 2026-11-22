package shhash

import (
	"context"
	"hash/maphash"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedHash is a partitioned map storing hash objects across multiple shards.
type ShardedHash struct {
	shardCount uint8
	seed       maphash.Seed
	mask       uint64
	shards     []*shard.Shard[*vals.Hash]
}

// NewShardedHash creates a new ShardedHash with the given shard count and optional per-shard key limit.
// shardCount is normalized to a power of two in [1, 128].
func NewShardedHash(shardCount uint8, maxKeysPerShard ...int) *ShardedHash {
	shardCount = shard.NormalizeShardCount(shardCount)
	var limit int
	if len(maxKeysPerShard) > 0 && maxKeysPerShard[0] > 0 {
		limit = maxKeysPerShard[0]
	}
	shards := make([]*shard.Shard[*vals.Hash], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Hash](limit)
	}
	return &ShardedHash{
		shardCount: shardCount,
		seed:       maphash.MakeSeed(),
		mask:       uint64(shardCount - 1),
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key using a per-instance seeded hash.
func (shh *ShardedHash) getShard(key string) *shard.Shard[*vals.Hash] {
	return shh.shards[shard.SeededShardIndex(shh.seed, key, shh.mask)]
}

// Get retrieves a copy of all field-value pairs for the given key.
// Returns nil and false if the key does not exist or is expired.
func (shh *ShardedHash) Get(key string) (map[string]string, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return nil, false
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return nil, false
	}
	return hash.GetCopy(), true
}

// Set creates a new empty hash for the given key with optional TTL, overwriting any existing hash.
func (shh *ShardedHash) Set(key string, ttl int64) {
	h := vals.NewHash()
	if ttl > 0 {
		h.Expire(ttl)
	}
	shh.getShard(key).SetWithEvictCheck(key, h, (*vals.Hash).IsExpired)
}

// Delete removes the hash for the given key.
func (shh *ShardedHash) Delete(key string) {
	shh.getShard(key).Delete(key)
}

// Expire sets the expiration time for the hash of the given key.
// Returns StatusSuccess if expiration was set, StatusInvalidTTL if ttl is invalid or overflows,
// or StatusNotFound/StatusExpired if the hash does not exist or is already expired.
func (shh *ShardedHash) Expire(key string, ttl int64) shard.Status {
	if !vals.IsValidTTL(ttl) {
		return shard.StatusInvalidTTL
	}
	now := vals.MonotonicNow()
	return shh.getShard(key).ExpireInPlace(
		key,
		func(h *vals.Hash) bool { return h.IsExpiredAt(now) },
		func(h *vals.Hash) bool { return h.ExpireAt(now, ttl) },
	)
}

// TTL returns the time-to-live of the hash for the given key in seconds.
// Returns -1 if the hash exists and has no expiration time.
// Returns -2 if the hash does not exist or is expired.
func (shh *ShardedHash) TTL(key string) int64 {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return -2
	}
	ttl := hash.TTL()
	if ttl == -2 {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return -2
	}
	return ttl
}

// Exists returns true if an unexpired hash exists for the given key.
func (shh *ShardedHash) Exists(key string) bool {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return false
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return false
	}
	return true
}

// Len returns the number of fields in the hash for the given key.
// Returns 0 and false if the hash does not exist or is expired.
func (shh *ShardedHash) Len(key string) (int64, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return 0, false
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return 0, false
	}
	return hash.Len(), true
}

// Keys returns all field names in the hash for the given key.
// Returns nil and false if the hash does not exist or is expired.
func (shh *ShardedHash) Keys(key string) ([]string, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return nil, false
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return nil, false
	}
	return hash.Keys(), true
}

// Values returns all field values in the hash for the given key.
// Returns nil and false if the hash does not exist or is expired.
func (shh *ShardedHash) Values(key string) ([]string, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return nil, false
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return nil, false
	}
	return hash.Values(), true
}

// GetField retrieves the value of the specified field from the hash for the given key.
// Returns value and StatusSuccess if found.
// Returns empty string and failure Status if the hash or field does not exist, or if the hash is expired.
func (shh *ShardedHash) GetField(key string, field string) (string, shard.Status) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return "", shard.StatusNotFound
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return "", shard.StatusExpired
	}
	f, ok := hash.Get(field)
	if !ok {
		return "", shard.StatusFieldNotFound
	}
	return f, shard.StatusSuccess
}

// SetField sets or updates a field in the hash for the given key.
// Atomically creates or replaces the hash if one does not exist or is expired.
func (shh *ShardedHash) SetField(key, field, value string) {
	sh := shh.getShard(key)
	hash, _ := sh.GetOrReplace(key, (*vals.Hash).IsExpired, vals.NewHash)
	hash.Set(field, value)
}

// DeleteField removes a field from the hash for the given key.
// Lazily deletes the hash entry if it is already expired.
func (shh *ShardedHash) DeleteField(key, field string) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return
	}
	if hash.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.Hash) bool { return cur == hash })
		return
	}
	hash.Delete(field)
}

// CleanExpired removes all expired hashes across all shards.
func (shh *ShardedHash) CleanExpired() {
	shh.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes all expired hashes across all shards while respecting ctx cancellation.
func (shh *ShardedHash) CleanExpiredCtx(ctx context.Context) {
	for _, sh := range shh.shards {
		if ctx.Err() != nil {
			return
		}
		now := vals.MonotonicNow()
		sh.CleanCtx(ctx, func(_ string, h *vals.Hash) bool {
			return h.IsExpiredAt(now)
		})
	}
}

// Flush removes all hashes across all shards.
func (shh *ShardedHash) Flush() {
	for _, sh := range shh.shards {
		sh.Flush()
	}
}
