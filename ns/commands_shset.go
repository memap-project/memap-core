package ns

// SAdd adds a member to a set in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) SAdd(ns, key, member string, ttl int64) error {
	if ns == "" {
		nm.defaultNs.shset.Add(key, member, ttl)
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shset.Add(key, member, ttl)
	return nil
}

// SRemove removes a member from a set in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) SRemove(ns, key, member string) error {
	if ns == "" {
		nm.defaultNs.shset.Remove(key, member)
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shset.Remove(key, member)
	return nil
}

// SIsMember returns true if the member is a member of the set in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) SIsMember(ns, key, member string) (bool, error) {
	if ns == "" {
		return nm.defaultNs.shset.IsMember(key, member), nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return false, ErrNamespaceNotFound
	}
	return n.shset.IsMember(key, member), nil
}

// SCard returns the number of members in the set in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the set does not exist or is expired.
func (nm *NamespaceManager) SCard(ns, key string) (int64, error) {
	if ns == "" {
		len, ok := nm.defaultNs.shset.Card(key)
		if !ok {
			return 0, ErrKeyNotFound
		}
		return len, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	len, ok := n.shset.Card(key)
	if !ok {
		return 0, ErrKeyNotFound
	}
	return len, nil
}

// SMembers returns all members of the set in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the set does not exist or is expired.
func (nm *NamespaceManager) SMembers(ns, key string) ([]string, error) {
	if ns == "" {
		members, ok := nm.defaultNs.shset.Members(key)
		if !ok {
			return []string{}, ErrKeyNotFound
		}
		return members, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return []string{}, ErrNamespaceNotFound
	}
	members, ok := n.shset.Members(key)
	if !ok {
		return []string{}, ErrKeyNotFound
	}
	return members, nil
}

// SExpire sets the expiration time for the set of the given key in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the set does not exist or is expired.
func (nm *NamespaceManager) SExpire(ns, key string, ttl int64) error {
	if ns == "" {
		ok := nm.defaultNs.shset.Expire(key, ttl)
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	ok := n.shset.Expire(key, ttl)
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}

// STTL returns the time-to-live of the set for the given key in seconds from the specified namespace.
// If ns is empty, the default namespace is used.
// Returns -1 if the set has no expiration time.
// Returns -2 if the set does not exist or is expired.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the set does not exist or is expired.
func (nm *NamespaceManager) STTL(ns, key string) (int64, error) {
	if ns == "" {
		ttl := nm.defaultNs.shset.TTL(key)
		if ttl == -2 {
			return ttl, ErrKeyNotFound
		}
		return ttl, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	ttl := n.shset.TTL(key)
	if ttl == -2 {
		return ttl, ErrKeyNotFound
	}
	return ttl, nil
}
