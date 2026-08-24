package ns

// BInit initializes a ring buffer with the given capacity and optional TTL in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyAlreadyExists] if a counter already exists for the given key.
func (nm *NamespaceManager) BInit(ns, key string, capacity, ttl int64) error {
	if ns == "" {
		ok := nm.defaultNs.shrbuffer.Init(key, capacity, ttl)
		if !ok {
			return ErrKeyAlreadyExists
		}
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	ok := n.shrbuffer.Init(key, capacity, ttl)
	if !ok {
		return ErrKeyAlreadyExists
	}
	return nil
}

// BPush pushes a value to the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist.
func (nm *NamespaceManager) BPush(ns, key, value string) error {
	if ns == "" {
		ok := nm.defaultNs.shrbuffer.Push(key, value)
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	ok := n.shrbuffer.Push(key, value)
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}

// BPop pops a value from the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist or is expired.
// Returns [ErrBufferEmpty] if the ring buffer is empty.
func (nm *NamespaceManager) BPop(ns, key string) (string, error) {
	if ns == "" {
		value, status := nm.defaultNs.shrbuffer.Pop(key)
		return value, statusToError(status)
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return "", ErrNamespaceNotFound
	}
	value, status := n.shrbuffer.Pop(key)
	return value, statusToError(status)
}

// BAt returns the value at the given index in the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist or is expired.
// Returns [ErrIndexOutOfBounds] if the index is out of range.
func (nm *NamespaceManager) BAt(ns, key string, index int64) (string, error) {
	if ns == "" {
		value, status := nm.defaultNs.shrbuffer.At(key, index)
		return value, statusToError(status)
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return "", ErrNamespaceNotFound
	}
	value, status := n.shrbuffer.At(key, index)
	return value, statusToError(status)
}

// BSlice returns all values in the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist.
func (nm *NamespaceManager) BSlice(ns, key string) ([]string, error) {
	if ns == "" {
		values, ok := nm.defaultNs.shrbuffer.Slice(key)
		if !ok {
			return nil, ErrKeyNotFound
		}
		return values, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return nil, ErrNamespaceNotFound
	}
	values, ok := n.shrbuffer.Slice(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return values, nil
}

// BPeek returns the value at the head (oldest) of the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist or is expired.
// Returns [ErrBufferEmpty] if the ring buffer is empty.
func (nm *NamespaceManager) BPeek(ns, key string) (string, error) {
	if ns == "" {
		value, status := nm.defaultNs.shrbuffer.Peek(key)
		return value, statusToError(status)
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return "", ErrNamespaceNotFound
	}
	value, status := n.shrbuffer.Peek(key)
	return value, statusToError(status)
}

// BBack returns the value at the tail (newest) of the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist or is expired.
// Returns [ErrBufferEmpty] if the ring buffer is empty.
func (nm *NamespaceManager) BBack(ns, key string) (string, error) {
	if ns == "" {
		value, status := nm.defaultNs.shrbuffer.Back(key)
		return value, statusToError(status)
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return "", ErrNamespaceNotFound
	}
	value, status := n.shrbuffer.Back(key)
	return value, statusToError(status)
}

// BCap returns the capacity of the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist.
func (nm *NamespaceManager) BCap(ns, key string) (int64, error) {
	if ns == "" {
		cap, ok := nm.defaultNs.shrbuffer.Cap(key)
		if !ok {
			return cap, ErrKeyNotFound
		}
		return cap, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	cap, ok := n.shrbuffer.Cap(key)
	if !ok {
		return cap, ErrKeyNotFound
	}
	return cap, nil
}

// BLen returns the number of values in the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist.
func (nm *NamespaceManager) BLen(ns, key string) (int64, error) {
	if ns == "" {
		len, ok := nm.defaultNs.shrbuffer.Len(key)
		if !ok {
			return len, ErrKeyNotFound
		}
		return len, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	len, ok := n.shrbuffer.Len(key)
	if !ok {
		return len, ErrKeyNotFound
	}
	return len, nil
}

// BReset resets the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) BReset(ns, key string) error {
	if ns == "" {
		nm.defaultNs.shrbuffer.Reset(key)
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shrbuffer.Reset(key)
	return nil
}

// BDel deletes the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
func (nm *NamespaceManager) BDel(ns, key string) error {
	if ns == "" {
		nm.defaultNs.shrbuffer.Delete(key)
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	n.shrbuffer.Delete(key)
	return nil
}

// BExpire sets the expiration time of the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist.
func (nm *NamespaceManager) BExpire(ns, key string, ttl int64) error {
	if ns == "" {
		ok := nm.defaultNs.shrbuffer.Expire(key, ttl)
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return ErrNamespaceNotFound
	}
	ok := n.shrbuffer.Expire(key, ttl)
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}

// BTTL returns the time-to-live of the ring buffer in the specified namespace.
// If ns is empty, the default namespace is used.
// Returns [ErrNamespaceNotFound] if the namespace does not exist.
// Returns [ErrKeyNotFound] if the key does not exist.
func (nm *NamespaceManager) BTTL(ns, key string) (int64, error) {
	if ns == "" {
		ttl, ok := nm.defaultNs.shrbuffer.TTL(key)
		if !ok {
			return ttl, ErrKeyNotFound
		}
		return ttl, nil
	}
	n, exists := nm.GetNs(ns)
	if !exists {
		return 0, ErrNamespaceNotFound
	}
	ttl, ok := n.shrbuffer.TTL(key)
	if !ok {
		return ttl, ErrKeyNotFound
	}
	return ttl, nil
}
