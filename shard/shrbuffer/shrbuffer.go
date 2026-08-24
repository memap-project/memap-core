package shrbuffer

import (
	"hash/fnv"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedRingBuffer is a partitioned storage for ring buffers across multiple shards.
type ShardedRingBuffer struct {
	shardCount uint8
	shards     []*shard.Shard[*vals.RingBuffer[string]]
}

// NewShardedRingBuffer creates a new ShardedRingBuffer with the given shard count.
func NewShardedRingBuffer(shardCount uint8) *ShardedRingBuffer {
	shards := make([]*shard.Shard[*vals.RingBuffer[string]], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.RingBuffer[string]]()
	}
	return &ShardedRingBuffer{
		shardCount: shardCount,
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key based on FNV-1a hash.
func (s *ShardedRingBuffer) getShard(key string) *shard.Shard[*vals.RingBuffer[string]] {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() & uint32(s.shardCount-1)
	return s.shards[idx]
}

// Init initializes a new ring buffer with the given key, capacity, and optional TTL.
// Returns true if the ring buffer was created.
// Returns false if a ring buffer already exists for the given key.
func (s *ShardedRingBuffer) Init(key string, cap, ttl int64) bool {
	rb, ok := s.getShard(key).GetOrInit(key, func() *vals.RingBuffer[string] {
		return vals.NewRingBuffer[string](cap)
	})
	if ok {
		return false
	}
	rb.Expire(ttl)
	return true
}

// Push adds a value to the ring buffer with the given key.
// Returns false if the ring buffer does not exist or is expired.
func (s *ShardedRingBuffer) Push(key, value string) bool {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return false
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return false
	}
	rb.Push(value)
	return true
}

// Pop removes and returns the value at the head of the ring buffer.
// Returns value and StatusSuccess if popped.
// Returns empty string and failure Status if the ring buffer does not exist, is expired, or is empty.
func (s *ShardedRingBuffer) Pop(key string) (string, shard.Status) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return "", shard.StatusNotFound
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return "", shard.StatusExpired
	}
	if val, ok := rb.Pop(); ok {
		return val, shard.StatusSuccess
	}
	return "", shard.StatusBufferEmpty
}

// At returns the value at the specified index from the ring buffer.
// Returns value and StatusSuccess if found.
// Returns empty string and failure Status if the ring buffer does not exist, is expired, or index is out of bounds.
func (s *ShardedRingBuffer) At(key string, index int64) (string, shard.Status) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return "", shard.StatusNotFound
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return "", shard.StatusExpired
	}
	if val, ok := rb.At(index); ok {
		return val, shard.StatusSuccess
	}
	return "", shard.StatusIndexOutOfBounds
}

// Slice returns all elements currently in the ring buffer in logical order (oldest to newest).
// Returns nil and false if the ring buffer does not exist or is expired.
func (s *ShardedRingBuffer) Slice(key string) ([]string, bool) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return nil, false
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return nil, false
	}
	return rb.Slice(), true
}

// Peek returns the oldest value at the head of the ring buffer without removing it.
// Returns value and StatusSuccess if found.
// Returns empty string and failure Status if the ring buffer does not exist, is expired, or is empty.
func (s *ShardedRingBuffer) Peek(key string) (string, shard.Status) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return "", shard.StatusNotFound
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return "", shard.StatusExpired
	}
	if value, ok := rb.Peek(); ok {
		return value, shard.StatusSuccess
	}
	return "", shard.StatusBufferEmpty
}

// Back returns the newest value at the tail of the ring buffer without removing it.
// Returns value and StatusSuccess if found.
// Returns empty string and failure Status if the ring buffer does not exist, is expired, or is empty.
func (s *ShardedRingBuffer) Back(key string) (string, shard.Status) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return "", shard.StatusNotFound
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return "", shard.StatusExpired
	}
	if value, ok := rb.Back(); ok {
		return value, shard.StatusSuccess
	}
	return "", shard.StatusBufferEmpty
}

// Cap returns the capacity of the ring buffer with the given key.
// Returns 0 and false if the ring buffer does not exist or is expired.
func (s *ShardedRingBuffer) Cap(key string) (int64, bool) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return 0, false
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return 0, false
	}
	return rb.Cap(), true
}

// Len returns the number of elements in the ring buffer with the given key.
// Returns 0 and false if the ring buffer does not exist or is expired.
func (s *ShardedRingBuffer) Len(key string) (int64, bool) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return 0, false
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return 0, false
	}
	return rb.Len(), true
}

// Reset resets the ring buffer with the given key to an empty state.
// Returns false if the ring buffer does not exist or is expired.
func (s *ShardedRingBuffer) Reset(key string) bool {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return false
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return false
	}
	rb.Reset()
	return true
}

// Delete removes the ring buffer with the given key.
func (s *ShardedRingBuffer) Delete(key string) {
	s.getShard(key).Delete(key)
}

// Expire sets the expiration time for the ring buffer of the given key.
// Returns true if the ring buffer exists and expiration was set.
// Returns false if the ring buffer does not exist or is already expired.
func (s *ShardedRingBuffer) Expire(key string, ttl int64) bool {
	return s.getShard(key).Update(key, func(rb *vals.RingBuffer[string]) bool {
		if rb.IsExpired() {
			return false
		}
		return rb.Expire(ttl)
	})
}

// TTL returns the remaining time-to-live of the ring buffer in seconds.
// Returns -1 and true if the ring buffer exists and has no expiration time.
// Returns -2 and false if the ring buffer does not exist or is expired.
func (s *ShardedRingBuffer) TTL(key string) (int64, bool) {
	sh := s.getShard(key)
	rb, ok := sh.Get(key)
	if !ok || rb.IsExpired() {
		return -2, false
	}
	if rb.TTL() == 0 {
		return -1, true
	}
	return rb.TTL(), true
}

// CleanExpired removes all expired ring buffers across all shards.
func (s *ShardedRingBuffer) CleanExpired() {
	for _, shard := range s.shards {
		shard.Clean(func(key string, rb *vals.RingBuffer[string]) bool {
			return rb.IsExpired()
		})
	}
}

// Flush removes all ring buffers across all shards.
func (s *ShardedRingBuffer) Flush() {
	for _, shard := range s.shards {
		shard.Flush()
	}
}
