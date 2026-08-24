package shard

// Statuses maps status codes to human-readable strings.
// Use in the ns package (commands) if the method has many reasons for failing.
type Status uint8

const (
	StatusSuccess Status = iota
	StatusNotFound
	StatusExpired
	StatusBufferEmpty
	StatusIndexOutOfBounds
	StatusLimitExceeded
	StatusFieldNotFound
)
