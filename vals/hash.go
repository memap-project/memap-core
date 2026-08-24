package vals

import (
	"maps"
	"sync"
	"time"
)

// Hash is a thread-safe map for string key-value pairs with optional expiration.
// If expiresAt is 0, the Hash has no expiration.
type Hash struct {
	mu        sync.RWMutex
	hash      map[string]string
	expiresAt int64
}

// NewHash creates a new Hash.
func NewHash() *Hash {
	return &Hash{
		hash:      make(map[string]string, 8),
		expiresAt: 0,
	}
}

// IsExpired returns true if the Hash has an expiration time and is expired.
// Returns false if the Hash has no expiration time or is not yet expired.
func (h *Hash) IsExpired() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.expiresAt == 0 {
		return false
	}
	return time.Now().Unix() > h.expiresAt
}

// Expire sets the expiration time for the Hash.
// Returns false if the Hash is already expired or if ttl is negative.
func (h *Hash) Expire(ttl int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.IsExpired() {
		return false
	}
	if ttl < 0 {
		return false
	}
	h.expiresAt = time.Now().Unix() + ttl
	return true
}

// TTL returns the remaining time-to-live in seconds.
// Returns 0 if the Hash has no expiration time.
func (h *Hash) TTL() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.expiresAt == 0 {
		return 0
	}
	return h.expiresAt - time.Now().Unix()
}

// GetCopy returns a copy of the Hash.
func (h *Hash) GetCopy() map[string]string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	cp := make(map[string]string, len(h.hash))
	maps.Copy(cp, h.hash)
	return cp
}

// Get retrieves the value for the field in the Hash.
// Returns empty string and false if the field does not exist.
func (h *Hash) Get(key string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	value, ok := h.hash[key]
	if !ok {
		return "", false
	}
	return value, true
}

// Set sets or updates the value for the field in the Hash.
func (h *Hash) Set(key string, value string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hash[key] = value
}

// Delete removes the specified field from the Hash.
func (h *Hash) Delete(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.hash, key)
}

// Len returns the number of fields stored in the Hash.
func (h *Hash) Len() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return int64(len(h.hash))
}

// Keys returns a slice of all field names in the Hash.
func (h *Hash) Keys() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]string, 0, len(h.hash))
	for k := range h.hash {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of all field values in the Hash.
func (h *Hash) Values() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	values := make([]string, 0, len(h.hash))
	for _, v := range h.hash {
		values = append(values, v)
	}
	return values
}
