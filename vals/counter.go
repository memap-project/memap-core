package vals

import (
	"sync"
	"time"
)

// Counter is a thread-safe integer counter with optional limit and expiration.
// If expiresAt is 0, the Counter has no expiration.
// If limit is 0, the Counter has no upper limit.
type Counter struct {
	mu        sync.RWMutex
	value     int64
	limit     int64
	expiresAt int64
}

// NewCounter creates a new Counter with zero values.
func NewCounter() *Counter {
	return &Counter{
		mu: sync.RWMutex{},
	}
}

// IsExpired returns true if the Counter has an expiration time and is expired.
// Returns false if the Counter has no expiration time or is not yet expired.
func (c *Counter) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.expiresAt == 0 {
		return false
	}
	return time.Now().Unix() > c.expiresAt
}

// Expire sets the expiration time for the Counter.
// Returns false if the Counter is already expired or if ttl is negative.
func (c *Counter) Expire(ttl int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.IsExpired() {
		return false
	}
	if ttl < 0 {
		return false
	}
	c.expiresAt = time.Now().Unix() + ttl
	return true
}

// TTL returns the remaining time-to-live in seconds.
// Returns 0 if the Counter has no expiration time.
func (c *Counter) TTL() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.expiresAt == 0 {
		return 0
	}
	return c.expiresAt - time.Now().Unix()
}

// SetLimit sets or updates the upper limit of the Counter.
func (c *Counter) SetLimit(limit int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.limit = limit
}

// GetLimit returns the upper limit of the Counter.
func (c *Counter) GetLimit() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.limit
}

// GetValue returns the current value of the Counter.
func (c *Counter) GetValue() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// IncrBy increments the Counter value by alpha up to limit if set.
// Returns true if the increment succeeded.
// Returns false if the increment would exceed limit.
func (c *Counter) IncrBy(alpha int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.limit > 0 && c.value+alpha > c.limit {
		return false
	}
	c.value += alpha
	return true
}

// DecrBy decrements the Counter value by alpha down to 0.
// Returns true if the decrement succeeded.
// Returns false if the decrement would result in a negative value.
func (c *Counter) DecrBy(alpha int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.value-alpha >= 0 {
		c.value -= alpha
		return true
	}
	return false
}

// Reset resets the Counter value to 0.
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}
