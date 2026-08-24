package vals

import "time"

// Item represents a string value with optional expiration time.
// If expiresAt is 0, the Item has no expiration.
type Item struct {
	value     string
	expiresAt int64
}

// NewItem creates a new Item.
func NewItem(value string) *Item {
	return &Item{
		value: value,
	}
}

// GetValue returns the string value of the Item.
func (i *Item) GetValue() string {
	return i.value
}

// IsExpired returns true if the Item has an expiration time and is expired.
// Returns false if the Item has no expiration time or is not yet expired.
func (i *Item) IsExpired() bool {
	if i.expiresAt == 0 {
		return false
	}
	return time.Now().Unix() > i.expiresAt
}

// Expire sets the expiration time for the Item.
// Returns false if the Item is already expired or if ttl is negative.
func (i *Item) Expire(ttl int64) bool {
	if i.IsExpired() {
		return false
	}
	if ttl < 0 {
		return false
	}
	i.expiresAt = time.Now().Unix() + ttl
	return true
}

// TTL returns the remaining time-to-live in seconds.
// Returns 0 if the Item has no expiration time.
func (i *Item) TTL() int64 {
	if i.expiresAt == 0 {
		return 0
	}
	return i.expiresAt - time.Now().Unix()
}
