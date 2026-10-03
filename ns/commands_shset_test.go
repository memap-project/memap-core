package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SAdd_and_SIsMember(t *testing.T) {
	nm := newTestManager()

	// Add to default namespace
	require.NoError(t, nm.SAdd("", "tags", "go", 0))
	require.NoError(t, nm.SAdd("", "tags", "database", 0))

	isMem, err := nm.SIsMember("", "tags", "go")
	require.NoError(t, err)
	assert.True(t, isMem)

	isMem, err = nm.SIsMember("", "tags", "rust")
	require.NoError(t, err)
	assert.False(t, isMem)

	// IsMember on non-existent set returns false, nil
	isMem, err = nm.SIsMember("", "missing_set", "item")
	require.NoError(t, err)
	assert.False(t, isMem)

	// Non-existent namespace
	err = nm.SAdd("missing_ns", "tags", "go", 0)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.SIsMember("missing_ns", "tags", "go")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	require.NoError(t, nm.SAdd("custom_ns", "roles", "admin", 0))

	isMem, err = nm.SIsMember("custom_ns", "roles", "admin")
	require.NoError(t, err)
	assert.True(t, isMem)

	// Isolation between namespaces
	isMem, err = nm.SIsMember("", "roles", "admin")
	require.NoError(t, err)
	assert.False(t, isMem)
}

func Test_SRemove(t *testing.T) {
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

	// SRemove on non-existent member or set does not error
	require.NoError(t, nm.SRemove("", "set1", "missing_member"))
	require.NoError(t, nm.SRemove("", "missing_set", "m1"))

	// Non-existent namespace
	err = nm.SRemove("missing_ns", "set1", "m1")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_SCard_and_SMembers(t *testing.T) {
	nm := newTestManager()

	// Missing set
	_, err := nm.SCard("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.SMembers("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Add members (including duplicate)
	require.NoError(t, nm.SAdd("", "fruits", "apple", 0))
	require.NoError(t, nm.SAdd("", "fruits", "banana", 0))
	require.NoError(t, nm.SAdd("", "fruits", "apple", 0))

	card, err := nm.SCard("", "fruits")
	require.NoError(t, err)
	assert.Equal(t, int64(2), card)

	members, err := nm.SMembers("", "fruits")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"apple", "banana"}, members)

	// Non-existent namespace
	_, err = nm.SCard("missing_ns", "fruits")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.SMembers("missing_ns", "fruits")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_SExpire_and_STTL(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.SAdd("", "exp_set", "item", 0))

	// Set without expiration returns -1
	ttl, err := nm.STTL("", "exp_set")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), ttl)

	// Set expiration
	require.NoError(t, nm.SExpire("", "exp_set", 60))
	ttl, err = nm.STTL("", "exp_set")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))
	assert.LessOrEqual(t, ttl, int64(60))

	// Missing set
	err = nm.SExpire("", "missing", 10)
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.STTL("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.SExpire("missing_ns", "exp_set", 10)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.STTL("missing_ns", "exp_set")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Expiration behavior
	require.NoError(t, nm.SAdd("", "ttl_set", "val", 0))
	require.NoError(t, nm.SExpire("", "ttl_set", 1))

	time.Sleep(2100 * time.Millisecond)

	_, err = nm.SCard("", "ttl_set")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.STTL("", "ttl_set")
	require.ErrorIs(t, err, ErrKeyNotFound)
}
