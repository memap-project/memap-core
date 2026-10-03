package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HSet_and_HGet(t *testing.T) {
	nm := newTestManager()

	// HSet initializes empty hash
	require.NoError(t, nm.HSet("", "hash1", 0))

	val, err := nm.HGet("", "hash1")
	require.NoError(t, err)
	assert.Empty(t, val)

	// Non-existent hash
	_, err = nm.HGet("", "missing_hash")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.HSet("missing_ns", "hash1", 0)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.HGet("missing_ns", "hash1")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	require.NoError(t, nm.HSet("custom_ns", "hash1", 0))

	val, err = nm.HGet("custom_ns", "hash1")
	require.NoError(t, err)
	assert.Empty(t, val)
}

func Test_HFSet_and_HFGet(t *testing.T) {
	nm := newTestManager()

	// HFSet auto-initializes hash
	require.NoError(t, nm.HFSet("", "user:1", "name", "Alice"))
	require.NoError(t, nm.HFSet("", "user:1", "role", "Admin"))

	val, err := nm.HFGet("", "user:1", "name")
	require.NoError(t, err)
	assert.Equal(t, "Alice", val)

	val, err = nm.HFGet("", "user:1", "role")
	require.NoError(t, err)
	assert.Equal(t, "Admin", val)

	// Update existing field
	require.NoError(t, nm.HFSet("", "user:1", "name", "Bob"))
	val, err = nm.HFGet("", "user:1", "name")
	require.NoError(t, err)
	assert.Equal(t, "Bob", val)

	// Field not found
	_, err = nm.HFGet("", "user:1", "missing_field")
	require.ErrorIs(t, err, ErrFieldNotFound)

	// Hash not found
	_, err = nm.HFGet("", "missing_hash", "field")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.HFSet("missing_ns", "h", "f", "v")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.HFGet("missing_ns", "h", "f")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	require.NoError(t, nm.HFSet("custom_ns", "user:2", "email", "bob@example.com"))

	val, err = nm.HFGet("custom_ns", "user:2", "email")
	require.NoError(t, err)
	assert.Equal(t, "bob@example.com", val)
}

func Test_HFDel(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.HFSet("", "h", "f1", "v1"))
	require.NoError(t, nm.HFSet("", "h", "f2", "v2"))

	require.NoError(t, nm.HFDel("", "h", "f1"))

	_, err := nm.HFGet("", "h", "f1")
	require.ErrorIs(t, err, ErrFieldNotFound)

	val, err := nm.HFGet("", "h", "f2")
	require.NoError(t, err)
	assert.Equal(t, "v2", val)

	// HFDel on non-existent field or hash does not error
	require.NoError(t, nm.HFDel("", "h", "missing_field"))
	require.NoError(t, nm.HFDel("", "missing_hash", "field"))

	// Non-existent namespace
	err = nm.HFDel("missing_ns", "h", "f")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_HDel(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.HFSet("", "h", "f1", "v1"))
	require.NoError(t, nm.HDel("", "h"))

	_, err := nm.HGet("", "h")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// HDel on non-existent hash
	require.NoError(t, nm.HDel("", "missing"))

	// Non-existent namespace
	err = nm.HDel("missing_ns", "h")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_HExists(t *testing.T) {
	nm := newTestManager()

	exists, err := nm.HExists("", "h")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, nm.HFSet("", "h", "f", "v"))
	exists, err = nm.HExists("", "h")
	require.NoError(t, err)
	assert.True(t, exists)

	// Non-existent namespace
	_, err = nm.HExists("missing_ns", "h")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_HLen_HKeys_HValues(t *testing.T) {
	nm := newTestManager()

	// Missing hash
	_, err := nm.HLen("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.HKeys("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.HValues("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Populate hash
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

	// Non-existent namespace
	_, err = nm.HLen("missing_ns", "profile")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.HKeys("missing_ns", "profile")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.HValues("missing_ns", "profile")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_HExpire_and_HTTL(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.HFSet("", "h", "f", "v"))

	// Hash without expiration returns -1
	ttl, err := nm.HTTL("", "h")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), ttl)

	// Expire sets TTL
	require.NoError(t, nm.HExpire("", "h", 60))
	ttl, err = nm.HTTL("", "h")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))
	assert.LessOrEqual(t, ttl, int64(60))

	// Missing hash
	err = nm.HExpire("", "missing", 10)
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.HTTL("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.HExpire("missing_ns", "h", 10)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.HTTL("missing_ns", "h")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Expiration behavior
	require.NoError(t, nm.HSet("", "exp_h", 1))
	require.NoError(t, nm.HFSet("", "exp_h", "f", "v"))
	require.NoError(t, nm.HExpire("", "exp_h", 1))

	time.Sleep(2100 * time.Millisecond)

	_, err = nm.HGet("", "exp_h")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.HTTL("", "exp_h")
	require.ErrorIs(t, err, ErrKeyNotFound)
}
