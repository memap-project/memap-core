package memap_test

import (
	"fmt"
	"testing"

	memap "github.com/memap-project/memap-core"
	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/shard/shmap"
)

func makeKeys(n int, prefix string) []string {
	keys := make([]string, n)
	for i := range n {
		keys[i] = fmt.Sprintf("%s:%d", prefix, i)
	}
	return keys
}

func BenchmarkKV_SetNew(b *testing.B) {
	keys := makeKeys(4096, "kv")
	c := memap.New()
	defer c.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		if i&4095 == 0 {
			b.StopTimer()
			c.Flush()
			b.StartTimer()
		}
		_ = c.Set(keys[i&4095], "payload-value", 60)
	}
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

func BenchmarkKV_GetMiss(b *testing.B) {
	c := memap.New()
	defer c.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get("missing-key")
	}
}

func BenchmarkKV_GetHitParallel(b *testing.B) {
	const keyCount = 4096
	keys := makeKeys(keyCount, "kv")
	c := memap.New()
	defer c.Close()
	for _, k := range keys {
		_ = c.Set(k, "payload-value", 60)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = c.Get(keys[i&(keyCount-1)])
			i++
		}
	})
}

func BenchmarkKV_MixedReadWriteParallel(b *testing.B) {
	const keyCount = 4096
	keys := makeKeys(keyCount, "mix")
	c := memap.New()
	defer c.Close()
	for _, k := range keys {
		_ = c.Set(k, "payload-value", 60)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := keys[i&(keyCount-1)]
			if i%5 == 0 {
				_ = c.Set(k, "updated-value", 60)
			} else {
				_, _ = c.Get(k)
			}
			i++
		}
	})
}

func BenchmarkKV_SetWithEviction(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.Namespace.MaxKeysPerShard = 128
	c, err := memap.NewWithConfig(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()

	keys := makeKeys(4096, "evict")
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		_ = c.Set(keys[i&4095], "payload-value", 60)
	}
}

func BenchmarkHash_HFSet(b *testing.B) {
	c := memap.New()
	defer c.Close()
	fields := makeKeys(16, "field")

	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		_ = c.HFSet("user:1", fields[i&15], "value")
	}
}

func BenchmarkHash_HFGetHit(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.HFSet("user:1", "email", "alice@example.com")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.HFGet("user:1", "email")
	}
}

func BenchmarkHash_HGetCopy(b *testing.B) {
	c := memap.New()
	defer c.Close()
	for i := range 10 {
		_ = c.HFSet("user:1", fmt.Sprintf("f%d", i), "value")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.HGet("user:1")
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

func BenchmarkCounter_IncrBy(b *testing.B) {
	c := memap.New()
	defer c.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.CIncrBy("rl:user:1", 1)
	}
}

func BenchmarkCounter_IncrByParallelHotKey(b *testing.B) {
	c := memap.New()
	defer c.Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.CIncrBy("rl:hot", 1)
		}
	})
}

func BenchmarkCounter_IncrByParallelSharded(b *testing.B) {
	const keyCount = 1024
	keys := makeKeys(keyCount, "rl")
	c := memap.New()
	defer c.Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = c.CIncrBy(keys[i&(keyCount-1)], 1)
			i++
		}
	})
}

func BenchmarkRingBuffer_PushOverwrite(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.BInit("events", 64, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.BPush("events", "event-payload")
	}
}

func BenchmarkRingBuffer_PushPop(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.BInit("queue", 64, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.BPush("queue", "event-payload")
		_, _ = c.BPop("queue")
	}
}

func BenchmarkRingBuffer_Slice64(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.BInit("events", 64, 0)
	for i := range 64 {
		_ = c.BPush("events", fmt.Sprintf("evt-%d", i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.BSlice("events")
	}
}

func BenchmarkSet_SAdd(b *testing.B) {
	c := memap.New()
	defer c.Close()
	members := makeKeys(64, "member")

	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		_ = c.SAdd("tags", members[i&63], 0)
	}
}

func BenchmarkSet_SIsMemberParallel(b *testing.B) {
	c := memap.New()
	defer c.Close()
	members := makeKeys(64, "member")
	for _, m := range members {
		_ = c.SAdd("tags", m, 0)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = c.SIsMember("tags", members[i&63])
			i++
		}
	})
}

func BenchmarkNamespace_ScopedGetHit(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.CreateNamespace("tenant-a")
	ns := c.WithNamespace("tenant-a")
	_ = ns.Set("k", "v", 60)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = ns.Get("k")
	}
}

func BenchmarkNamespace_CreateExisting(b *testing.B) {
	c := memap.New()
	defer c.Close()
	_ = c.CreateNamespace("existing")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.CreateNamespace("existing")
	}
}

func BenchmarkCleanExpired_10kActiveKeys(b *testing.B) {
	shm := shmap.NewShardedMap(8)
	keys := makeKeys(10_000, "clean")
	for _, k := range keys {
		shm.Set(k, "val", 3600)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		shm.CleanExpired()
	}
}
