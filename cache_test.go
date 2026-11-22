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

func TestCache_HighSecurityFixes(t *testing.T) {
	t.Run("counter rejects negative alpha, negative limit, and int64 overflow", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		require.ErrorIs(t, c.CSLimit("cnt", -1), ns.ErrInvalidArg)
		_, err := c.CGet("cnt")
		require.ErrorIs(t, err, ns.ErrKeyNotFound)

		require.NoError(t, c.CSLimit("cnt", 100))
		v, err := c.CIncrBy("cnt", 50)
		require.NoError(t, err)
		assert.Equal(t, int64(50), v)

		_, err = c.CIncrBy("cnt", math.MaxInt64)
		require.ErrorIs(t, err, ns.ErrLimitExceeded)
		_, err = c.CIncrBy("cnt", -5)
		require.ErrorIs(t, err, ns.ErrInvalidArg)
		_, err = c.CDecrBy("cnt", -1000)
		require.ErrorIs(t, err, ns.ErrInvalidArg)

		v, err = c.CGet("cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(50), v)
	})

	t.Run("Expire returns ErrInvalidTTL for non-positive or overflowing TTL on existing keys", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		require.NoError(t, c.Set("k", "v", 0))
		require.ErrorIs(t, c.Expire("k", -1), ns.ErrInvalidTTL)
		require.ErrorIs(t, c.Expire("k", 0), ns.ErrInvalidTTL)
		require.ErrorIs(t, c.Expire("k", math.MaxInt64), ns.ErrInvalidTTL)

		require.NoError(t, c.HFSet("h", "f", "v"))
		require.ErrorIs(t, c.HExpire("h", -1), ns.ErrInvalidTTL)

		_, err := c.CIncrBy("cnt", 1)
		require.NoError(t, err)
		require.ErrorIs(t, c.CExpire("cnt", 0), ns.ErrInvalidTTL)

		require.NoError(t, c.BInit("rb", 4, 0))
		require.ErrorIs(t, c.BExpire("rb", -5), ns.ErrInvalidTTL)

		require.NoError(t, c.SAdd("s", "m", 0))
		require.ErrorIs(t, c.SExpire("s", math.MaxInt64), ns.ErrInvalidTTL)
	})

	t.Run("enforces MaxKeyBytes, MaxValueBytes, MaxNamespaces, and MaxKeysPerShard", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Namespace.MaxKeyBytes = 8
		cfg.Namespace.MaxValueBytes = 16
		cfg.Namespace.MaxNamespaces = 2
		cfg.Namespace.MaxKeysPerShard = 2
		cfg.Namespace.ShardCounts.Shmap = 1

		c, err := memap.NewWithConfig(cfg)
		require.NoError(t, err)
		defer c.Close()

		require.ErrorIs(t, c.Set("123456789", "ok", 0), ns.ErrKeyTooLarge)
		require.ErrorIs(t, c.Set("ok", "12345678901234567", 0), ns.ErrValueTooLarge)

		require.NoError(t, c.CreateNamespace("ns1"))
		require.NoError(t, c.CreateNamespace("ns2"))
		require.ErrorIs(t, c.CreateNamespace("ns3"), ns.ErrNamespaceLimitExceeded)
		require.NoError(t, c.DropNamespace("ns1"))
		require.NoError(t, c.CreateNamespace("ns3"))

		// MaxKeysPerShard=2 on 1 shard evicts when adding 3rd key
		require.NoError(t, c.Set("k1", "v1", 0))
		require.NoError(t, c.Set("k2", "v2", 0))
		require.NoError(t, c.Set("k3", "v3", 0))
		present := 0
		for _, k := range []string{"k1", "k2", "k3"} {
			if _, err := c.Get(k); err == nil {
				present++
			}
		}
		assert.Equal(t, 2, present)
	})
}

func TestCache_MediumFixes(t *testing.T) {
	t.Run("Core interface includes SetStore, SDel, and CReset", func(t *testing.T) {
		var core memap.Core = memap.New()
		defer core.Close()

		require.NoError(t, core.SAdd("tags", "go", 0))
		require.NoError(t, core.SAdd("tags", "cache", 0))
		require.NoError(t, core.SDel("tags"))
		_, err := core.SCard("tags")
		require.ErrorIs(t, err, memap.ErrKeyNotFound)

		_, err = core.CIncrBy("hits", 5)
		require.NoError(t, err)
		require.NoError(t, core.CReset("hits"))
		v, err := core.CGet("hits")
		require.NoError(t, err)
		assert.Equal(t, int64(0), v)
		require.ErrorIs(t, core.CReset("missing_counter"), memap.ErrKeyNotFound)
	})

	t.Run("HSet overwrites existing hash and clears old fields", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		require.NoError(t, c.HFSet("session:1", "secret", "top-secret"))
		require.NoError(t, c.HSet("session:1", 60))

		_, err := c.HFGet("session:1", "secret")
		require.ErrorIs(t, err, memap.ErrFieldNotFound)
		l, err := c.HLen("session:1")
		require.NoError(t, err)
		assert.Equal(t, int64(0), l)
	})

	t.Run("CreateNamespace rejects empty name", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		require.ErrorIs(t, c.CreateNamespace(""), memap.ErrInvalidNamespace)
	})

	t.Run("BReset returns ErrKeyNotFound for missing buffer", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		require.ErrorIs(t, c.BReset("missing_buf"), memap.ErrKeyNotFound)
	})

	t.Run("Set, HSet, SAdd, and BInit reject negative or overflowing TTL", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		require.ErrorIs(t, c.Set("k", "v", -1), memap.ErrInvalidTTL)
		require.ErrorIs(t, c.Set("k", "v", math.MaxInt64), memap.ErrInvalidTTL)
		require.ErrorIs(t, c.HSet("h", -1), memap.ErrInvalidTTL)
		require.ErrorIs(t, c.SAdd("s", "m", -1), memap.ErrInvalidTTL)
		require.ErrorIs(t, c.BInit("b", 4, -1), memap.ErrInvalidTTL)
	})

	t.Run("TTL methods return 0 consistently on ErrNamespaceNotFound and miss paths return nil", func(t *testing.T) {
		c := memap.New()
		defer c.Close()

		missing := c.WithNamespace("no_such_ns")
		for _, fn := range []func(string) (int64, error){
			missing.TTL, missing.HTTL, missing.CTTL, missing.BTTL, missing.STTL,
		} {
			ttl, err := fn("k")
			assert.Equal(t, int64(0), ttl)
			assert.ErrorIs(t, err, memap.ErrNamespaceNotFound)
		}

		h, err := c.HGet("missing")
		assert.Nil(t, h)
		assert.ErrorIs(t, err, memap.ErrKeyNotFound)

		hk, err := missing.HKeys("k")
		assert.Nil(t, hk)
		assert.ErrorIs(t, err, memap.ErrNamespaceNotFound)

		sm, err := c.SMembers("missing")
		assert.Nil(t, sm)
		assert.ErrorIs(t, err, memap.ErrKeyNotFound)
	})

	t.Run("nil Cache Close is safe", func(t *testing.T) {
		var nilCache *memap.Cache
		assert.NotPanics(t, func() {
			nilCache.Close()
		})
	})
}

func BenchmarkKV_SetOverwrite(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.Set("bench-key", "val", 60)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.Set("bench-key", "val2", 60)
	}
}

func BenchmarkKV_GetHit(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.Set("bench-key", "val", 60)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get("bench-key")
	}
}

func BenchmarkHash_GetMiss(b *testing.B) {
	c := memap.New()
	defer c.Close()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.HGet("missing-hash")
	}
}
