package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CIncrBy(t *testing.T) {
	nm := newTestManager()

	// Auto-initialization in default namespace
	val, err := nm.CIncrBy("", "cnt1", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Further increment on active counter
	val, err = nm.CIncrBy("", "cnt1", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(4), val)

	// Non-existent namespace returns ErrNamespaceNotFound
	_, err = nm.CIncrBy("custom_ns", "cnt1", 5)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	val, err = nm.CIncrBy("custom_ns", "cnt1", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)

	// Limit enforcement
	require.NoError(t, nm.CSLimit("", "cnt_limit", 10))
	val, err = nm.CIncrBy("", "cnt_limit", 8)
	require.NoError(t, err)
	assert.Equal(t, int64(8), val)

	// Exceeding limit returns ErrLimitExceeded
	_, err = nm.CIncrBy("", "cnt_limit", 3)
	require.ErrorIs(t, err, ErrLimitExceeded)

	// Counter value remains unchanged after exceeding limit
	val, err = nm.CGet("", "cnt_limit")
	require.NoError(t, err)
	assert.Equal(t, int64(8), val)
}

func Test_CDecrBy(t *testing.T) {
	nm := newTestManager()

	// DecrBy on non-existent counter returns ErrKeyNotFound
	_, err := nm.CDecrBy("", "missing_cnt", 1)
	require.ErrorIs(t, err, ErrKeyNotFound)

	// DecrBy on non-existent namespace
	_, err = nm.CDecrBy("missing_ns", "cnt", 1)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Increment first, then decrement
	val, err := nm.CIncrBy("", "cnt_decr", 10)
	require.NoError(t, err)
	assert.Equal(t, int64(10), val)

	val, err = nm.CDecrBy("", "cnt_decr", 4)
	require.NoError(t, err)
	assert.Equal(t, int64(6), val)

	// Decrement below 0 returns ErrLimitExceeded
	_, err = nm.CDecrBy("", "cnt_decr", 7)
	require.ErrorIs(t, err, ErrLimitExceeded)

	// Value remains 6
	val, err = nm.CGet("", "cnt_decr")
	require.NoError(t, err)
	assert.Equal(t, int64(6), val)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	_, err = nm.CIncrBy("custom_ns", "cnt", 15)
	require.NoError(t, err)

	val, err = nm.CDecrBy("custom_ns", "cnt", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(10), val)
}

func Test_CSLimit_and_CGLimit(t *testing.T) {
	nm := newTestManager()

	// CSLimit auto-initializes counter with given limit
	require.NoError(t, nm.CSLimit("", "cnt", 100))

	lim, err := nm.CGLimit("", "cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(100), lim)

	val, err := nm.CGet("", "cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(0), val)

	// CGLimit on missing counter
	_, err = nm.CGLimit("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Missing namespace
	err = nm.CSLimit("missing_ns", "cnt", 50)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.CGLimit("missing_ns", "cnt")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("ns1"))
	require.NoError(t, nm.CSLimit("ns1", "cnt", 50))
	lim, err = nm.CGLimit("ns1", "cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(50), lim)
}

func Test_CGet(t *testing.T) {
	nm := newTestManager()

	// Missing counter
	_, err := nm.CGet("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Missing namespace
	_, err = nm.CGet("missing_ns", "cnt")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Default namespace
	_, err = nm.CIncrBy("", "cnt", 7)
	require.NoError(t, err)
	val, err := nm.CGet("", "cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(7), val)

	// Custom namespace
	require.NoError(t, nm.Create("ns1"))
	_, err = nm.CIncrBy("ns1", "cnt", 42)
	require.NoError(t, err)
	val, err = nm.CGet("ns1", "cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(42), val)
}

func Test_CDel(t *testing.T) {
	nm := newTestManager()

	// Delete counter in default namespace
	_, err := nm.CIncrBy("", "cnt", 10)
	require.NoError(t, err)

	require.NoError(t, nm.CDel("", "cnt"))

	_, err = nm.CGet("", "cnt")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// CDel on missing namespace
	err = nm.CDel("missing_ns", "cnt")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("ns1"))
	_, err = nm.CIncrBy("ns1", "cnt", 10)
	require.NoError(t, err)

	require.NoError(t, nm.CDel("ns1", "cnt"))
	_, err = nm.CGet("ns1", "cnt")
	require.ErrorIs(t, err, ErrKeyNotFound)
}

func Test_CExpire_and_CTTL(t *testing.T) {
	nm := newTestManager()

	// Missing counter
	err := nm.CExpire("", "missing", 10)
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.CTTL("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Missing namespace
	err = nm.CExpire("missing_ns", "cnt", 10)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.CTTL("missing_ns", "cnt")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Existing counter without TTL returns -1
	_, err = nm.CIncrBy("", "cnt", 5)
	require.NoError(t, err)

	ttl, err := nm.CTTL("", "cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), ttl)

	// Set TTL
	require.NoError(t, nm.CExpire("", "cnt", 60))
	ttl, err = nm.CTTL("", "cnt")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))
	assert.LessOrEqual(t, ttl, int64(60))

	// Custom namespace
	require.NoError(t, nm.Create("ns1"))
	_, err = nm.CIncrBy("ns1", "cnt", 5)
	require.NoError(t, err)

	require.NoError(t, nm.CExpire("ns1", "cnt", 100))
	ttl, err = nm.CTTL("ns1", "cnt")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))

	// Expiration behavior
	_, err = nm.CIncrBy("", "exp_cnt", 1)
	require.NoError(t, err)
	require.NoError(t, nm.CExpire("", "exp_cnt", 1))

	time.Sleep(2100 * time.Millisecond)
	_, err = nm.CGet("", "exp_cnt")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.CTTL("", "exp_cnt")
	require.ErrorIs(t, err, ErrKeyNotFound)
}
