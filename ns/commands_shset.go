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

func (nm *NamespaceManager) STTL(ns, key string) (int64, error) {
	if ns == "" {
		return nm.defaultNs.shset.TTL(key), nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	return n.shset.TTL(key), nil
}
