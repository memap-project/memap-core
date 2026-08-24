package ns

// Get retrieves the value of the key from the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist or is expired.
func (nm *NamespaceManager) Get(ns, key string) (string, error) {
	if ns == "" {
		val, ok := nm.defaultNs.shmap.Get(key)
		if !ok {
			return val, ErrKeyNotFound
		}
		return val, nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return "", ErrNamespaceNotFound
	}

	val, ok := n.shmap.Get(key)
	if !ok {
		return val, ErrKeyNotFound
	}
	return val, nil
}

// Set stores a key-value pair with optional TTL in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) Set(ns, key, value string, ttl int64) error {
	if ns == "" {
		nm.defaultNs.shmap.Set(key, value, ttl)
		return nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}

	n.shmap.Set(key, value, ttl)
	return nil
}

// Del removes a key from the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) Del(ns, key string) error {
	if ns == "" {
		nm.defaultNs.shmap.Delete(key)
		return nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shmap.Delete(key)
	return nil
}

// Expire sets the expiration time for the key in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist or is expired.
func (nm *NamespaceManager) Expire(ns, key string, ttl int64) error {
	if ns == "" {
		if ok := nm.defaultNs.shmap.Expire(key, ttl); !ok {
			return ErrKeyNotFound
		}
		return nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}

	if ok := n.shmap.Expire(key, ttl); !ok {
		return ErrKeyNotFound
	}
	return nil
}

// TTL returns the time-to-live of the key in seconds from the specified namespace.
// If ns is empty, the default namespace is used.
// Returns -1 if the key has no expiration time.
// Returns -2 if the key does not exist or is expired.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) TTL(ns, key string) (int64, error) {
	if ns == "" {
		return nm.defaultNs.shmap.TTL(key), nil
	}

	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	return n.shmap.TTL(key), nil
}
