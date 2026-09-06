package memap

// CInit initializes a counter with the given limit and optional TTL in the default namespace.
func (c *Cache) CInit(key string, limit, ttl int64) error {
	return c.manager.CInit("", key, limit, ttl)
}

// CSLimit sets or updates the upper limit of a counter in the default namespace.
func (c *Cache) CSLimit(key string, limit int64) error {
	return c.manager.CSLimit("", key, limit)
}

// CGLimit returns the upper limit of a counter in the default namespace.
func (c *Cache) CGLimit(key string) (int64, error) {
	return c.manager.CGLimit("", key)
}

// CGet returns the value of a counter in the default namespace.
func (c *Cache) CGet(key string) (int64, error) {
	return c.manager.CGet("", key)
}

// CDel removes a counter from the default namespace.
func (c *Cache) CDel(key string) error {
	return c.manager.CDel("", key)
}

// CExpire sets the expiration time for the counter in the default namespace.
func (c *Cache) CExpire(key string, ttl int64) error {
	return c.manager.CExpire("", key, ttl)
}

// CTTL returns the time-to-live of the counter in seconds from the default namespace.
func (c *Cache) CTTL(key string) (int64, error) {
	return c.manager.CTTL("", key)
}

// CIncrBy increments the counter by alpha in the default namespace.
func (c *Cache) CIncrBy(key string, alpha int64) (int64, error) {
	return c.manager.CIncrBy("", key, alpha)
}

// CDecrBy decrements the counter by alpha in the default namespace.
func (c *Cache) CDecrBy(key string, alpha int64) (int64, error) {
	return c.manager.CDecrBy("", key, alpha)
}

// CInit initializes a counter with the given limit and optional TTL in this namespace.
func (ctx *Namespace) CInit(key string, limit, ttl int64) error {
	return ctx.manager.CInit(ctx.name, key, limit, ttl)
}

// CSLimit sets or updates the upper limit of a counter in this namespace.
func (ctx *Namespace) CSLimit(key string, limit int64) error {
	return ctx.manager.CSLimit(ctx.name, key, limit)
}

// CGLimit returns the upper limit of a counter in this namespace.
func (ctx *Namespace) CGLimit(key string) (int64, error) {
	return ctx.manager.CGLimit(ctx.name, key)
}

// CGet returns the value of a counter in this namespace.
func (ctx *Namespace) CGet(key string) (int64, error) {
	return ctx.manager.CGet(ctx.name, key)
}

// CDel removes a counter from this namespace.
func (ctx *Namespace) CDel(key string) error {
	return ctx.manager.CDel(ctx.name, key)
}

// CExpire sets the expiration time for the counter in this namespace.
func (ctx *Namespace) CExpire(key string, ttl int64) error {
	return ctx.manager.CExpire(ctx.name, key, ttl)
}

// CTTL returns the time-to-live of the counter in seconds from this namespace.
func (ctx *Namespace) CTTL(key string) (int64, error) {
	return ctx.manager.CTTL(ctx.name, key)
}

// CIncrBy increments the counter by alpha in this namespace.
func (ctx *Namespace) CIncrBy(key string, alpha int64) (int64, error) {
	return ctx.manager.CIncrBy(ctx.name, key, alpha)
}

// CDecrBy decrements the counter by alpha in this namespace.
func (ctx *Namespace) CDecrBy(key string, alpha int64) (int64, error) {
	return ctx.manager.CDecrBy(ctx.name, key, alpha)
}
