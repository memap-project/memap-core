package vals

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItem_NewItem_and_GetValue(t *testing.T) {
	t.Run("creates item with value and default state", func(t *testing.T) {
		item := NewItem("hello world")
		require.NotNil(t, item)

		assert.Equal(t, "hello world", item.GetValue())
		assert.False(t, item.IsExpired())
		assert.Equal(t, int64(-1), item.TTL())
	})
}

func TestItem_Expiration(t *testing.T) {
	t.Run("non-positive and overflowing TTL rejected", func(t *testing.T) {
		item := NewItem("data")
		assert.False(t, item.Expire(-1))
		assert.False(t, item.Expire(0))
		assert.False(t, item.Expire(math.MaxInt64))
		assert.False(t, item.IsExpired())
		assert.Equal(t, int64(-1), item.TTL())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		item := NewItem("data")
		assert.True(t, item.Expire(120))
		assert.False(t, item.IsExpired())
		assert.Greater(t, item.TTL(), int64(0))
		assert.LessOrEqual(t, item.TTL(), int64(120))
	})

	t.Run("detects expired state and boundary now == exp returns -2 not -1", func(t *testing.T) {
		item := NewItem("data")
		now := MonotonicNow()
		item.expiresAt.Store(now)
		assert.True(t, item.IsExpiredAt(now))
		assert.Equal(t, int64(-2), item.TTLAt(now))

		item.expiresAt.Store(now + 1)
		assert.False(t, item.IsExpiredAt(now))
		assert.Equal(t, int64(1), item.TTLAt(now))
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		item := NewItem("data")
		item.expiresAt.Store(MonotonicNow() - 1)
		assert.True(t, item.IsExpired())
		assert.Equal(t, int64(-2), item.TTL())
		assert.False(t, item.Expire(60))
	})
}
