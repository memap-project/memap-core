package shard

import (
	"context"
	"hash/maphash"
	"maps"
	"sync"
)

// FNV-1a 32-bit constants.
const (
	offset32 uint32 = 2166136261
	prime32  uint32 = 16777619
)

const (
	shrinkMinPeak    = 256
	cleanDeleteBatch = 256
	evictSampleSize  = 16
)

// HashKey computes the 32-bit FNV-1a hash of a string key.
func HashKey(key string) uint32 {
	h := uint32(offset32)
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= prime32
	}
	return h
}

// NormalizeShardCount converts any uint8 shardCount to a valid power of two in [1, 128].
// Values <= 1 become 1; values > 128 are capped at 128; other non-powers of two are rounded up.
func NormalizeShardCount(shardCount uint8) uint8 {
	if shardCount <= 1 {
		return 1
	}
	if shardCount&(shardCount-1) == 0 {
		return shardCount
	}
	if shardCount > 128 {
		return 128
	}
	var p uint8 = 1
	for p < shardCount {
		p <<= 1
	}
	return p
}

// ShardIndex returns the shard index for the given key and shardCount.
// shardCount must be a power of two.
func ShardIndex(key string, shardCount uint8) uint32 {
	if shardCount <= 1 {
		return 0
	}
	return HashKey(key) & uint32(shardCount-1)
}

// SeededShardIndex computes a randomized shard index using hash/maphash,
// preventing shard-pinning hash-flooding DoS attacks.
// mask must be shardCount - 1 where shardCount is a power of two.
func SeededShardIndex(seed maphash.Seed, key string, mask uint64) uint64 {
	if mask == 0 {
		return 0
	}
	return maphash.String(seed, key) & mask
}

// Shard is a generic thread-safe map around Go's built-in map with optional bounded
// capacity (sampled eviction) and automatic bucket shrinking.
type Shard[V any] struct {
	mu      sync.RWMutex
	items   map[string]V
	maxKeys int
	peakLen int
}

// NewShard creates a new Shard with optional maximum key limit (0 = unlimited).
func NewShard[V any](maxKeys ...int) *Shard[V] {
	var limit int
	if len(maxKeys) > 0 && maxKeys[0] > 0 {
		limit = maxKeys[0]
	}
	return &Shard[V]{
		mu:      sync.RWMutex{},
		items:   make(map[string]V, 8),
		maxKeys: limit,
	}
}

// evictOneLocked removes a single entry when the shard is at capacity.
// Caller must hold s.mu.Lock(). Prefers expired entries if isExpired is provided.
func (s *Shard[V]) evictOneLocked(isExpired func(V) bool) {
	if s.maxKeys <= 0 || len(s.items) < s.maxKeys {
		return
	}
	var victim string
	var hasVictim bool
	checked := 0
	for k, v := range s.items {
		if !hasVictim {
			victim = k
			hasVictim = true
		}
		if isExpired != nil && isExpired(v) {
			delete(s.items, k)
			return
		}
		checked++
		if checked >= evictSampleSize {
			break
		}
	}
	if hasVictim {
		delete(s.items, victim)
	}
}

func (s *Shard[V]) maybeShrinkLocked() {
	if s.peakLen >= shrinkMinPeak && len(s.items)*4 <= s.peakLen {
		newCap := max(len(s.items), 8)
		shrunk := make(map[string]V, newCap)
		maps.Copy(shrunk, s.items)
		s.items = shrunk
		s.peakLen = len(s.items)
	}
}

// Get retrieves an item for the given key from the Shard.
// Returns zero value and false if the key does not exist.
func (s *Shard[V]) Get(key string) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	return value, true
}

