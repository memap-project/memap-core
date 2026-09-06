package shhash

import (
	"hash/fnv"

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
func (s *ShardedHash) getShard(key string) *shard.Shard[*vals.Hash] {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() & uint32(s.shardCount-1)
	return s.shards[idx]
}

// Get retrieves a copy of all field-value pairs for the given key.
// Returns empty map and false if the key does not exist or is expired.
func (s *ShardedHash) Get(key string) (map[string]string, bool) {
	h, ok := s.getShard(key).Get(key)
	if !ok {
		return map[string]string{}, false
	}
	if h.IsExpired() {
		s.getShard(key).Delete(key)
		return map[string]string{}, false
	}
	return h.GetCopy(), true
}

// Set creates a new empty hash for the given key with optional TTL, overwriting any existing hash.
func (s *ShardedHash) Set(key string, ttl int64) bool {
	h := vals.NewHash()
	if ttl > 0 {
		h.Expire(ttl)
	}
	s.getShard(key).Set(key, h)
	return true
}

// Delete removes the hash for the given key.
func (s *ShardedHash) Delete(key string) {
	s.getShard(key).Delete(key)
}

// Expire sets the expiration time for the hash of the given key.
// Returns true if the hash exists and expiration was set.
// Returns false if the hash does not exist or is already expired.
func (s *ShardedHash) Expire(key string, ttl int64) bool {
	return s.getShard(key).Update(key, func(h *vals.Hash) bool {
		if h.IsExpired() {
			return false
		}
		h.Expire(ttl)
		return true
	})
}

// TTL returns the time-to-live of the hash for the given key in seconds.
// Returns -1 if the hash exists and has no expiration time.
// Returns -2 if the hash does not exist or is expired.
func (s *ShardedHash) TTL(key string) int64 {
	h, ok := s.getShard(key).Get(key)
	if !ok || h.IsExpired() {
		return -2
	}
	if h.TTL() == 0 {
		return -1
	}
	return h.TTL()
}

// Exists returns true if an unexpired hash exists for the given key.
func (s *ShardedHash) Exists(key string) bool {
	h, ok := s.getShard(key).Get(key)
	if !ok {
		return false
	}
	if h.IsExpired() {
		s.getShard(key).Delete(key)
		return false
	}
	return true
}

// Len returns the number of fields in the hash for the given key.
// Returns 0 and false if the hash does not exist or is expired.
func (s *ShardedHash) Len(key string) (int64, bool) {
	h, ok := s.getShard(key).Get(key)
	if !ok {
		return 0, false
	}
	if h.IsExpired() {
		s.getShard(key).Delete(key)
		return 0, false
	}
	return h.Len(), true
}

// Keys returns all field names in the hash for the given key.
// Returns empty slice and false if the hash does not exist or is expired.
func (s *ShardedHash) Keys(key string) ([]string, bool) {
	h, ok := s.getShard(key).Get(key)
	if !ok {
		return []string{}, false
	}
	if h.IsExpired() {
		s.getShard(key).Delete(key)
		return []string{}, false
	}
	return h.Keys(), true
}

// Values returns all field values in the hash for the given key.
// Returns empty slice and false if the hash does not exist or is expired.
func (s *ShardedHash) Values(key string) ([]string, bool) {
	h, ok := s.getShard(key).Get(key)
	if !ok {
		return []string{}, false
	}
	if h.IsExpired() {
		s.getShard(key).Delete(key)
		return []string{}, false
	}
	return h.Values(), true
}

// GetField retrieves the value of the specified field from the hash for the given key.
// Returns value and StatusSuccess if found.
// Returns empty string and failure Status if the hash or field does not exist, or if the hash is expired.
func (s *ShardedHash) GetField(key string, field string) (string, shard.Status) {
	sh := s.getShard(key)
	h, ok := sh.Get(key)
	if !ok {
		return "", shard.StatusNotFound
	}
	if h.IsExpired() {
		sh.Delete(key)
		return "", shard.StatusExpired
	}
	v, ok := h.Get(field)
	if !ok {
		return "", shard.StatusFieldNotFound
	}
	return v, shard.StatusSuccess
}

// SetField sets or updates a field in the hash for the given key. Creates a new hash if one does not exist.
func (s *ShardedHash) SetField(key, field, value string) bool {
	h, _ := s.getShard(key).GetOrInit(key, vals.NewHash)
	if h.IsExpired() {
		h = vals.NewHash()
		s.getShard(key).Set(key, h)
	}
	h.Set(field, value)
	return true
}

// DeleteField removes a field from the hash for the given key.
func (s *ShardedHash) DeleteField(key, field string) {
	h, ok := s.getShard(key).Get(key)
	if ok {
		h.Delete(field)
	}
}

// CleanExpired removes all expired hashes across all shards.
func (s *ShardedHash) CleanExpired() {
	for _, shard := range s.shards {
		shard.Clean(func(_ string, h *vals.Hash) bool {
			return h.IsExpired()
		})
	}
}

// Flush removes all hashes across all shards.
func (s *ShardedHash) Flush() {
	for _, shard := range s.shards {
		shard.Flush()
	}
}
