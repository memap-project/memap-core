package shrbuffer

import (
	"context"
	"hash/maphash"

	"github.com/memap-project/memap-core/shard"
	"github.com/memap-project/memap-core/vals"
)

// ShardedRingBuffer is a partitioned map storing ring buffers across multiple shards.
type ShardedRingBuffer struct {
	shardCount uint8
	seed       maphash.Seed
	mask       uint64
	maxCap     int64
	shards     []*shard.Shard[*vals.RingBuffer[string]]
}

// NewShardedRingBuffer creates a new ShardedRingBuffer with the given shard count and optional max capacity per buffer.
// shardCount is normalized to a power of two in [1, 128].
// If maxCap is omitted or <= 0, vals.DefaultMaxRingBufferCap is used.
func NewShardedRingBuffer(shardCount uint8, maxCap ...int64) *ShardedRingBuffer {
	var capLimit int64
	if len(maxCap) > 0 {
		capLimit = maxCap[0]
	}
	return NewShardedRingBufferWithLimits(shardCount, capLimit, 0)
}

// NewShardedRingBufferWithLimits creates a new ShardedRingBuffer with explicit per-buffer capacity
// and per-shard key limits.
func NewShardedRingBufferWithLimits(shardCount uint8, maxCap int64, maxKeysPerShard int) *ShardedRingBuffer {
	shardCount = shard.NormalizeShardCount(shardCount)
	limit := vals.NormalizeMaxRingBufferCap(maxCap)
	shards := make([]*shard.Shard[*vals.RingBuffer[string]], shardCount)
	for i := range shards {
		shards[i] = shard.NewShard[*vals.RingBuffer[string]](maxKeysPerShard)
	}
	return &ShardedRingBuffer{
		shardCount: shardCount,
		seed:       maphash.MakeSeed(),
		mask:       uint64(shardCount - 1),
		maxCap:     limit,
		shards:     shards,
	}
}

// getShard returns the shard corresponding to the given key using a per-instance seeded hash.
func (shrb *ShardedRingBuffer) getShard(key string) *shard.Shard[*vals.RingBuffer[string]] {
	return shrb.shards[shard.SeededShardIndex(shrb.seed, key, shrb.mask)]
}

// Init initializes a new ring buffer with the given key, capacity, and optional TTL.
// Atomically replaces an expired ring buffer if present.
// Returns StatusSuccess if the ring buffer was created.
// Returns StatusInvalidCapacity if capacity is <= 0 or exceeds the configured max capacity.
// Returns StatusAlreadyExists if an unexpired ring buffer already exists for the given key.
func (shrb *ShardedRingBuffer) Init(key string, cap, ttl int64) shard.Status {
	if cap <= 0 || cap > shrb.maxCap {
		return shard.StatusInvalidCapacity
	}
	_, existed := shrb.getShard(key).GetOrReplace(
		key,
		(*vals.RingBuffer[string]).IsExpired,
		func() *vals.RingBuffer[string] {
			rb, err := vals.NewRingBufferWithMax[string](cap, shrb.maxCap)
			if err != nil {
				panic(err)
			}
			if ttl > 0 {
				rb.Expire(ttl)
			}
			return rb
		},
	)
	if existed {
		return shard.StatusAlreadyExists
	}
	return shard.StatusSuccess
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
		return "", shard.StatusExpired
	}
	if val, ok := rb.At(index); ok {
		return val, shard.StatusSuccess
	}
	return "", shard.StatusIndexOutOfBounds
}

// Slice returns all elements currently in the ring buffer in logical order (oldest to newest).
// Returns nil and false if the ring buffer does not exist or is expired.
func (shrb *ShardedRingBuffer) Slice(key string) ([]string, bool) {
	sh := shrb.getShard(key)
	rb, ok := sh.Get(key)
	if !ok {
		return nil, false
	}
	if rb.IsExpired() {
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
		return nil, false
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
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
// Returns StatusSuccess if expiration was set, StatusInvalidTTL if ttl is invalid or overflows,
// or StatusNotFound/StatusExpired if the ring buffer does not exist or is already expired.
func (shrb *ShardedRingBuffer) Expire(key string, ttl int64) shard.Status {
	if !vals.IsValidTTL(ttl) {
		return shard.StatusInvalidTTL
	}
	now := vals.MonotonicNow()
	return shrb.getShard(key).ExpireInPlace(
		key,
		func(rb *vals.RingBuffer[string]) bool { return rb.IsExpiredAt(now) },
		func(rb *vals.RingBuffer[string]) bool { return rb.ExpireAt(now, ttl) },
	)
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
	ttl := rb.TTL()
	if ttl == -2 {
		sh.DeleteIf(key, func(cur *vals.RingBuffer[string]) bool { return cur == rb })
		return -2
	}
	return ttl
}

// CleanExpired removes all expired ring buffers across all shards.
func (shrb *ShardedRingBuffer) CleanExpired() {
	shrb.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes all expired ring buffers across all shards while respecting ctx cancellation.
func (shrb *ShardedRingBuffer) CleanExpiredCtx(ctx context.Context) {
	for _, sh := range shrb.shards {
		if ctx.Err() != nil {
			return
		}
		now := vals.MonotonicNow()
		sh.CleanCtx(ctx, func(_ string, rb *vals.RingBuffer[string]) bool {
			return rb.IsExpiredAt(now)
		})
	}
}

// Flush removes all ring buffers across all shards.
func (shrb *ShardedRingBuffer) Flush() {
	for _, sh := range shrb.shards {
		sh.Flush()
	}
}
