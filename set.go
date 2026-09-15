package memap

func (c *Cache) SAdd(key, member string, ttl int64) error {
	return c.manager.SAdd("", key, member, ttl)
}

func (c *Cache) SRemove(key, member string) error {
	return c.manager.SRemove("", key, member)
}

func (c *Cache) SIsMember(key, member string) (bool, error) {
	return c.manager.SIsMember("", key, member)
}

func (c *Cache) SCard(key string) (int64, error) {
	return c.manager.SCard("", key)
}

func (c *Cache) SMembers(key string) ([]string, error) {
	return c.manager.SMembers("", key)
}

func (c *Cache) SExpire(key string, ttl int64) error {
	return c.manager.SExpire("", key, ttl)
}

func (c *Cache) STTL(key string) (int64, error) {
	return c.manager.STTL("", key)
}

func (ctx *Namespace) SAdd(key, member string, ttl int64) error {
	return ctx.manager.SAdd(ctx.name, key, member, ttl)
}

func (ctx *Namespace) SRemove(key, member string) error {
	return ctx.manager.SRemove(ctx.name, key, member)
}

func (ctx *Namespace) SIsMember(key, member string) (bool, error) {
	return ctx.manager.SIsMember(ctx.name, key, member)
}

func (ctx *Namespace) SCard(key string) (int64, error) {
	return ctx.manager.SCard(ctx.name, key)
}

func (ctx *Namespace) SMembers(key string) ([]string, error) {
	return ctx.manager.SMembers(ctx.name, key)
}

func (ctx *Namespace) SExpire(key string, ttl int64) error {
	return ctx.manager.SExpire(ctx.name, key, ttl)
}

func (ctx *Namespace) STTL(key string) (int64, error) {
	return ctx.manager.STTL(ctx.name, key)
}
