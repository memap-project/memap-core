package config

import (
	"errors"
	"math"
	"time"
)

const maxCleanerIntervalSec int64 = math.MaxInt64 / int64(time.Second)

// Configuration validation errors.
var (
	ErrNilConfig                    = errors.New("config must not be nil")
	ErrInvalidShardCount            = errors.New("shard count must be a power of two")
	ErrInvalidCleanerInterval       = errors.New("cleanerInterval must be greater than 0 and not overflow time.Duration")
	ErrInvalidMaxRingBufferCapacity = errors.New("maxRingBufferCapacity must be non-negative and within allocatable bounds")
	ErrInvalidResourceLimit         = errors.New("resource limits must be non-negative")
)

// isPowerOfTwo reports whether n is a power of two.
func isPowerOfTwo(n uint8) bool {
	return n > 0 && (n&(n-1)) == 0
}

// ValidateCleanerInterval returns an error if the cleaner interval is not strictly positive
// or would overflow time.Duration when multiplied by time.Second.
func (c *Config) ValidateCleanerInterval() error {
	if c == nil {
		return ErrNilConfig
	}
	if c.CleanerInterval <= 0 || int64(c.CleanerInterval) > maxCleanerIntervalSec {
		return ErrInvalidCleanerInterval
	}
	return nil
}

// ValidateShardCount returns an error if any shard count is not a power of two.
func (c *Config) ValidateShardCount() error {
	if c == nil {
		return ErrNilConfig
	}
	counts := [5]uint8{
		c.Namespace.ShardCounts.Shmap,
		c.Namespace.ShardCounts.Shhash,
		c.Namespace.ShardCounts.Shcounter,
		c.Namespace.ShardCounts.Shrbuffer,
		c.Namespace.ShardCounts.Shset,
	}
	for _, n := range counts {
		if !isPowerOfTwo(n) {
			return ErrInvalidShardCount
		}
	}
	return nil
}

// ValidateMaxRingBufferCapacity returns an error if MaxRingBufferCapacity is negative
// or exceeds the maximum safe allocation limit. Zero is allowed and selects DefaultMaxRingBufferCapacity.
func (c *Config) ValidateMaxRingBufferCapacity() error {
	if c == nil {
		return ErrNilConfig
	}
	if c.Namespace.MaxRingBufferCapacity < 0 || c.Namespace.MaxRingBufferCapacity > maxAllowedRingBufferCap {
		return ErrInvalidMaxRingBufferCapacity
	}
	return nil
}

// ValidateResourceLimits returns an error if any resource limit is negative.
func (c *Config) ValidateResourceLimits() error {
	if c == nil {
		return ErrNilConfig
	}
	if c.Namespace.MaxKeysPerShard < 0 ||
		c.Namespace.MaxKeyBytes < 0 ||
		c.Namespace.MaxValueBytes < 0 ||
		c.Namespace.MaxNamespaces < 0 {
		return ErrInvalidResourceLimit
	}
	return nil
}

// Validate validates the core configuration.
func (c *Config) Validate() error {
	if c == nil {
		return ErrNilConfig
	}
	if err := c.ValidateCleanerInterval(); err != nil {
		return err
	}
	if err := c.ValidateShardCount(); err != nil {
		return err
	}
	if err := c.ValidateMaxRingBufferCapacity(); err != nil {
		return err
	}
	if err := c.ValidateResourceLimits(); err != nil {
		return err
	}
	return nil
}
