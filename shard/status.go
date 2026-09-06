package shard

// Status represents an operation result status code.
// Used in storage and namespace operations when a method has multiple reasons for failing.
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
