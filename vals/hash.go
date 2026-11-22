package vals

import (
	"maps"
	"sync"
)

const mapShrinkMinPeak = 256

// Hash is a thread-safe map for string key-value pairs with optional monotonic expiration.
type Hash struct {
	expiry
	mu      sync.RWMutex
	hash    map[string]string
	peakLen int
}

// NewHash creates a new Hash.
func NewHash() *Hash {
	return &Hash{
		mu:   sync.RWMutex{},
		hash: make(map[string]string, 8),
	}
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
	if len(h.hash) > h.peakLen {
		h.peakLen = len(h.hash)
	}
}

// Delete removes the specified field from the Hash and shrinks the underlying map
// if occupancy drops significantly below its peak size.
func (h *Hash) Delete(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.hash, key)
	if h.peakLen >= mapShrinkMinPeak && len(h.hash)*4 <= h.peakLen {
		newCap := max(len(h.hash), 8)
		shrunk := make(map[string]string, newCap)
		maps.Copy(shrunk, h.hash)
		h.hash = shrunk
		h.peakLen = len(h.hash)
	}
}

// Len returns the number of fields stored in the Hash.
func (h *Hash) Len() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return int64(len(h.hash))
}

// Keys returns all field names in the Hash.
func (h *Hash) Keys() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]string, 0, len(h.hash))
	for k := range h.hash {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all field values in the Hash.
func (h *Hash) Values() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	values := make([]string, 0, len(h.hash))
	for _, v := range h.hash {
		values = append(values, v)
	}
	return values
}
