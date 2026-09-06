package config

import (
	"errors"
)

// Configuration validation errors.
var (
	ErrInvalidShardCount      = errors.New("shard count must be a power of two")
	ErrInvalidCleanerInterval = errors.New("cleanerInterval must be greater than 0")
)

// isPowerOfTwo reports whether n is a power of two.
func isPowerOfTwo(n uint8) bool {
	return n > 0 && (n&(n-1)) == 0
}

// ValidateCleanerInterval returns an error if the cleaner interval is not greater than 0.
func (c *Config) ValidateCleanerInterval() error {
	if c.CleanerInterval == 0 {
		return ErrInvalidCleanerInterval
	}
	return nil
}

// ValidateShardCount returns an error if any shard count is not a power of two.
func (c *Config) ValidateShardCount() error {
	if !isPowerOfTwo(c.Namespace.ShardCounts.Shmap) {
		return ErrInvalidShardCount
	}
	if !isPowerOfTwo(c.Namespace.ShardCounts.Shhash) {
		return ErrInvalidShardCount
	}
	if !isPowerOfTwo(c.Namespace.ShardCounts.Shcounter) {
		return ErrInvalidShardCount
	}
	if !isPowerOfTwo(c.Namespace.ShardCounts.Shrbuffer) {
		return ErrInvalidShardCount
	}
	return nil
}

// Validate validates the core configuration.
func (c *Config) Validate() error {
	if err := c.ValidateCleanerInterval(); err != nil {
		return err
	}
	if err := c.ValidateShardCount(); err != nil {
		return err
	}
	return nil
}
