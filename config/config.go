package config

import "math"

// DefaultMaxRingBufferCapacity is the default upper bound for a single ring buffer's capacity.
const DefaultMaxRingBufferCapacity int64 = 1 << 20

// Config represents configuration for the memap core.
type Config struct {
	CleanerInterval int             `yaml:"cleanerInterval" json:"cleanerInterval"`
	Namespace       NamespaceConfig `yaml:"namespace" json:"namespace"`
}

// NamespaceConfig represents configuration for namespaces.
type NamespaceConfig struct {
	ShardCounts           ShardCounts `yaml:"shardCounts" json:"shardCounts"`
	MaxRingBufferCapacity int64       `yaml:"maxRingBufferCapacity" json:"maxRingBufferCapacity"`
}

// ShardCounts specifies the number of shards for each storage component.
type ShardCounts struct {
	Shmap     uint8 `yaml:"shmap" json:"shmap"`
	Shhash    uint8 `yaml:"shhash" json:"shhash"`
	Shcounter uint8 `yaml:"shcounter" json:"shcounter"`
	Shrbuffer uint8 `yaml:"shrbuffer" json:"shrbuffer"`
	Shset     uint8 `yaml:"shset" json:"shset"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		CleanerInterval: 10,
		Namespace: NamespaceConfig{
			ShardCounts: ShardCounts{
				Shmap:     8,
				Shhash:    8,
				Shcounter: 8,
				Shrbuffer: 8,
				Shset:     8,
			},
			MaxRingBufferCapacity: DefaultMaxRingBufferCapacity,
		},
	}
}

// Normalize returns a copy of Config where invalid or unset fields are replaced
// with safe defaults so the configuration can be used without panicking.
func (c Config) Normalize() Config {
	def := DefaultConfig()
	if c.CleanerInterval <= 0 || int64(c.CleanerInterval) > maxCleanerIntervalSec {
		c.CleanerInterval = def.CleanerInterval
	}
	c.Namespace.ShardCounts.Shmap = normalizeShardCount(c.Namespace.ShardCounts.Shmap, def.Namespace.ShardCounts.Shmap)
	c.Namespace.ShardCounts.Shhash = normalizeShardCount(c.Namespace.ShardCounts.Shhash, def.Namespace.ShardCounts.Shhash)
	c.Namespace.ShardCounts.Shcounter = normalizeShardCount(c.Namespace.ShardCounts.Shcounter, def.Namespace.ShardCounts.Shcounter)
	c.Namespace.ShardCounts.Shrbuffer = normalizeShardCount(c.Namespace.ShardCounts.Shrbuffer, def.Namespace.ShardCounts.Shrbuffer)
	c.Namespace.ShardCounts.Shset = normalizeShardCount(c.Namespace.ShardCounts.Shset, def.Namespace.ShardCounts.Shset)
	if c.Namespace.MaxRingBufferCapacity <= 0 || c.Namespace.MaxRingBufferCapacity > maxAllowedRingBufferCap {
		c.Namespace.MaxRingBufferCapacity = def.Namespace.MaxRingBufferCapacity
	}
	return c
}

const maxAllowedRingBufferCap int64 = math.MaxInt / 16

func normalizeShardCount(n, fallback uint8) uint8 {
	if n == 0 {
		return fallback
	}
	if isPowerOfTwo(n) {
		return n
	}
	if n > 128 {
		return 128
	}
	var p uint8 = 1
	for p < n {
		p <<= 1
	}
	return p
}
