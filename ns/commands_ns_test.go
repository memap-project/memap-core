package ns_test

import (
	"testing"
	"time"

	"github.com/memap-project/memap-core/config"
	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager() *NamespaceManager {
	cfg := &config.NamespaceConfig{
		ShardCounts: config.ShardCounts{
			Shmap:     4,
			Shhash:    4,
			Shcounter: 4,
			Shrbuffer: 4,
			Shset:     4,
		},
	}
	return NewNamespaceManager(cfg)
}

func Test_Create_and_Drop(t *testing.T) {
	nm := newTestManager()

	// Create new namespace
	err := nm.Create("test_ns")
	require.NoError(t, err)

	// Creating existing namespace returns ErrNamespaceAlreadyExists
	err = nm.Create("test_ns")
	require.ErrorIs(t, err, ErrNamespaceAlreadyExists)

	// Namespace exists
	ns, exists := nm.GetNs("test_ns")
	require.True(t, exists)
	assert.NotNil(t, ns)

	// Drop namespace
	err = nm.Drop("test_ns")
	require.NoError(t, err)

	// Namespace no longer exists
	ns, exists = nm.GetNs("test_ns")
	require.False(t, exists)
	assert.Nil(t, ns)
}

func Test_Flush(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.Create("custom_ns"))

	// Set data in default namespace
	require.NoError(t, nm.Set("", "k1", "v1", 0))
	_, err := nm.CIncrBy("", "cnt1", 5)
	require.NoError(t, err)

	// Set data in custom namespace
	require.NoError(t, nm.Set("custom_ns", "k2", "v2", 0))
	_, err = nm.CIncrBy("custom_ns", "cnt2", 10)
	require.NoError(t, err)

	// Flush removes all keys across all namespaces
	nm.Flush()

	// Keys in default namespace are gone
	_, err = nm.Get("", "k1")
	require.ErrorIs(t, err, ErrKeyNotFound)
	_, err = nm.CGet("", "cnt1")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Keys in custom namespace are gone
	_, err = nm.Get("custom_ns", "k2")
	require.ErrorIs(t, err, ErrKeyNotFound)
	_, err = nm.CGet("custom_ns", "cnt2")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Custom namespace still exists after Flush
	_, exists := nm.GetNs("custom_ns")
	assert.True(t, exists)
}

func Test_Erase(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.Create("custom1"))
	require.NoError(t, nm.Create("custom2"))

	// Set data in default and custom namespaces
	require.NoError(t, nm.Set("", "k_default", "v_default", 0))
	require.NoError(t, nm.Set("custom1", "k_c1", "v_c1", 0))

	// Erase flushes default namespace and drops all custom namespaces
	nm.Erase()

	// Default namespace is cleared
	_, err := nm.Get("", "k_default")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Custom namespaces are completely dropped
	_, exists := nm.GetNs("custom1")
	assert.False(t, exists)
	_, exists = nm.GetNs("custom2")
	assert.False(t, exists)

	// Operating on custom namespace now returns ErrNamespaceNotFound
	_, err = nm.Get("custom1", "k_c1")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_CleanExpired(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.Create("custom_ns"))

	require.NoError(t, nm.Set("", "exp_def", "v", 1))
	require.NoError(t, nm.Set("custom_ns", "exp_cust", "v", 1))

	require.NoError(t, nm.Set("", "keep_def", "v", 0))
	require.NoError(t, nm.Set("custom_ns", "keep_cust", "v", 0))

	time.Sleep(2100 * time.Millisecond)

	nm.CleanExpired()

	_, err := nm.Get("", "exp_def")
	require.ErrorIs(t, err, ErrKeyNotFound)
	_, err = nm.Get("custom_ns", "exp_cust")
	require.ErrorIs(t, err, ErrKeyNotFound)

	val, err := nm.Get("", "keep_def")
	require.NoError(t, err)
	assert.Equal(t, "v", val)

	val, err = nm.Get("custom_ns", "keep_cust")
	require.NoError(t, err)
	assert.Equal(t, "v", val)
}
