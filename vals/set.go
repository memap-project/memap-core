package vals

import (
	"maps"
	"sync"
)

// Set is a thread-safe set for string members with optional monotonic expiration.
type Set struct {
	expiry
	mu      sync.RWMutex
	set     map[string]struct{}
	peakLen int
}

// NewSet creates a new Set.
func NewSet() *Set {
	return &Set{
		mu:  sync.RWMutex{},
		set: make(map[string]struct{}, 8),
	}
}

// Add adds a member to the Set.
func (s *Set) Add(m string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set[m] = struct{}{}
	if len(s.set) > s.peakLen {
		s.peakLen = len(s.set)
	}
}

// Remove removes a member from the Set and shrinks the underlying map
// if occupancy drops significantly below its peak size.
func (s *Set) Remove(m string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.set, m)
	if s.peakLen >= mapShrinkMinPeak && len(s.set)*4 <= s.peakLen {
		newCap := max(len(s.set), 8)
		shrunk := make(map[string]struct{}, newCap)
		maps.Copy(shrunk, s.set)
		s.set = shrunk
		s.peakLen = len(s.set)
	}
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
