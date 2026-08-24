package memap

// CreateNamespace creates a new isolated namespace with the given name.
func (c *Cache) CreateNamespace(name string) error {
	return c.manager.Create(name)
}

// DropNamespace deletes an isolated namespace and all its data by name.
func (c *Cache) DropNamespace(name string) error {
	return c.manager.Drop(name)
}

// Erase flushes all data in the default namespace and drops all custom namespaces.
func (c *Cache) Erase() {
	c.manager.Erase()
}

// Flush removes all keys across all namespaces including the default namespace.
func (c *Cache) Flush() {
	c.manager.Flush()
}
