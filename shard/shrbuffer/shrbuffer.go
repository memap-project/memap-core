package shrbuffer

import (
	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedRingBuffer is a partitioned map storing ring buffers across multiple shards.
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
func (shrb *ShardedRingBuffer) getShard(key string) *shard.Shard[*vals.RingBuffer[string]] {
	return shrb.shards[shard.ShardIndex(key, shrb.shardCount)]
}

// Init initializes a new ring buffer with the given key, capacity, and optional TTL.
// Returns true if the ring buffer was created.
// Returns false if a ring buffer already exists for the given key.
func (shrb *ShardedRingBuffer) Init(key string, cap, ttl int64) bool {
	rb, ok := shrb.getShard(key).GetOrInit(key, func() *vals.RingBuffer[string] {
		return vals.NewRingBuffer[string](cap)
	})
	if ok {
		return false
	}
	rb.Expire(ttl)
	return true
}

// Push adds a value to the ring buffer for the given key.
// Returns false if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) Push(key, value string) bool {
	sh := shrb.getShard(key)
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
func (shrb *ShardedRingBuffer) Pop(key string) (string, shard.Status) {
	sh := shrb.getShard(key)
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
func (shrb *ShardedRingBuffer) At(key string, index int64) (string, shard.Status) {
	sh := shrb.getShard(key)
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
// Returns empty slice and false if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) Slice(key string) ([]string, bool) {
	sh := shrb.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return []string{}, false
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return []string{}, false
	}
	return rb.Slice(), true
}

// Peek returns the oldest value at the head of the ring buffer without removing it.
// Returns value and StatusSuccess if found.
// Returns empty string and failure Status if the ring buffer does not exist, is expired, or is empty.
func (shrb *ShardedRingBuffer) Peek(key string) (string, shard.Status) {
	sh := shrb.getShard(key)
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
func (shrb *ShardedRingBuffer) Back(key string) (string, shard.Status) {
	sh := shrb.getShard(key)
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

// Cap returns the capacity of the ring buffer for the given key.
// Returns 0 and false if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) Cap(key string) (int64, bool) {
	sh := shrb.getShard(key)
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

// Len returns the number of elements in the ring buffer for the given key.
// Returns 0 and false if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) Len(key string) (int64, bool) {
	sh := shrb.getShard(key)
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

// Reset resets the ring buffer for the given key to an empty state.
// Returns false if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) Reset(key string) bool {
	sh := shrb.getShard(key)
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

// Delete removes the ring buffer for the given key.
func (shrb *ShardedRingBuffer) Delete(key string) {
	shrb.getShard(key).Delete(key)
}

// Expire sets the expiration time for the ring buffer of the given key.
// Returns true if the ring buffer exists and expiration was set.
// Returns false if the ring buffer does not exist or is already expired.
func (shrb *ShardedRingBuffer) Expire(key string, ttl int64) bool {
	ok := shrb.getShard(key).Update(key, func(rb *vals.RingBuffer[string]) bool {
		if rb.IsExpired() {
			return false
		}
		rb.Expire(ttl)
		return true
	})
	return ok
}

// TTL returns the remaining time-to-live of the ring buffer for the given key in seconds.
// Returns -1 if the ring buffer exists and has no expiration time.
// Returns -2 if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) TTL(key string) int64 {
	sh := shrb.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return -2
	}
	if rb.IsExpired() {
		sh.Delete(key)
		return -2
	}
	if rb.TTL() == 0 {
		return -1
	}
	return rb.TTL()
}

// CleanExpired removes all expired ring buffers across all shards.
func (shrb *ShardedRingBuffer) CleanExpired() {
	for _, shard := range shrb.shards {
		shard.Clean(func(key string, rb *vals.RingBuffer[string]) bool {
			return rb.IsExpired()
		})
	}
}

// Flush removes all ring buffers across all shards.
func (shrb *ShardedRingBuffer) Flush() {
	for _, shard := range shrb.shards {
		shard.Flush()
	}
}
