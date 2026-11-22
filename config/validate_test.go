package config

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		cfg := DefaultConfig()
		require.NoError(t, cfg.Validate())
	})

	t.Run("nil config returns ErrNilConfig", func(t *testing.T) {
		var cfg *Config
		require.ErrorIs(t, cfg.Validate(), ErrNilConfig)
		require.ErrorIs(t, cfg.ValidateCleanerInterval(), ErrNilConfig)
		require.ErrorIs(t, cfg.ValidateShardCount(), ErrNilConfig)
		require.ErrorIs(t, cfg.ValidateMaxRingBufferCapacity(), ErrNilConfig)
	})

	t.Run("cleaner interval rejects zero, negative, and duration overflow", func(t *testing.T) {
		invalidIntervals := []int{
			math.MinInt,
			-10,
			-1,
			0,
			int(math.MaxInt64/int64(time.Second)) + 1,
			math.MaxInt,
		}
		for _, iv := range invalidIntervals {
			cfg := DefaultConfig()
			cfg.CleanerInterval = iv
			assert.ErrorIs(t, cfg.Validate(), ErrInvalidCleanerInterval, "interval=%d", iv)
		}

		valid := DefaultConfig()
		valid.CleanerInterval = 1
		assert.NoError(t, valid.Validate())
	})

	t.Run("shard count validation checks all five components including Shset", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*Config)
		}{
			{"shmap=0", func(c *Config) { c.Namespace.ShardCounts.Shmap = 0 }},
			{"shmap=3", func(c *Config) { c.Namespace.ShardCounts.Shmap = 3 }},
			{"shhash=0", func(c *Config) { c.Namespace.ShardCounts.Shhash = 0 }},
			{"shhash=6", func(c *Config) { c.Namespace.ShardCounts.Shhash = 6 }},
			{"shcounter=0", func(c *Config) { c.Namespace.ShardCounts.Shcounter = 0 }},
			{"shcounter=5", func(c *Config) { c.Namespace.ShardCounts.Shcounter = 5 }},
			{"shrbuffer=0", func(c *Config) { c.Namespace.ShardCounts.Shrbuffer = 0 }},
			{"shrbuffer=12", func(c *Config) { c.Namespace.ShardCounts.Shrbuffer = 12 }},
			{"shset=0", func(c *Config) { c.Namespace.ShardCounts.Shset = 0 }},
			{"shset=7", func(c *Config) { c.Namespace.ShardCounts.Shset = 7 }},
			{"shset=255", func(c *Config) { c.Namespace.ShardCounts.Shset = 255 }},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := DefaultConfig()
				tc.mutate(&cfg)
				assert.ErrorIs(t, cfg.Validate(), ErrInvalidShardCount)
			})
		}
	})

	t.Run("max ring buffer capacity validation", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Namespace.MaxRingBufferCapacity = 0 // 0 means default
		assert.NoError(t, cfg.Validate())

		cfg.Namespace.MaxRingBufferCapacity = -1
		assert.ErrorIs(t, cfg.Validate(), ErrInvalidMaxRingBufferCapacity)

		cfg.Namespace.MaxRingBufferCapacity = math.MaxInt64
		assert.ErrorIs(t, cfg.Validate(), ErrInvalidMaxRingBufferCapacity)

		cfg.Namespace.MaxRingBufferCapacity = 1024
		assert.NoError(t, cfg.Validate())
	})

	t.Run("resource limits reject negative values", func(t *testing.T) {
		cases := []func(*Config){
			func(c *Config) { c.Namespace.MaxKeysPerShard = -1 },
			func(c *Config) { c.Namespace.MaxKeyBytes = -1 },
			func(c *Config) { c.Namespace.MaxValueBytes = -1 },
			func(c *Config) { c.Namespace.MaxNamespaces = -1 },
		}
		for _, mutate := range cases {
			cfg := DefaultConfig()
			mutate(&cfg)
			assert.ErrorIs(t, cfg.Validate(), ErrInvalidResourceLimit)
		}
	})
}

func TestConfig_Normalize(t *testing.T) {
	t.Run("zero config normalizes to valid default config", func(t *testing.T) {
		norm := (Config{}).Normalize()
		require.NoError(t, norm.Validate())
		assert.Equal(t, DefaultConfig(), norm)
	})

	t.Run("invalid values normalize to safe powers of two and defaults", func(t *testing.T) {
		cfg := Config{
			CleanerInterval: -5,
			Namespace: NamespaceConfig{
				ShardCounts: ShardCounts{
					Shmap:     3,
					Shhash:    5,
					Shcounter: 200,
					Shrbuffer: 0,
					Shset:     1,
				},
				MaxRingBufferCapacity: -10,
			},
		}
		norm := cfg.Normalize()
		require.NoError(t, norm.Validate())
		assert.Equal(t, DefaultConfig().CleanerInterval, norm.CleanerInterval)
		assert.Equal(t, uint8(4), norm.Namespace.ShardCounts.Shmap)
		assert.Equal(t, uint8(8), norm.Namespace.ShardCounts.Shhash)
		assert.Equal(t, uint8(128), norm.Namespace.ShardCounts.Shcounter)
		assert.Equal(t, DefaultConfig().Namespace.ShardCounts.Shrbuffer, norm.Namespace.ShardCounts.Shrbuffer)
		assert.Equal(t, uint8(1), norm.Namespace.ShardCounts.Shset)
		assert.Equal(t, DefaultMaxRingBufferCapacity, norm.Namespace.MaxRingBufferCapacity)
	})
}
