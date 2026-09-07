package vals

import (
	"sync/atomic"
	"time"
)

// Counter is a thread-safe integer counter with optional limit and expiration.
// If expiresAt is 0, the Counter has no expiration.
// If limit is 0, the Counter has no upper limit.
type Counter struct {
	value     atomic.Int64
	limit     atomic.Int64
	expiresAt atomic.Int64
}

// NewCounter creates a new Counter.
func NewCounter() *Counter {
	return &Counter{}
}

// IsExpired returns true if the Counter has an expiration time and is expired.
// Returns false if the Counter has no expiration time or is not yet expired.
func (c *Counter) IsExpired() bool {
	exp := c.expiresAt.Load()
	if exp == 0 {
		return false
	}
	return time.Now().Unix() > exp
}

// Expire sets the expiration time for the Counter.
// Returns false if the Counter is already expired or if ttl is negative.
func (c *Counter) Expire(ttl int64) bool {
	if c.IsExpired() {
		return false
	}
	if ttl < 0 {
		return false
	}
	c.expiresAt.Store(time.Now().Unix() + ttl)
	return true
}

// TTL returns the remaining time-to-live in seconds.
// Returns 0 if the Counter has no expiration time.
func (c *Counter) TTL() int64 {
	exp := c.expiresAt.Load()
	if exp == 0 {
		return 0
	}
	return exp - time.Now().Unix()
}

// SetLimit sets or updates the upper limit of the Counter.
func (c *Counter) SetLimit(limit int64) {
	c.limit.Store(limit)
}

// GetLimit returns the upper limit of the Counter.
func (c *Counter) GetLimit() int64 {
	return c.limit.Load()
}

// GetValue returns the current value of the Counter.
func (c *Counter) GetValue() int64 {
	return c.value.Load()
}

// IncrBy increments the Counter value by alpha up to limit if set.
// Returns true if the increment succeeded.
// Returns false if the increment would exceed limit.
func (c *Counter) IncrBy(alpha int64) bool {
	for {
		val := c.value.Load()
		lim := c.limit.Load()
		if lim > 0 && val+alpha > lim {
			return false
		}
		if c.value.CompareAndSwap(val, val+alpha) {
			return true
		}
	}
}

// DecrBy decrements the Counter value by alpha down to 0.
// Returns true if the decrement succeeded.
// Returns false if the decrement would result in a negative value.
func (c *Counter) DecrBy(alpha int64) bool {
	for {
		val := c.value.Load()
		if val-alpha < 0 {
			return false
		}
		if c.value.CompareAndSwap(val, val-alpha) {
			return true
		}
	}
}

// Reset resets the Counter value to 0.
func (c *Counter) Reset() {
	c.value.Store(0)
}
