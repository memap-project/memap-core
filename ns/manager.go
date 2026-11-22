package ns

import (
	"sync/atomic"

	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/tsmap"
)

type nsRegistry struct {
	m     tsmap.TypedSyncMap[string, *Namespace]
	count atomic.Int64
}

// NamespaceManager manages multiple isolated namespaces and a default namespace.
type NamespaceManager struct {
	cfg       config.NamespaceConfig
	registry  atomic.Pointer[nsRegistry]
	defaultNs *Namespace
}

// NewNamespaceManager creates a new NamespaceManager with an initialized default namespace.
// If cfg is nil, default namespace configuration is used.
func NewNamespaceManager(cfg *config.NamespaceConfig) *NamespaceManager {
	def := config.DefaultConfig().Namespace
	var nsCfg config.NamespaceConfig
	if cfg == nil {
		nsCfg = def
	} else {
		nsCfg = *cfg
		if nsCfg.MaxRingBufferCapacity <= 0 {
			nsCfg.MaxRingBufferCapacity = def.MaxRingBufferCapacity
		}
		if nsCfg.MaxKeyBytes <= 0 {
			nsCfg.MaxKeyBytes = def.MaxKeyBytes
		}
		if nsCfg.MaxValueBytes <= 0 {
			nsCfg.MaxValueBytes = def.MaxValueBytes
		}
		if nsCfg.MaxNamespaces <= 0 {
			nsCfg.MaxNamespaces = def.MaxNamespaces
		}
	}
	nm := &NamespaceManager{
		cfg:       nsCfg,
		defaultNs: NewNamespace(&nsCfg),
	}
	nm.registry.Store(&nsRegistry{})
	return nm
}

func (nm *NamespaceManager) validateKey(key string) error {
	if nm.cfg.MaxKeyBytes > 0 && len(key) > nm.cfg.MaxKeyBytes {
		return ErrKeyTooLarge
	}
	return nil
}

func (nm *NamespaceManager) validateValue(val string) error {
	if nm.cfg.MaxValueBytes > 0 && len(val) > nm.cfg.MaxValueBytes {
		return ErrValueTooLarge
	}
	return nil
}
