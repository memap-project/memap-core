package ns

import (
	"sync"

	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/shard/shcounter"
	"github.com/memap-project/memap-core/shard/shhash"
	"github.com/memap-project/memap-core/shard/shmap"
	"github.com/memap-project/memap-core/shard/shrbuffer"
)

// Namespace represents an isolated container storing maps, hashes, and counters.
type Namespace struct {
	mu        sync.RWMutex
	cfg       *config.NamespaceConfig
	shmap     *shmap.ShardedMap
	shhash    *shhash.ShardedHash
	shcounter *shcounter.ShardedCounter
	shrbuffer *shrbuffer.ShardedRingBuffer
}

// NewNamespace creates a new Namespace with initialized storage components.
func NewNamespace(cfg *config.NamespaceConfig) *Namespace {
	return &Namespace{
		mu:        sync.RWMutex{},
		cfg:       cfg,
		shmap:     shmap.NewShardedMap(cfg.ShardCounts.Shmap),
		shhash:    shhash.NewShardedHash(cfg.ShardCounts.Shhash),
		shcounter: shcounter.NewShardedCounter(cfg.ShardCounts.Shcounter),
		shrbuffer: shrbuffer.NewShardedRingBuffer(cfg.ShardCounts.Shrbuffer),
	}
}

// CleanExpired removes expired keys across all storage components in the namespace.
func (n *Namespace) CleanExpired() {
	n.shmap.CleanExpired()
	n.shhash.CleanExpired()
	n.shcounter.CleanExpired()
	n.shrbuffer.CleanExpired()
}

// Flush removes all keys across all storage components in the namespace.
func (n *Namespace) Flush() {
	n.shmap.Flush()
	n.shhash.Flush()
	n.shcounter.Flush()
	n.shrbuffer.Flush()
}
