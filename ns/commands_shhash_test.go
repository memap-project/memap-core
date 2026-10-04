package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HSet_and_HGet(t *testing.T) {
	t.Run("HSet initializes empty hash", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HSet("", "hash1", 0))

		val, err := nm.HGet("", "hash1")
		require.NoError(t, err)
		assert.Empty(t, val)
	})

	t.Run("non-existent hash returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.HGet("", "missing_hash")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.HSet("missing_ns", "hash1", 0)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.HGet("missing_ns", "hash1")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("HSet and HGet in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))
		require.NoError(t, nm.HSet("custom_ns", "hash1", 0))

		val, err := nm.HGet("custom_ns", "hash1")
		require.NoError(t, err)
		assert.Empty(t, val)
	})
}

func Test_HFSet_and_HFGet(t *testing.T) {
	t.Run("HFSet auto-initializes hash and sets fields", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "user:1", "name", "Alice"))
		require.NoError(t, nm.HFSet("", "user:1", "role", "Admin"))

		val, err := nm.HFGet("", "user:1", "name")
		require.NoError(t, err)
		assert.Equal(t, "Alice", val)

		val, err = nm.HFGet("", "user:1", "role")
		require.NoError(t, err)
		assert.Equal(t, "Admin", val)
	})

	t.Run("update existing field", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "user:1", "name", "Alice"))
		require.NoError(t, nm.HFSet("", "user:1", "name", "Bob"))

		val, err := nm.HFGet("", "user:1", "name")
		require.NoError(t, err)
		assert.Equal(t, "Bob", val)
	})

	t.Run("field not found returns ErrFieldNotFound", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "user:1", "name", "Bob"))

		_, err := nm.HFGet("", "user:1", "missing_field")
		require.ErrorIs(t, err, ErrFieldNotFound)
	})

	t.Run("hash not found returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.HFGet("", "missing_hash", "field")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.HFSet("missing_ns", "h", "f", "v")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.HFGet("missing_ns", "h", "f")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("HFSet and HFGet in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))
		require.NoError(t, nm.HFSet("custom_ns", "user:2", "email", "bob@example.com"))

		val, err := nm.HFGet("custom_ns", "user:2", "email")
		require.NoError(t, err)
		assert.Equal(t, "bob@example.com", val)
	})
}

func Test_HFDel(t *testing.T) {
	t.Run("delete existing field", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "h", "f1", "v1"))
		require.NoError(t, nm.HFSet("", "h", "f2", "v2"))

		require.NoError(t, nm.HFDel("", "h", "f1"))

		_, err := nm.HFGet("", "h", "f1")
		require.ErrorIs(t, err, ErrFieldNotFound)

		val, err := nm.HFGet("", "h", "f2")
		require.NoError(t, err)
		assert.Equal(t, "v2", val)
	})

	t.Run("delete non-existent field or from non-existent hash succeeds", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "h", "f1", "v1"))
		require.NoError(t, nm.HFDel("", "h", "missing_field"))
		require.NoError(t, nm.HFDel("", "missing_hash", "field"))
	})

	t.Run("delete in non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.HFDel("missing_ns", "h", "f")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_HDel(t *testing.T) {
	t.Run("delete existing hash", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "h", "f1", "v1"))
		require.NoError(t, nm.HDel("", "h"))

		_, err := nm.HGet("", "h")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("delete non-existent hash succeeds as no-op", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HDel("", "missing"))
	})

	t.Run("delete in non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.HDel("missing_ns", "h")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_HExists(t *testing.T) {
	t.Run("returns false for non-existent hash and true for existing", func(t *testing.T) {
		nm := newTestManager()
		exists, err := nm.HExists("", "h")
		require.NoError(t, err)
		assert.False(t, exists)

		require.NoError(t, nm.HFSet("", "h", "f", "v"))
		exists, err = nm.HExists("", "h")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.HExists("missing_ns", "h")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_HLen_HKeys_HValues(t *testing.T) {
	t.Run("missing hash returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.HLen("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.HKeys("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.HValues("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("returns length, keys, and values of populated hash", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "profile", "first", "John"))
		require.NoError(t, nm.HFSet("", "profile", "last", "Doe"))
		require.NoError(t, nm.HFSet("", "profile", "age", "30"))

		length, err := nm.HLen("", "profile")
		require.NoError(t, err)
		assert.Equal(t, int64(3), length)

		keys, err := nm.HKeys("", "profile")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"first", "last", "age"}, keys)

		values, err := nm.HValues("", "profile")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"John", "Doe", "30"}, values)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.HLen("missing_ns", "profile")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.HKeys("missing_ns", "profile")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.HValues("missing_ns", "profile")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_HExpire_and_HTTL(t *testing.T) {
	t.Run("hash without expiration returns -1", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "h", "f", "v"))

		ttl, err := nm.HTTL("", "h")
		require.NoError(t, err)
		assert.Equal(t, int64(-1), ttl)
	})

	t.Run("expire sets TTL on existing hash", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HFSet("", "h", "f", "v"))

		require.NoError(t, nm.HExpire("", "h", 60))
		ttl, err := nm.HTTL("", "h")
		require.NoError(t, err)
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(60))
	})

	t.Run("missing hash returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.HExpire("", "missing", 10)
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.HTTL("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.HExpire("missing_ns", "h", 10)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.HTTL("missing_ns", "h")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("expired hash returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.HSet("", "exp_h", 1))
		require.NoError(t, nm.HFSet("", "exp_h", "f", "v"))
		require.NoError(t, nm.HExpire("", "exp_h", 1))

		time.Sleep(2100 * time.Millisecond)

		_, err := nm.HGet("", "exp_h")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.HTTL("", "exp_h")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})
}
