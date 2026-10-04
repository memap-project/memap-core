package tests

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
	t.Run("create new namespace", func(t *testing.T) {
		nm := newTestManager()
		err := nm.Create("test_ns")
		require.NoError(t, err)

		ns, exists := nm.GetNs("test_ns")
		require.True(t, exists)
		assert.NotNil(t, ns)
	})

	t.Run("creating duplicate namespace returns ErrNamespaceAlreadyExists", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("test_ns"))

		err := nm.Create("test_ns")
		require.ErrorIs(t, err, ErrNamespaceAlreadyExists)
	})

	t.Run("drop namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("test_ns"))

		err := nm.Drop("test_ns")
		require.NoError(t, err)

		ns, exists := nm.GetNs("test_ns")
		require.False(t, exists)
		assert.Nil(t, ns)
	})

	t.Run("get non-existent namespace returns false", func(t *testing.T) {
		nm := newTestManager()
		ns, exists := nm.GetNs("non_existent")
		require.False(t, exists)
		assert.Nil(t, ns)
	})
}

func Test_Flush(t *testing.T) {
	t.Run("flushes data across all namespaces while preserving namespaces", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))

		// Populate data
		require.NoError(t, nm.Set("", "k1", "v1", 0))
		_, err := nm.CIncrBy("", "cnt1", 5)
		require.NoError(t, err)

		require.NoError(t, nm.Set("custom_ns", "k2", "v2", 0))
		_, err = nm.CIncrBy("custom_ns", "cnt2", 10)
		require.NoError(t, err)

		// Flush
		nm.Flush()

		// Verify keys cleared in default namespace
		_, err = nm.Get("", "k1")
		require.ErrorIs(t, err, ErrKeyNotFound)
		_, err = nm.CGet("", "cnt1")
		require.ErrorIs(t, err, ErrKeyNotFound)

		// Verify keys cleared in custom namespace
		_, err = nm.Get("custom_ns", "k2")
		require.ErrorIs(t, err, ErrKeyNotFound)
		_, err = nm.CGet("custom_ns", "cnt2")
		require.ErrorIs(t, err, ErrKeyNotFound)

		// Custom namespace itself still exists
		_, exists := nm.GetNs("custom_ns")
		assert.True(t, exists)
	})
}

func Test_Erase(t *testing.T) {
	t.Run("clears default namespace and drops custom namespaces", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom1"))
		require.NoError(t, nm.Create("custom2"))

		require.NoError(t, nm.Set("", "k_default", "v_default", 0))
		require.NoError(t, nm.Set("custom1", "k_c1", "v_c1", 0))

		nm.Erase()

		// Default namespace cleared
		_, err := nm.Get("", "k_default")
		require.ErrorIs(t, err, ErrKeyNotFound)

		// Custom namespaces dropped
		_, exists := nm.GetNs("custom1")
		assert.False(t, exists)
		_, exists = nm.GetNs("custom2")
		assert.False(t, exists)

		// Operations on dropped namespace fail
		_, err = nm.Get("custom1", "k_c1")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_CleanExpired(t *testing.T) {
	t.Run("removes expired keys across all namespaces", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))

		require.NoError(t, nm.Set("", "exp_def", "v", 1))
		require.NoError(t, nm.Set("custom_ns", "exp_cust", "v", 1))

		require.NoError(t, nm.Set("", "keep_def", "v", 0))
		require.NoError(t, nm.Set("custom_ns", "keep_cust", "v", 0))

		time.Sleep(2100 * time.Millisecond)

		nm.CleanExpired()

		// Expired keys are removed
		_, err := nm.Get("", "exp_def")
		require.ErrorIs(t, err, ErrKeyNotFound)
		_, err = nm.Get("custom_ns", "exp_cust")
		require.ErrorIs(t, err, ErrKeyNotFound)

		// Non-expired keys remain
		val, err := nm.Get("", "keep_def")
		require.NoError(t, err)
		assert.Equal(t, "v", val)

		val, err = nm.Get("custom_ns", "keep_cust")
		require.NoError(t, err)
		assert.Equal(t, "v", val)
	})
}
