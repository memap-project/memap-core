package tsmap

import "sync"

// TypedSyncMap is a generic thread-safe map wrapper around sync.Map.
type TypedSyncMap[K comparable, V any] struct {
	m sync.Map
}

// Load retrieves a value for the given key from the map.
// Returns zero value and false if the key does not exist.
func (tm *TypedSyncMap[K, V]) Load(key K) (V, bool) {
	actual, loaded := tm.m.Load(key)
	if !loaded {
		var zero V
		return zero, false
	}
	return actual.(V), true
}

// Store sets or updates the value for the given key.
func (tm *TypedSyncMap[K, V]) Store(key K, value V) {
	tm.m.Store(key, value)
}

// LoadOrStore returns the existing value for the key if present, otherwise stores and returns the given value.
// Returns true if the value was loaded, false if stored.
func (tm *TypedSyncMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actual, loaded := tm.m.LoadOrStore(key, value)
	if !loaded {
		var zero V
		return zero, false
	}
	return actual.(V), true
}

// Delete removes the key and its value from the map.
func (tm *TypedSyncMap[K, V]) Delete(key K) {
	tm.m.Delete(key)
}

// Range calls the provided function for each key-value pair in the map.
// If fn returns false, Range stops iteration.
func (tm *TypedSyncMap[K, V]) Range(fn func(key K, value V) bool) {
	tm.m.Range(func(key, value any) bool {
		return fn(key.(K), value.(V))
	})
}
