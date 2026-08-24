package memap

// Get retrieves the value of the key from the default namespace.
func (c *Cache) Get(key string) (string, error) {
	return c.manager.Get("", key)
}

// Set stores a key-value pair with optional TTL in the default namespace.
func (c *Cache) Set(key, value string, ttl int64) error {
	return c.manager.Set("", key, value, ttl)
}

// Del removes a key from the default namespace.
func (c *Cache) Del(key string) error {
	return c.manager.Del("", key)
}

// Expire sets the expiration time for the key in the default namespace.
func (c *Cache) Expire(key string, ttl int64) error {
	return c.manager.Expire("", key, ttl)
}

// TTL returns the time-to-live of the key in seconds from the default namespace.
func (c *Cache) TTL(key string) (int64, error) {
	return c.manager.TTL("", key)
}

// Get retrieves the value of the key from this namespace.
func (ctx *Namespace) Get(key string) (string, error) {
	return ctx.manager.Get(ctx.name, key)
}

// Set stores a key-value pair with optional TTL in this namespace.
func (ctx *Namespace) Set(key, value string, ttl int64) error {
	return ctx.manager.Set(ctx.name, key, value, ttl)
}

// Del removes a key from this namespace.
func (ctx *Namespace) Del(key string) error {
	return ctx.manager.Del(ctx.name, key)
}

// Expire sets the expiration time for the key in this namespace.
func (ctx *Namespace) Expire(key string, ttl int64) error {
	return ctx.manager.Expire(ctx.name, key, ttl)
}

// TTL returns the time-to-live of the key in seconds from this namespace.
func (ctx *Namespace) TTL(key string) (int64, error) {
	return ctx.manager.TTL(ctx.name, key)
}
