package vals

import (
	"errors"
	"math"
	"sync"
)

// DefaultMaxRingBufferCap is the default upper bound for a RingBuffer capacity.
const DefaultMaxRingBufferCap int64 = 1 << 20

// MaxSafeRingBufferCap is the hard upper bound on ring buffer capacity to avoid makeslice panics.
const MaxSafeRingBufferCap int64 = math.MaxInt / 16

// ErrInvalidCapacity is returned when a ring buffer capacity is out of valid bounds.
var ErrInvalidCapacity = errors.New("ring buffer capacity must be between 1 and max capacity")

// NormalizeMaxRingBufferCap returns a safe positive upper bound for ring buffer capacity.
func NormalizeMaxRingBufferCap(maxCap int64) int64 {
	if maxCap <= 0 {
		return DefaultMaxRingBufferCap
	}
	if maxCap > MaxSafeRingBufferCap {
		return MaxSafeRingBufferCap
	}
	return maxCap
}

// intstr represents types supported by RingBuffer (int64 or string).
type intstr interface {
	int64 | string
}

// RingBuffer is a thread-safe circular buffer with fixed capacity and optional monotonic expiration.
type RingBuffer[T intstr] struct {
	expiry
	mu   sync.RWMutex
	buf  []T
	cap  int64
	len  int64
	head int64
	tail int64
}

// NewRingBufferWithMax creates a new RingBuffer after validating that 1 <= cap <= maxCap.
// If maxCap <= 0, DefaultMaxRingBufferCap is used.
func NewRingBufferWithMax[T intstr](cap, maxCap int64) (*RingBuffer[T], error) {
	maxCap = NormalizeMaxRingBufferCap(maxCap)
	if cap <= 0 || cap > maxCap {
		return nil, ErrInvalidCapacity
	}
	return &RingBuffer[T]{
		mu:   sync.RWMutex{},
		buf:  make([]T, cap),
		cap:  cap,
		len:  0,
		head: 0,
		tail: 0,
	}, nil
}

// NewRingBuffer creates a new RingBuffer with the given capacity.
// Non-positive capacities produce a safe zero-capacity buffer (Push is a no-op).
// Capacities exceeding DefaultMaxRingBufferCap are capped to prevent OOM panics;
// use NewRingBufferWithMax when explicit validation or custom limits are required.
func NewRingBuffer[T intstr](cap int64) *RingBuffer[T] {
	if cap <= 0 {
		cap = 0
	} else if cap > DefaultMaxRingBufferCap {
		cap = DefaultMaxRingBufferCap
	}
	return &RingBuffer[T]{
		mu:   sync.RWMutex{},
		buf:  make([]T, cap),
		cap:  cap,
		len:  0,
		head: 0,
		tail: 0,
	}
}

// Push adds a value to the RingBuffer, overwriting the oldest element if full.
func (rb *RingBuffer[T]) Push(val T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.cap <= 0 {
		return
	}
	rb.buf[rb.tail] = val
	rb.tail = (rb.tail + 1) % rb.cap
	if rb.len < rb.cap {
		rb.len++
	} else {
		rb.head = (rb.head + 1) % rb.cap
	}
}

// Pop removes and returns the value at the head of the RingBuffer.
// Clears the popped slot so any referenced backing memory can be garbage collected.
// Returns zero value and false if the RingBuffer is empty.
func (rb *RingBuffer[T]) Pop() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	var zero T
	if rb.len == 0 {
		return zero, false
	}
	val := rb.buf[rb.head]
	rb.buf[rb.head] = zero
	rb.head = (rb.head + 1) % rb.cap
	rb.len--
	return val, true
}

// At returns the value at the specified logical index (0 is oldest).
// Returns zero value and false if index is out of range.
func (rb *RingBuffer[T]) At(index int64) (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	var zero T
	if index < 0 || index >= rb.len {
		return zero, false
	}
	val := rb.buf[(rb.head+index)%rb.cap]
	return val, true
}

// Slice retrieves a copy of all current elements in logical order (oldest to newest).
// Returns nil if the RingBuffer is empty.
func (rb *RingBuffer[T]) Slice() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	result := make([]T, rb.len)
	if rb.len == 0 {
		return result
	}
	firstPart := min(rb.cap-rb.head, rb.len)
	copy(result, rb.buf[rb.head:rb.head+firstPart])
	if secondPart := rb.len - firstPart; secondPart > 0 {
		copy(result[firstPart:], rb.buf[:secondPart])
	}
	return result
}

// Peek returns the value at the head (oldest) of the RingBuffer without removing it.
// Returns zero value and false if the RingBuffer is empty.
func (rb *RingBuffer[T]) Peek() (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	var zero T
	if rb.len == 0 {
		return zero, false
	}
	val := rb.buf[rb.head]
	return val, true
}

// Back returns the value at the tail (newest) of the RingBuffer without removing it.
// Returns zero value and false if the RingBuffer is empty.
func (rb *RingBuffer[T]) Back() (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	var zero T
	if rb.len == 0 {
		return zero, false
	}
	idx := (rb.tail - 1 + rb.cap) % rb.cap
	val := rb.buf[idx]
	return val, true
}

// Cap returns the capacity of the RingBuffer.
func (rb *RingBuffer[T]) Cap() int64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.cap
}

// Len returns the number of elements currently stored in the RingBuffer.
func (rb *RingBuffer[T]) Len() int64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.len
}

// IsEmpty returns true if the RingBuffer has no elements.
func (rb *RingBuffer[T]) IsEmpty() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.len == 0
}

// IsFull returns true if the RingBuffer is at full capacity.
func (rb *RingBuffer[T]) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.len == rb.cap
}

// Reset resets the RingBuffer to its initial empty state and clears all references
// in the backing slice so values can be garbage collected.
func (rb *RingBuffer[T]) Reset() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	clear(rb.buf)
	rb.len = 0
	rb.head = 0
	rb.tail = 0
}
