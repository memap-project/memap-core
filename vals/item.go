package vals

// Item represents a string value with optional monotonic expiration.
type Item struct {
	expiry
	value string
}

// NewItem creates a new Item with the given value.
func NewItem(value string) *Item {
	return &Item{
		value: value,
	}
}

// GetValue returns the string value of the Item.
func (i *Item) GetValue() string {
	return i.value
}
