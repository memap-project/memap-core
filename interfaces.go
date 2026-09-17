package memap

// Core is the unified interface for the Memap in-memory database.
type Core interface {
	KVStore
	HashStore
	CounterStore
	RingBufferStore
	NamespaceOps
	Close()
}

// KVStore defines key-value operations.
type KVStore interface {
	Get(key string) (string, error)
	Set(key, value string, ttl int64) error
	Del(key string) error
	Expire(key string, ttl int64) error
	TTL(key string) (int64, error)
}

// HashStore defines hash map operations.
type HashStore interface {
	HGet(key string) (map[string]string, error)
	HSet(key string, ttl int64) error
	HDel(key string) error
	HExpire(key string, ttl int64) error
	HTTL(key string) (int64, error)
	HExists(key string) (bool, error)
	HLen(key string) (int64, error)
	HKeys(key string) ([]string, error)
	HValues(key string) ([]string, error)
	HFGet(key, field string) (string, error)
	HFSet(key, field, value string) error
	HFDel(key, field string) error
}

// CounterStore defines counter operations.
type CounterStore interface {
	CSLimit(key string, limit int64) error
	CGLimit(key string) (int64, error)
	CGet(key string) (int64, error)
	CDel(key string) error
	CExpire(key string, ttl int64) error
	CTTL(key string) (int64, error)
	CIncrBy(key string, alpha int64) (int64, error)
	CDecrBy(key string, alpha int64) (int64, error)
}

// RingBufferStore defines ring buffer operations.
type RingBufferStore interface {
	BInit(key string, capacity, ttl int64) error
	BPush(key, value string) error
	BPop(key string) (string, error)
	BAt(key string, index int64) (string, error)
	BSlice(key string) ([]string, error)
	BPeek(key string) (string, error)
	BBack(key string) (string, error)
	BCap(key string) (int64, error)
	BLen(key string) (int64, error)
	BReset(key string) error
	BDel(key string) error
	BExpire(key string, ttl int64) error
	BTTL(key string) (int64, error)
}

// SetStore defines set operations.
type SetStore interface {
	SAdd(key, member string, ttl int64) error
	SRemove(key, member string) error
	SIsMember(key, member string) (bool, error)
	SCard(key string) (int64, error)
	SMembers(key string) ([]string, error)
	SExpire(key string, ttl int64) error
	STTL(key string) (int64, error)
}

// NamespaceOps defines namespace management operations and scoping.
type NamespaceOps interface {
	CreateNamespace(name string) error
	DropNamespace(name string) error
	Erase()
	Flush()
	WithNamespace(name string) *Namespace
}
