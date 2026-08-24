package memap

// BInit initializes a ring buffer with the given capacity and optional TTL in the default namespace.
func (c *Cache) BInit(key string, capacity, ttl int64) error {
	return c.manager.BInit("", key, capacity, ttl)
}

// BPush pushes a value to the ring buffer in the default namespace.
func (c *Cache) BPush(key, value string) error {
	return c.manager.BPush("", key, value)
}

// BPop pops a value from the ring buffer in the default namespace.
func (c *Cache) BPop(key string) (string, error) {
	return c.manager.BPop("", key)
}

// BAt returns the value at the given index in the ring buffer in the default namespace.
func (c *Cache) BAt(key string, index int64) (string, error) {
	return c.manager.BAt("", key, index)
}

// BSlice returns all values in the ring buffer in the default namespace.
func (c *Cache) BSlice(key string) ([]string, error) {
	return c.manager.BSlice("", key)
}

// BPeek returns the value at the head (oldest) of the ring buffer in the default namespace.
func (c *Cache) BPeek(key string) (string, error) {
	return c.manager.BPeek("", key)
}

// BBack returns the value at the tail (newest) of the ring buffer in the default namespace.
func (c *Cache) BBack(key string) (string, error) {
	return c.manager.BBack("", key)
}

// BCap returns the capacity of the ring buffer in the default namespace.
func (c *Cache) BCap(key string) (int64, error) {
	return c.manager.BCap("", key)
}

// BLen returns the number of values in the ring buffer in the default namespace.
func (c *Cache) BLen(key string) (int64, error) {
	return c.manager.BLen("", key)
}

// BReset resets the ring buffer in the default namespace.
func (c *Cache) BReset(key string) error {
	return c.manager.BReset("", key)
}

// BDel deletes the ring buffer in the default namespace.
func (c *Cache) BDel(key string) error {
	return c.manager.BDel("", key)
}

// BExpire sets the expiration time of the ring buffer in the default namespace.
func (c *Cache) BExpire(key string, ttl int64) error {
	return c.manager.BExpire("", key, ttl)
}

// BTTL returns the time-to-live of the ring buffer in the default namespace.
func (c *Cache) BTTL(key string) (int64, error) {
	return c.manager.BTTL("", key)
}

// BInit initializes a ring buffer with the given capacity and optional TTL in this namespace.
func (ctx *Namespace) BInit(key string, capacity, ttl int64) error {
	return ctx.manager.BInit(ctx.name, key, capacity, ttl)
}

// BPush pushes a value to the ring buffer in this namespace.
func (ctx *Namespace) BPush(key, value string) error {
	return ctx.manager.BPush(ctx.name, key, value)
}

// BPop pops a value from the ring buffer in this namespace.
func (ctx *Namespace) BPop(key string) (string, error) {
	return ctx.manager.BPop(ctx.name, key)
}

// BAt returns the value at the given index in the ring buffer in this namespace.
func (ctx *Namespace) BAt(key string, index int64) (string, error) {
	return ctx.manager.BAt(ctx.name, key, index)
}

// BSlice returns all values in the ring buffer in this namespace.
func (ctx *Namespace) BSlice(key string) ([]string, error) {
	return ctx.manager.BSlice(ctx.name, key)
}

// BPeek returns the value at the head (oldest) of the ring buffer in this namespace.
func (ctx *Namespace) BPeek(key string) (string, error) {
	return ctx.manager.BPeek(ctx.name, key)
}

// BBack returns the value at the tail (newest) of the ring buffer in this namespace.
func (ctx *Namespace) BBack(key string) (string, error) {
	return ctx.manager.BBack(ctx.name, key)
}

// BCap returns the capacity of the ring buffer in this namespace.
func (ctx *Namespace) BCap(key string) (int64, error) {
	return ctx.manager.BCap(ctx.name, key)
}

// BLen returns the number of values in the ring buffer in this namespace.
func (ctx *Namespace) BLen(key string) (int64, error) {
	return ctx.manager.BLen(ctx.name, key)
}

// BReset resets the ring buffer in this namespace.
func (ctx *Namespace) BReset(key string) error {
	return ctx.manager.BReset(ctx.name, key)
}

// BDel deletes the ring buffer in this namespace.
func (ctx *Namespace) BDel(key string) error {
	return ctx.manager.BDel(ctx.name, key)
}

// BExpire sets the expiration time of the ring buffer in this namespace.
func (ctx *Namespace) BExpire(key string, ttl int64) error {
	return ctx.manager.BExpire(ctx.name, key, ttl)
}

// BTTL returns the time-to-live of the ring buffer in this namespace.
func (ctx *Namespace) BTTL(key string) (int64, error) {
	return ctx.manager.BTTL(ctx.name, key)
}
