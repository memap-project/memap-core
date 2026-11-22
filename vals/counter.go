package vals

import (
	"math"
	"sync/atomic"
)

// Counter is a thread-safe non-negative integer counter with optional limit and monotonic expiration.
// If limit is 0, the Counter has no upper limit.
type Counter struct {
	expiry
	value atomic.Int64
	limit atomic.Int64
}

// NewCounter creates a new Counter.
func NewCounter() *Counter {
	return &Counter{}
}

// SetLimit sets or updates the upper limit of the Counter.
// Returns false if limit is negative.
func (c *Counter) SetLimit(limit int64) bool {
	if limit < 0 {
		return false
	}
	c.limit.Store(limit)
	return true
}

// GetLimit returns the upper limit of the Counter.
func (c *Counter) GetLimit() int64 {
	return c.limit.Load()
}

// GetValue returns the current value of the Counter.
func (c *Counter) GetValue() int64 {
	return c.value.Load()
}

// IncrBy atomically increments the Counter value by alpha (>= 0) up to limit if set.
// Returns the new value and true if the increment succeeded.
// Returns 0 and false if alpha < 0, the addition would overflow int64, or it would exceed limit.
func (c *Counter) IncrBy(alpha int64) (int64, bool) {
	if alpha < 0 {
		return 0, false
	}
	for {
		val := c.value.Load()
		lim := c.limit.Load()
		if alpha > math.MaxInt64-val {
			return 0, false
		}
		next := val + alpha
		if lim > 0 && next > lim {
			return 0, false
		}
		if c.value.CompareAndSwap(val, next) {
			return next, true
		}
	}
}

// DecrBy atomically decrements the Counter value by alpha (>= 0) down to 0.
// Returns the new value and true if the decrement succeeded.
// Returns 0 and false if alpha < 0 or the decrement would result in a negative value.
func (c *Counter) DecrBy(alpha int64) (int64, bool) {
	if alpha < 0 {
		return 0, false
	}
	for {
		val := c.value.Load()
		if alpha > val {
			return 0, false
		}
		next := val - alpha
		if c.value.CompareAndSwap(val, next) {
			return next, true
		}
	}
}

// Reset resets the Counter value to 0.
func (c *Counter) Reset() {
	c.value.Store(0)
}
