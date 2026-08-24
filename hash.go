package memap

// HGet retrieves a copy of all field-value pairs for the given key from the default namespace.
func (c *Cache) HGet(key string) (map[string]string, error) {
	return c.manager.HGet("", key)
}

// HSet creates or overwrites a hash for the given key with optional TTL in the default namespace.
func (c *Cache) HSet(key string, ttl int64) error {
	return c.manager.HSet("", key, ttl)
}

// HDel removes a hash for the given key from the default namespace.
func (c *Cache) HDel(key string) error {
	return c.manager.HDel("", key)
}

// HExpire sets the expiration time for the hash of the given key in the default namespace.
func (c *Cache) HExpire(key string, ttl int64) error {
	return c.manager.HExpire("", key, ttl)
}

// HTTL returns the time-to-live of the hash for the given key in seconds from the default namespace.
func (c *Cache) HTTL(key string) (int64, error) {
	return c.manager.HTTL("", key)
}

// HExists checks whether an unexpired hash exists for the given key in the default namespace.
func (c *Cache) HExists(key string) (bool, error) {
	return c.manager.HExists("", key)
}

// HLen returns the number of fields in the hash for the given key from the default namespace.
func (c *Cache) HLen(key string) (int64, error) {
	return c.manager.HLen("", key)
}

// HKeys returns all field names in the hash for the given key from the default namespace.
func (c *Cache) HKeys(key string) ([]string, error) {
	return c.manager.HKeys("", key)
}

// HValues returns all field values in the hash for the given key from the default namespace.
func (c *Cache) HValues(key string) ([]string, error) {
	return c.manager.HValues("", key)
}

// HFGet retrieves the value of the specified field in the hash for the given key from the default namespace.
func (c *Cache) HFGet(key, field string) (string, error) {
	return c.manager.HFGet("", key, field)
}

// HFSet sets or updates a field in the hash for the given key in the default namespace.
func (c *Cache) HFSet(key, field, value string) error {
	return c.manager.HFSet("", key, field, value)
}

// HFDel removes a field from the hash for the given key in the default namespace.
func (c *Cache) HFDel(key, field string) error {
	return c.manager.HFDel("", key, field)
}

// HGet retrieves a copy of all field-value pairs for the given key from this namespace.
func (ctx *Namespace) HGet(key string) (map[string]string, error) {
	return ctx.manager.HGet(ctx.name, key)
}

// HSet creates or overwrites a hash for the given key with optional TTL in this namespace.
func (ctx *Namespace) HSet(key string, ttl int64) error {
	return ctx.manager.HSet(ctx.name, key, ttl)
}

// HDel removes a hash for the given key from this namespace.
func (ctx *Namespace) HDel(key string) error {
	return ctx.manager.HDel(ctx.name, key)
}

// HExpire sets the expiration time for the hash of the given key in this namespace.
func (ctx *Namespace) HExpire(key string, ttl int64) error {
	return ctx.manager.HExpire(ctx.name, key, ttl)
}

// HTTL returns the time-to-live of the hash for the given key in seconds from this namespace.
func (ctx *Namespace) HTTL(key string) (int64, error) {
	return ctx.manager.HTTL(ctx.name, key)
}

// HExists checks whether an unexpired hash exists for the given key in this namespace.
func (ctx *Namespace) HExists(key string) (bool, error) {
	return ctx.manager.HExists(ctx.name, key)
}

// HLen returns the number of fields in the hash for the given key from this namespace.
func (ctx *Namespace) HLen(key string) (int64, error) {
	return ctx.manager.HLen(ctx.name, key)
}

// HKeys returns all field names in the hash for the given key from this namespace.
func (ctx *Namespace) HKeys(key string) ([]string, error) {
	return ctx.manager.HKeys(ctx.name, key)
}

// HValues returns all field values in the hash for the given key from this namespace.
func (ctx *Namespace) HValues(key string) ([]string, error) {
	return ctx.manager.HValues(ctx.name, key)
}

// HFGet retrieves the value of the specified field in the hash for the given key from this namespace.
func (ctx *Namespace) HFGet(key, field string) (string, error) {
	return ctx.manager.HFGet(ctx.name, key, field)
}

// HFSet sets or updates a field in the hash for the given key in this namespace.
func (ctx *Namespace) HFSet(key, field, value string) error {
	return ctx.manager.HFSet(ctx.name, key, field, value)
}

// HFDel removes a field from the hash for the given key in this namespace.
func (ctx *Namespace) HFDel(key, field string) error {
	return ctx.manager.HFDel(ctx.name, key, field)
}
