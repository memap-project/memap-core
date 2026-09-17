package shhash

import (
	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedHash is a partitioned map storing hash objects across multiple shards.
type ShardedHash struct {
	shardCount uint8
	shards     []*shard.Shard[*vals.Hash]
}

// NewShardedHash creates a new ShardedHash with the given shard count.
func NewShardedHash(shardCount uint8) *ShardedHash {
	shards := make([]*shard.Shard[*vals.Hash], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.Hash]()
	}
	return &ShardedHash{
		shardCount: shardCount,
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key based on FNV-1a hash.
func (shh *ShardedHash) getShard(key string) *shard.Shard[*vals.Hash] {
	return shh.shards[shard.ShardIndex(key, shh.shardCount)]
}

// Get retrieves a copy of all field-value pairs for the given key.
// Returns empty map and false if the key does not exist or is expired.
func (shh *ShardedHash) Get(key string) (map[string]string, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return map[string]string{}, false
	}
	if hash.IsExpired() {
		shh.getShard(key).Delete(key)
		return map[string]string{}, false
	}
	return hash.GetCopy(), true
}

// Set creates a new empty hash for the given key with optional TTL, overwriting any existing hash.
func (shh *ShardedHash) Set(key string, ttl int64) {
	sh := shh.getShard(key)
	hash, _ := sh.GetOrInit(key, func() *vals.Hash {
		return vals.NewHash()
	})
	if hash.IsExpired() {
		hash = vals.NewHash()
		sh.Set(key, hash)
	}
	hash.Expire(ttl)
}

// Delete removes the hash for the given key.
func (s *ShardedHash) Delete(key string) {
	s.getShard(key).Delete(key)
}

// Expire sets the expiration time for the hash of the given key.
// Returns true if the hash exists and expiration was set.
// Returns false if the hash does not exist or is already expired.
func (shh *ShardedHash) Expire(key string, ttl int64) bool {
	ok := shh.getShard(key).Update(key, func(h *vals.Hash) bool {
		if h.IsExpired() {
			return false
		}
		h.Expire(ttl)
		return true
	})
	return ok
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
	if hash.IsExpired() {
		sh.Delete(key)
		return -2
	}
	if hash.TTL() == 0 {
		return -1
	}
	return hash.TTL()
}

// Exists returns true if an unexpired hash exists for the given key.
func (shh *ShardedHash) Exists(key string) bool {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return false
	}
	if hash.IsExpired() {
		sh.Delete(key)
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
		sh.Delete(key)
		return 0, false
	}
	return hash.Len(), true
}

// Keys returns all field names in the hash for the given key.
// Returns empty slice and false if the hash does not exist or is expired.
func (shh *ShardedHash) Keys(key string) ([]string, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return []string{}, false
	}
	if hash.IsExpired() {
		sh.Delete(key)
		return []string{}, false
	}
	return hash.Keys(), true
}

// Values returns all field values in the hash for the given key.
// Returns empty slice and false if the hash does not exist or is expired.
func (shh *ShardedHash) Values(key string) ([]string, bool) {
	sh := shh.getShard(key)
	hash, ok := sh.Get(key)
	if !ok {
		return []string{}, false
	}
	if hash.IsExpired() {
		sh.Delete(key)
		return []string{}, false
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
		sh.Delete(key)
		return "", shard.StatusExpired
	}
	f, ok := hash.Get(field)
	if !ok {
		return "", shard.StatusFieldNotFound
	}
	return f, shard.StatusSuccess
}

// SetField sets or updates a field in the hash for the given key. Creates a new hash if one does not exist.
func (shh *ShardedHash) SetField(key, field, value string) {
	sh := shh.getShard(key)
	hash, _ := sh.GetOrInit(key, vals.NewHash)
	if hash.IsExpired() {
		hash = vals.NewHash()
		sh.Set(key, hash)
	}
	hash.Set(field, value)
}

// DeleteField removes a field from the hash for the given key.
func (shh *ShardedHash) DeleteField(key, field string) {
	hash, ok := shh.getShard(key).Get(key)
	if ok {
		hash.Delete(field)
	}
}

// CleanExpired removes all expired hashes across all shards.
func (shh *ShardedHash) CleanExpired() {
	for _, sh := range shh.shards {
		sh.Clean(func(_ string, h *vals.Hash) bool {
			return h.IsExpired()
		})
	}
}

// Flush removes all hashes across all shards.
func (shh *ShardedHash) Flush() {
	for _, sh := range shh.shards {
		sh.Flush()
	}
}
