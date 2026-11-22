package ns

import (
	"context"

	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/shard/shcounter"
	"github.com/memap-project/memap-core/shard/shhash"
	"github.com/memap-project/memap-core/shard/shmap"
	"github.com/memap-project/memap-core/shard/shrbuffer"
	"github.com/memap-project/memap-core/shard/shset"
)

// Namespace represents an isolated container storing maps, hashes, counters, ring buffers, and sets.
type Namespace struct {
	shmap     *shmap.ShardedMap
	shhash    *shhash.ShardedHash
	shcounter *shcounter.ShardedCounter
	shrbuffer *shrbuffer.ShardedRingBuffer
	shset     *shset.ShardedSet
}

// NewNamespace creates a new Namespace with initialized storage components.
// If cfg is nil, default namespace configuration is used.
func NewNamespace(cfg *config.NamespaceConfig) *Namespace {
	if cfg == nil {
		def := config.DefaultConfig().Namespace
		cfg = &def
	}
	maxKeys := cfg.MaxKeysPerShard
	return &Namespace{
		shmap:     shmap.NewShardedMap(cfg.ShardCounts.Shmap, maxKeys),
		shhash:    shhash.NewShardedHash(cfg.ShardCounts.Shhash, maxKeys),
		shcounter: shcounter.NewShardedCounter(cfg.ShardCounts.Shcounter, maxKeys),
		shrbuffer: shrbuffer.NewShardedRingBufferWithLimits(cfg.ShardCounts.Shrbuffer, cfg.MaxRingBufferCapacity, maxKeys),
		shset:     shset.NewShardedSet(cfg.ShardCounts.Shset, maxKeys),
	}
}

// CleanExpired removes expired keys across all storage components in the namespace.
func (n *Namespace) CleanExpired() {
	n.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes expired keys across all storage components while respecting ctx cancellation.
func (n *Namespace) CleanExpiredCtx(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	n.shmap.CleanExpiredCtx(ctx)
	n.shhash.CleanExpiredCtx(ctx)
	n.shcounter.CleanExpiredCtx(ctx)
	n.shrbuffer.CleanExpiredCtx(ctx)
	n.shset.CleanExpiredCtx(ctx)
}

// Flush removes all keys across all storage components in the namespace.
func (n *Namespace) Flush() {
	n.shmap.Flush()
	n.shhash.Flush()
	n.shcounter.Flush()
	n.shrbuffer.Flush()
	n.shset.Flush()
}
