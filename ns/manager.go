package ns

import (
	"sync/atomic"

	"github.com/memap-project/memap-core/config"
	"github.com/memap-project/memap-core/tsmap"
)

// NamespaceManager manages multiple isolated namespaces and a default namespace.
type NamespaceManager struct {
	cfg        config.NamespaceConfig
	namespaces atomic.Pointer[tsmap.TypedSyncMap[string, *Namespace]]
	defaultNs  *Namespace
}

// NewNamespaceManager creates a new NamespaceManager with an initialized default namespace.
// If cfg is nil, default namespace configuration is used.
func NewNamespaceManager(cfg *config.NamespaceConfig) *NamespaceManager {
	var nsCfg config.NamespaceConfig
	if cfg == nil {
		nsCfg = config.DefaultConfig().Namespace
	} else {
		nsCfg = *cfg
	}
	nm := &NamespaceManager{
		cfg:       nsCfg,
		defaultNs: NewNamespace(&nsCfg),
	}
	nm.namespaces.Store(&tsmap.TypedSyncMap[string, *Namespace]{})
	return nm
}
