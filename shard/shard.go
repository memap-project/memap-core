package shard

import (
	"sync"
)

// Shard is a thread-safe unit of storage containing a map of items protected by an RWMutex.
type Shard[V any] struct {
	mu    sync.RWMutex
	items map[string]V
}

// NewShard creates a new Shard.
func NewShard[V any]() *Shard[V] {
	return &Shard[V]{
		mu:    sync.RWMutex{},
		items: make(map[string]V, 8),
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
	val = initFn()
	s.items[key] = val
	return val, false
}

// Set sets or updates the item for the given key in the Shard.
func (s *Shard[V]) Set(key string, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
}

// Delete removes the item for the given key from the Shard.
func (s *Shard[V]) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// Update updates an item in the Shard under write lock if it exists.
// Returns true if the item was found and the update function returned true.
// Returns false if the item does not exist.
func (s *Shard[V]) Update(key string, fn func(val V) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.items[key]
	if !ok {
		return false
	}
	if fn(val) {
		s.items[key] = val
		return true
	}
	return false
}

// Clean removes items from the Shard for which the predicate returns true.
func (s *Shard[V]) Clean(predicate func(key string, value V) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.items {
		if predicate(k, v) {
			delete(s.items, k)
		}
	}
}

// Flush removes all items from the Shard.
func (s *Shard[V]) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]V, 8)
}
