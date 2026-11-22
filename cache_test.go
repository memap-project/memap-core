package memap_test

import (
	"fmt"
	"math"
	"sync"
	"testing"

	memap "github.com/memap-project/memap-core"
	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_NormalizesInvalidAndPartialConfig(t *testing.T) {
	t.Run("zero config works without panic", func(t *testing.T) {
		c := memap.New(config.Config{})
		defer c.Close()

		require.NoError(t, c.Set("k", "v", 0))
		val, err := c.Get("k")
		require.NoError(t, err)
		assert.Equal(t, "v", val)

		require.NoError(t, c.SAdd("s", "m", 0))
		ok, err := c.SIsMember("s", "m")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("negative cleaner interval and non-power-of-two shards normalize safely", func(t *testing.T) {
		c := memap.New(config.Config{
			CleanerInterval: -1,
			Namespace: config.NamespaceConfig{
				ShardCounts: config.ShardCounts{
					Shmap:     3,
					Shhash:    0,
					Shcounter: 5,
					Shrbuffer: 0,
					Shset:     0,
				},
			},
		})
		defer c.Close()

		normCfg := c.Config()
		require.NoError(t, normCfg.Validate())
		require.NoError(t, c.Set("k1", "v1", 60))
		require.NoError(t, c.HFSet("h1", "f1", "v1"))
		_, err := c.CIncrBy("c1", 1)
		require.NoError(t, err)
		require.NoError(t, c.BInit("b1", 4, 0))
		require.NoError(t, c.SAdd("s1", "m1", 0))
	})
}

func TestNewWithConfig_StrictValidation(t *testing.T) {
	t.Run("rejects invalid shard count including Shset", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Namespace.ShardCounts.Shset = 0
		c, err := memap.NewWithConfig(cfg)
		require.ErrorIs(t, err, config.ErrInvalidShardCount)
		assert.Nil(t, c)
	})

	t.Run("rejects negative and overflowing cleaner interval", func(t *testing.T) {
		for _, iv := range []int{-1, 0, math.MaxInt} {
			cfg := config.DefaultConfig()
			cfg.CleanerInterval = iv
			c, err := memap.NewWithConfig(cfg)
			require.ErrorIs(t, err, config.ErrInvalidCleanerInterval)
			assert.Nil(t, c)
		}
	})

	t.Run("accepts valid config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Namespace.MaxRingBufferCapacity = 32
		c, err := memap.NewWithConfig(cfg)
		require.NoError(t, err)
		require.NotNil(t, c)
		defer c.Close()

		require.ErrorIs(t, c.BInit("buf", 33, 0), ns.ErrInvalidCapacity)
		require.NoError(t, c.BInit("buf", 32, 0))
	})
}

func TestCache_BInit_InvalidCapacities(t *testing.T) {
	c := memap.New()
	defer c.Close()

	invalidCaps := []int64{
		math.MinInt64,
		-100,
		-1,
		0,
		config.DefaultMaxRingBufferCapacity + 1,
		math.MaxInt64,
	}

	for _, capVal := range invalidCaps {
		key := fmt.Sprintf("buf_%d", capVal)
		err := c.BInit(key, capVal, 0)
		assert.ErrorIs(t, err, ns.ErrInvalidCapacity, "cap=%d", capVal)
		assert.ErrorIs(t, c.BPush(key, "item"), ns.ErrKeyNotFound)
	}
}

func TestCache_ConcurrentEraseAndCleaner_NoRace(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CleanerInterval = 1
	c, err := memap.NewWithConfig(cfg)
	require.NoError(t, err)
	defer c.Close()

	const workers = 10
	const iterations = 400

	var wg sync.WaitGroup
	wg.Add(workers + 2)

	for g := range workers {
		go func(id int) {
			defer wg.Done()
			nsName := fmt.Sprintf("ns-%d", id%4)
			for i := range iterations {
				_ = c.CreateNamespace(nsName)
				scoped := c.WithNamespace(nsName)
				_ = scoped.Set("k", "v", 1)
				_, _ = scoped.Get("k")
				_ = scoped.SAdd("s", "m", 1)
				if i%19 == 0 {
					_ = c.DropNamespace(nsName)
				}
				if i%31 == 0 {
					c.Flush()
				}
			}
		}(g)
	}

	// Dedicated erasers racing against namespace ops and background Cleaner
	for range 2 {
		go func() {
			defer wg.Done()
			for range iterations / 4 {
				c.Erase()
			}
		}()
	}

	wg.Wait()
}
