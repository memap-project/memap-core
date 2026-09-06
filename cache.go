package memap

import (
	"context"

	"github.com/memap-project/memap-core/clean"
	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/ns"
)

// Cache is the main database instance implementing Core.
type Cache struct {
	manager *ns.NamespaceManager
	cleaner *clean.Cleaner
	cfg     config.Config
}

// Namespace represents scoped store operations within a specific namespace.
type Namespace struct {
	name    string
	manager *ns.NamespaceManager
}

// Ensure Cache implements Core.
var _ Core = (*Cache)(nil)

// Ensure Namespace implements store interfaces.
var (
	_ KVStore         = (*Namespace)(nil)
	_ HashStore       = (*Namespace)(nil)
	_ CounterStore    = (*Namespace)(nil)
	_ RingBufferStore = (*Namespace)(nil)
)

// New creates and starts a new Cache instance with optional configuration.
func New(cfg ...config.Config) *Cache {
	c := config.DefaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	manager := ns.NewNamespaceManager(&c.Namespace)
	cleaner := clean.NewCleaner(context.Background(), c.CleanerInterval, manager.CleanExpired)
	cleaner.Start()
	return &Cache{
		manager: manager,
		cleaner: cleaner,
		cfg:     c,
	}
}

// Close stops the background expiration cleaner.
func (c *Cache) Close() {
	c.cleaner.Stop()
}

// Config returns the configuration used by the Cache instance.
func (c *Cache) Config() config.Config {
	return c.cfg
}

// WithNamespace returns a Namespace scoped to the given namespace name.
func (c *Cache) WithNamespace(name string) *Namespace {
	return &Namespace{
		name:    name,
		manager: c.manager,
	}
}
