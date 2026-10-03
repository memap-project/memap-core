package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Set_and_Get(t *testing.T) {
	nm := newTestManager()

	// Default namespace
	require.NoError(t, nm.Set("", "key1", "val1", 0))
	val, err := nm.Get("", "key1")
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// Overwrite value
	require.NoError(t, nm.Set("", "key1", "val2", 0))
	val, err = nm.Get("", "key1")
	require.NoError(t, err)
	assert.Equal(t, "val2", val)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	require.NoError(t, nm.Set("custom_ns", "key1", "custom_val", 0))

	val, err = nm.Get("custom_ns", "key1")
	require.NoError(t, err)
	assert.Equal(t, "custom_val", val)

	// Verify isolation between namespaces
	val, err = nm.Get("", "key1")
	require.NoError(t, err)
	assert.Equal(t, "val2", val)

	// Non-existent namespace
	err = nm.Set("non_existent", "key", "val", 0)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.Get("non_existent", "key")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Non-existent key
	_, err = nm.Get("", "missing_key")
	require.ErrorIs(t, err, ErrKeyNotFound)
}

func Test_Del(t *testing.T) {
	nm := newTestManager()

	// Del on existing key in default ns
	require.NoError(t, nm.Set("", "key1", "val1", 0))
	require.NoError(t, nm.Del("", "key1"))

	_, err := nm.Get("", "key1")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Del on non-existent key (succeeds as no-op)
	require.NoError(t, nm.Del("", "missing_key"))

	// Del on non-existent namespace
	err = nm.Del("missing_ns", "key1")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	require.NoError(t, nm.Set("custom_ns", "key1", "val1", 0))
	require.NoError(t, nm.Del("custom_ns", "key1"))

	_, err = nm.Get("custom_ns", "key1")
	require.ErrorIs(t, err, ErrKeyNotFound)
}

func Test_Expire_and_TTL(t *testing.T) {
	nm := newTestManager()

	// Key without TTL returns -1
	require.NoError(t, nm.Set("", "key1", "val1", 0))
	ttl, err := nm.TTL("", "key1")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), ttl)

	// Expire sets TTL
	require.NoError(t, nm.Expire("", "key1", 60))
	ttl, err = nm.TTL("", "key1")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))
	assert.LessOrEqual(t, ttl, int64(60))

	// Set with initial TTL
	require.NoError(t, nm.Set("", "key_ttl", "val", 100))
	ttl, err = nm.TTL("", "key_ttl")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))
	assert.LessOrEqual(t, ttl, int64(100))

	// Expire non-existent key
	err = nm.Expire("", "missing_key", 10)
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.TTL("", "missing_key")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.Expire("missing_ns", "key", 10)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.TTL("missing_ns", "key")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Expiration behavior
	require.NoError(t, nm.Set("", "exp_key", "val", 1))
	time.Sleep(2100 * time.Millisecond)

	_, err = nm.Get("", "exp_key")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.TTL("", "exp_key")
	require.ErrorIs(t, err, ErrKeyNotFound)
}