// GetOrInit retrieves an item for the given key, or initializes and stores a new item if it does not exist.
// Returns the item and true if it existed, or the initialized item and false if it was newly created.
func (s *Shard[V]) GetOrInit(key string, initFn func() V) (V, bool) {
	s.mu.RLock()
	val, ok := s.items[key]
	s.mu.RUnlock()
	if ok {
		return val, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if val, ok = s.items[key]; ok {
		return val, true
	}
	s.evictOneLocked(nil)
	val = initFn()
	s.items[key] = val
	if len(s.items) > s.peakLen {
		s.peakLen = len(s.items)
	}
	return val, false
}

// GetOrReplace atomically returns the existing unexpired item, or initializes and stores
// a new item if the key is absent or expired.
// Returns (val, true) if a valid unexpired item already existed, or (val, false) if newly created/replaced.
func (s *Shard[V]) GetOrReplace(key string, isExpired func(V) bool, initFn func() V) (V, bool) {
	s.mu.RLock()
	val, ok := s.items[key]
	s.mu.RUnlock()
	if ok && !isExpired(val) {
		return val, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if val, ok = s.items[key]; ok && !isExpired(val) {
		return val, true
	}
	if !ok {
		s.evictOneLocked(isExpired)
	}
	val = initFn()
	s.items[key] = val
	if len(s.items) > s.peakLen {
		s.peakLen = len(s.items)
	}
	return val, false
}

// Set sets or updates the item for the given key in the Shard.
func (s *Shard[V]) Set(key string, value V) {
	s.SetWithEvictCheck(key, value, nil)
}

// SetWithEvictCheck sets or updates the item for the given key, preferring expired
// entries when sampled eviction is triggered by MaxKeysPerShard.
func (s *Shard[V]) SetWithEvictCheck(key string, value V, isExpired func(V) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[key]; !exists {
		s.evictOneLocked(isExpired)
	}
	s.items[key] = value
	if len(s.items) > s.peakLen {
		s.peakLen = len(s.items)
	}
}

// Delete removes the item for the given key from the Shard.
func (s *Shard[V]) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	s.maybeShrinkLocked()
}

// DeleteIf atomically deletes key only if it exists and match(currentVal) returns true.
// This prevents TOCTOU races where lazy expiration deletes a freshly overwritten key.
func (s *Shard[V]) DeleteIf(key string, match func(V) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.items[key]
	if !ok || !match(val) {
		return false
	}
	delete(s.items, key)
	s.maybeShrinkLocked()
	return true
}

// Update invokes fn on the item for key under write lock if it exists.
// Returns true if the item was found and fn returned true.
// Returns false if the item does not exist.
func (s *Shard[V]) Update(key string, fn func(val V) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.items[key]
	if !ok {
		return false
	}
	return fn(val)
}

// ModifyValue atomically inspects and optionally mutates or deletes a value-type entry
// under write lock without requiring heap-allocated pointer wrappers.
// If fn returns (updated, del, status): when del is true the key is removed;
// otherwise if updated is true the modified value is stored back into the map.
func (s *Shard[V]) ModifyValue(key string, fn func(val *V) (updated bool, del bool, status Status)) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.items[key]
	if !ok {
		return StatusNotFound
	}
	updated, del, status := fn(&val)
	if del {
		delete(s.items, key)
		s.maybeShrinkLocked()
		return status
	}
	if updated {
		s.items[key] = val
	}
	return status
}

// ExpireInPlace atomically checks existence and expiration under write lock,
// deleting the entry if already expired, or invoking expireFn if still alive.
func (s *Shard[V]) ExpireInPlace(key string, isExpired func(V) bool, expireFn func(V) bool) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.items[key]
	if !ok {
		return StatusNotFound
	}
	if isExpired(val) {
		delete(s.items, key)
		s.maybeShrinkLocked()
		return StatusExpired
	}
	if !expireFn(val) {
		return StatusExpired
	}
	return StatusSuccess
}

// Clean removes items from the Shard for which the predicate returns true.
func (s *Shard[V]) Clean(predicate func(key string, value V) bool) {
	s.CleanCtx(context.Background(), predicate)
}

// CleanCtx removes items for which predicate returns true using a two-phase batched strategy:
// Phase 1 scans under RLock (non-blocking for concurrent readers),
// Phase 2 deletes in small write-locked batches while re-verifying predicate (TOCTOU-safe).
// Respects context cancellation and shrinks the map if occupancy drops far below peak.
func (s *Shard[V]) CleanCtx(ctx context.Context, predicate func(key string, value V) bool) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return
	}

	var toDelete []string
	var needsShrink bool

	s.mu.RLock()
	if len(s.items) == 0 {
		needsShrink = s.peakLen >= shrinkMinPeak
		s.mu.RUnlock()
		if needsShrink {
			s.mu.Lock()
			s.maybeShrinkLocked()
			s.mu.Unlock()
		}
		return
	}

	scanned := 0
	for k, v := range s.items {
		scanned++
		if scanned&1023 == 0 && ctx.Err() != nil {
			s.mu.RUnlock()
			return
		}
		if predicate(k, v) {
			toDelete = append(toDelete, k)
		}
	}
	s.mu.RUnlock()

	if len(toDelete) == 0 {
		return
	}

	for start := 0; start < len(toDelete); start += cleanDeleteBatch {
		if ctx.Err() != nil {
			return
		}
		end := min(start+cleanDeleteBatch, len(toDelete))
		s.mu.Lock()
		for _, k := range toDelete[start:end] {
			if v, ok := s.items[k]; ok && predicate(k, v) {
				delete(s.items, k)
			}
		}
		if end == len(toDelete) {
			s.maybeShrinkLocked()
		}
		s.mu.Unlock()
	}
}

// Flush removes all items from the Shard.
func (s *Shard[V]) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]V, 8)
	s.peakLen = 0
}
