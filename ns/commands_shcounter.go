package ns

// CInit initializes a counter with the given limit and optional TTL in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyAlreadyExists] if a counter already exists for the given key.
func (nm *NamespaceManager) CInit(ns, key string, limit, ttl int64) error {
	if ns == "" {
		ok := nm.defaultNs.shcounter.Init(key, limit, ttl)
		if !ok {
			return ErrKeyAlreadyExists
		}
		return nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	ok := n.shcounter.Init(key, limit, ttl)
	if !ok {
		return ErrKeyAlreadyExists
	}
	return nil
}

// CSLimit sets or updates the upper limit of a counter in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the counter does not exist or is expired.
func (nm *NamespaceManager) CSLimit(ns, key string, limit int64) error {
	if ns == "" {
		ok := nm.defaultNs.shcounter.SetLimit(key, limit)
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	ok := n.shcounter.SetLimit(key, limit)
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}

// CGLimit returns the upper limit of a counter in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the counter does not exist or is expired.
func (nm *NamespaceManager) CGLimit(ns, key string) (int64, error) {
	if ns == "" {
		l, ok := nm.defaultNs.shcounter.GetLimit(key)
		if !ok {
			return l, ErrKeyNotFound
		}
		return l, nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	l, ok := n.shcounter.GetLimit(key)
	if !ok {
		return l, ErrKeyNotFound
	}
	return l, nil
}

// CGet returns the value of a counter in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the counter does not exist or is expired.
func (nm *NamespaceManager) CGet(ns, key string) (int64, error) {
	if ns == "" {
		count, ok := nm.defaultNs.shcounter.Get(key)
		if !ok {
			return count, ErrKeyNotFound
		}
		return count, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	count, ok := n.shcounter.Get(key)
	if !ok {
		return count, ErrKeyNotFound
	}
	return count, nil
}

// CDel removes a counter from the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) CDel(ns, key string) error {
	if ns == "" {
		nm.defaultNs.shcounter.Delete(key)
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shcounter.Delete(key)
	return nil
}

// CExpire sets the expiration time for the counter in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the counter does not exist or is expired.
func (nm *NamespaceManager) CExpire(ns, key string, ttl int64) error {
	if ns == "" {
		ok := nm.defaultNs.shcounter.Expire(key, ttl)
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shcounter.Expire(key, ttl)
	return nil
}

// CTTL returns the remaining time-to-live of the counter in seconds from the specified namespace.
// If ns is empty, the default namespace is used.
// Returns -1 if the counter has no expiration time.
// Returns -2 if the counter does not exist or is expired.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) CTTL(ns, key string) (int64, error) {
	if ns == "" {
		count, ok := nm.defaultNs.shcounter.TTL(key)
		if !ok {
			return count, ErrKeyNotFound
		}
		return count, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return -2, ErrNamespaceNotFound
	}
	count, ok := n.shcounter.TTL(key)
	if !ok {
		return count, ErrKeyNotFound
	}
	return count, nil
}

// CIncrBy increments the counter by alpha in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the counter does not exist or is expired.
// Returns [ErrLimitExceeded] if the increment exceeds limit.
func (nm *NamespaceManager) CIncrBy(ns, key string, alpha int64) (int64, error) {
	if ns == "" {
		count, status := nm.defaultNs.shcounter.IncrBy(key, alpha)
		return count, statusToError(status)
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	count, status := n.shcounter.IncrBy(key, alpha)
	return count, statusToError(status)
}

// CDecrBy decrements the counter by alpha in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the counter does not exist or is expired.
// Returns [ErrLimitExceeded] if the decrement would result in a negative value.
func (nm *NamespaceManager) CDecrBy(ns, key string, alpha int64) (int64, error) {
	if ns == "" {
		count, status := nm.defaultNs.shcounter.DecrBy(key, alpha)
		return count, statusToError(status)
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	count, status := n.shcounter.DecrBy(key, alpha)
	return count, statusToError(status)
}
