package vals

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpiry(t *testing.T) {
	t.Run("default state has no expiration", func(t *testing.T) {
		var e expiry
		assert.False(t, e.IsExpired())
		assert.Equal(t, int64(-1), e.TTL())
	})

	t.Run("non-positive and overflowing TTL rejected", func(t *testing.T) {
		var e expiry
		assert.False(t, e.Expire(-1))
		assert.False(t, e.Expire(0))
		assert.False(t, e.Expire(math.MaxInt64))
		assert.False(t, e.IsExpired())
		assert.Equal(t, int64(-1), e.TTL())
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		var e expiry
		assert.True(t, e.Expire(120))
		assert.False(t, e.IsExpired())
		assert.Greater(t, e.TTL(), int64(0))
		assert.LessOrEqual(t, e.TTL(), int64(120))
	})

	t.Run("boundary now == exp returns -2 not -1", func(t *testing.T) {
		var e expiry
		now := MonotonicNow()
		e.expiresAt.Store(now)
		assert.True(t, e.IsExpiredAt(now))
		assert.Equal(t, int64(-2), e.TTLAt(now))

		e.expiresAt.Store(now + 1)
		assert.False(t, e.IsExpiredAt(now))
		assert.Equal(t, int64(1), e.TTLAt(now))
	})

	t.Run("fails to expire when already expired", func(t *testing.T) {
		var e expiry
		e.expiresAt.Store(MonotonicNow() - 1)
		assert.True(t, e.IsExpired())
		assert.Equal(t, int64(-2), e.TTL())
		assert.False(t, e.Expire(60))
	})
}
