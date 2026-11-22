package ns

import "context"

// GetNs retrieves a namespace by name.
// Returns the namespace and true if found, nil and false otherwise.
func (nm *NamespaceManager) GetNs(name string) (*Namespace, bool) {
	return nm.registry.Load().m.Load(name)
}

// Create creates a new namespace with the given name.
// Returns [ErrKeyTooLarge] if name exceeds MaxKeyBytes.
// Returns [ErrNamespaceAlreadyExists] if the namespace already exists.
// Returns [ErrNamespaceLimitExceeded] if MaxNamespaces is reached.
func (nm *NamespaceManager) Create(name string) error {
	if err := nm.validateKey(name); err != nil {
		return err
	}
	reg := nm.registry.Load()
	if _, exists := reg.m.Load(name); exists {
		return ErrNamespaceAlreadyExists
	}
	if maxNS := int64(nm.cfg.MaxNamespaces); maxNS > 0 {
		for {
			cur := reg.count.Load()
			if cur >= maxNS {
				return ErrNamespaceLimitExceeded
			}
			if reg.count.CompareAndSwap(cur, cur+1) {
				break
			}
		}
	} else {
		reg.count.Add(1)
	}

	cfg := nm.cfg
	ns := NewNamespace(&cfg)
	if _, loaded := reg.m.LoadOrStore(name, ns); loaded {
		reg.count.Add(-1)
		return ErrNamespaceAlreadyExists
	}
	return nil
}

// Drop deletes a namespace by name.
func (nm *NamespaceManager) Drop(name string) error {
	reg := nm.registry.Load()
	if _, deleted := reg.m.LoadAndDelete(name); deleted {
		reg.count.Add(-1)
	}
	return nil
}

// Erase flushes all data in the default namespace and drops all custom namespaces.
func (nm *NamespaceManager) Erase() {
	nm.flushDefaultNs()
	nm.registry.Store(&nsRegistry{})
}

// Flush removes all keys across all namespaces including the default namespace.
func (nm *NamespaceManager) Flush() {
	nm.flushDefaultNs()
	nm.flushCustomNamespaces()
}

// flushDefaultNs flushes all keys in the default namespace.
func (nm *NamespaceManager) flushDefaultNs() {
	nm.defaultNs.Flush()
}

// flushCustomNamespaces flushes all keys in all custom namespaces.
func (nm *NamespaceManager) flushCustomNamespaces() {
	nm.registry.Load().m.Range(func(k string, v *Namespace) bool {
		v.Flush()
		return true
	})
}

// CleanExpired removes all expired keys across all namespaces including the default namespace.
func (nm *NamespaceManager) CleanExpired() {
	nm.CleanExpiredCtx(context.Background())
}

// CleanExpiredCtx removes all expired keys across all namespaces while respecting ctx cancellation.
func (nm *NamespaceManager) CleanExpiredCtx(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return
	}
	nm.defaultNs.CleanExpiredCtx(ctx)
	nm.registry.Load().m.Range(func(_ string, v *Namespace) bool {
		if ctx.Err() != nil {
			return false
		}
		v.CleanExpiredCtx(ctx)
		return true
	})
}
