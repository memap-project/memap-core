package tests

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SAdd_and_SIsMember(t *testing.T) {
	t.Run("add members and check membership in default namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "tags", "go", 0))
		require.NoError(t, nm.SAdd("", "tags", "database", 0))

		isMem, err := nm.SIsMember("", "tags", "go")
		require.NoError(t, err)
		assert.True(t, isMem)

		isMem, err = nm.SIsMember("", "tags", "rust")
		require.NoError(t, err)
		assert.False(t, isMem)
	})

	t.Run("membership on non-existent set returns false", func(t *testing.T) {
		nm := newTestManager()
		isMem, err := nm.SIsMember("", "missing_set", "item")
		require.NoError(t, err)
		assert.False(t, isMem)
	})

	t.Run("operations in non-existent namespace return ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.SAdd("missing_ns", "tags", "go", 0)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.SIsMember("missing_ns", "tags", "go")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("add and check membership in custom namespace with isolation", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))
		require.NoError(t, nm.SAdd("custom_ns", "roles", "admin", 0))

		isMem, err := nm.SIsMember("custom_ns", "roles", "admin")
		require.NoError(t, err)
		assert.True(t, isMem)

		isMem, err = nm.SIsMember("", "roles", "admin")
		require.NoError(t, err)
		assert.False(t, isMem)
	})
}

func Test_SRemove(t *testing.T) {
	t.Run("remove existing member", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "set1", "m1", 0))
		require.NoError(t, nm.SAdd("", "set1", "m2", 0))

		require.NoError(t, nm.SRemove("", "set1", "m1"))

		isMem, err := nm.SIsMember("", "set1", "m1")
		require.NoError(t, err)
		assert.False(t, isMem)

		isMem, err = nm.SIsMember("", "set1", "m2")
		require.NoError(t, err)
		assert.True(t, isMem)
	})

	t.Run("remove non-existent member or from non-existent set succeeds as no-op", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "set1", "m1", 0))
		require.NoError(t, nm.SRemove("", "set1", "missing_member"))
		require.NoError(t, nm.SRemove("", "missing_set", "m1"))
	})

	t.Run("remove in non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.SRemove("missing_ns", "set1", "m1")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_SCard_and_SMembers(t *testing.T) {
	t.Run("missing set returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.SCard("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.SMembers("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("cardinality and members of populated set", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "fruits", "apple", 0))
		require.NoError(t, nm.SAdd("", "fruits", "banana", 0))
		require.NoError(t, nm.SAdd("", "fruits", "apple", 0)) // duplicate

		card, err := nm.SCard("", "fruits")
		require.NoError(t, err)
		assert.Equal(t, int64(2), card)

		members, err := nm.SMembers("", "fruits")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"apple", "banana"}, members)
	})

	t.Run("operations in non-existent namespace return ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.SCard("missing_ns", "fruits")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.SMembers("missing_ns", "fruits")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_SExpire_and_STTL(t *testing.T) {
	t.Run("set without expiration returns -1", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "exp_set", "item", 0))

		ttl, err := nm.STTL("", "exp_set")
		require.NoError(t, err)
		assert.Equal(t, int64(-1), ttl)
	})

	t.Run("set valid TTL", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "exp_set", "item", 0))

		require.NoError(t, nm.SExpire("", "exp_set", 60))
		ttl, err := nm.STTL("", "exp_set")
		require.NoError(t, err)
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(60))
	})

	t.Run("missing set returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.SExpire("", "missing", 10)
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.STTL("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("missing namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.SExpire("missing_ns", "exp_set", 10)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.STTL("missing_ns", "exp_set")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("expired set returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.SAdd("", "ttl_set", "val", 0))
		require.NoError(t, nm.SExpire("", "ttl_set", 1))

		time.Sleep(2100 * time.Millisecond)

		_, err := nm.SCard("", "ttl_set")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.STTL("", "ttl_set")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})
}
