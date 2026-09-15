package vals

import (
	"sync"
	"sync/atomic"
	"time"
)

// Set is a thread-safe set for string members with optional expiration.
// If expiresAt is 0, the Set has no expiration.
type Set struct {
	mu        sync.RWMutex
	set       map[string]struct{}
	expiresAt atomic.Int64
}

// NewSet creates a new Set.
func NewSet() *Set {
	return &Set{
		mu:  sync.RWMutex{},
		set: make(map[string]struct{}, 8),
	}
}

// IsExpired returns true if the Set has an expiration time and is expired.
// Returns false if the Set has no expiration time or is not yet expired.
func (s *Set) IsExpired() bool {
	exp := s.expiresAt.Load()
	if exp == 0 {
		return false
	}
	return time.Now().Unix() > exp
}

// Expire sets the expiration time for the Set.
// Returns false if the Set is already expired or if ttl is negative.
func (s *Set) Expire(ttl int64) bool {
	if s.IsExpired() {
		return false
	}
	if ttl < 0 {
		return false
	}
	s.expiresAt.Store(time.Now().Unix() + ttl)
	return true
}

// TTL returns the remaining time-to-live in seconds.
// Returns 0 if the Set has no expiration time.
func (s *Set) TTL() int64 {
	exp := s.expiresAt.Load()
	if exp == 0 {
		return 0
	}
	return exp - time.Now().Unix()
}

// Add adds a member to the Set.
func (s *Set) Add(m string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set[m] = struct{}{}
}

// Remove removes a member from the Set.
func (s *Set) Remove(m string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.set, m)
}

// IsMember returns true if the member is in the Set.
func (s *Set) IsMember(m string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.set[m]
	return ok
}

// Card returns the number of members stored in the Set.
func (s *Set) Card() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.set))
}

// Members returns a slice of all members in the Set.
func (s *Set) Members() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members := make([]string, 0, len(s.set))
	for m := range s.set {
		members = append(members, m)
	}
	return members
}
