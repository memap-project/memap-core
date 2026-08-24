package ns

import (
	"sync"

	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/tsmap"
)

// NamespaceManager manages multiple isolated namespaces and a default namespace.
type NamespaceManager struct {
	mu         sync.RWMutex
	cfg        *config.NamespaceConfig
	namespaces tsmap.TypedSyncMap[string, *Namespace]
	defaultNs  *Namespace
}

// NewNamespaceManager creates a new NamespaceManager with an initialized default namespace.
func NewNamespaceManager(cfg *config.NamespaceConfig) *NamespaceManager {
	defaultNs := NewNamespace(cfg)
	return &NamespaceManager{
		mu:         sync.RWMutex{},
		cfg:        cfg,
		namespaces: tsmap.TypedSyncMap[string, *Namespace]{},
		defaultNs:  defaultNs,
	}
}
