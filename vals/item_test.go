package vals

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItem_NewItem_and_GetValue(t *testing.T) {
	t.Run("creates item with value and default state", func(t *testing.T) {
		item := NewItem("hello world")
		require.NotNil(t, item)

		assert.Equal(t, "hello world", item.GetValue())
		assert.False(t, item.IsExpired())
		assert.Equal(t, int64(0), item.TTL())
	})
}

func TestItem_Expiration(t *testing.T) {
	t.Run("negative TTL rejected", func(t *testing.T) {
		item := NewItem("data")
		assert.False(t, item.Expire(-1))
		assert.False(t, item.IsExpired())
		assert.Equal(t, int64(0), item.TTL())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		item := NewItem("data")
		assert.True(t, item.Expire(120))
		assert.False(t, item.IsExpired())
		assert.Greater(t, item.TTL(), int64(0))
		assert.LessOrEqual(t, item.TTL(), int64(120))
	})

	t.Run("detects expired state", func(t *testing.T) {
		item := NewItem("data")
		item.expiresAt.Store(time.Now().Unix() - 1)
		assert.True(t, item.IsExpired())
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		item := NewItem("data")
		item.expiresAt.Store(time.Now().Unix() - 1)
		assert.False(t, item.Expire(60))
	})
}
