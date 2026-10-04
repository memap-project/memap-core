package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CIncrBy(t *testing.T) {
	t.Run("auto-initialization and increment in default namespace", func(t *testing.T) {
		nm := newTestManager()

		val, err := nm.CIncrBy("", "cnt1", 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), val)

		val, err = nm.CIncrBy("", "cnt1", 3)
		require.NoError(t, err)
		assert.Equal(t, int64(4), val)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()

		_, err := nm.CIncrBy("custom_ns", "cnt1", 5)
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("increment in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))

		val, err := nm.CIncrBy("custom_ns", "cnt1", 5)
		require.NoError(t, err)
		assert.Equal(t, int64(5), val)
	})

	t.Run("fails when exceeding limit and preserves value", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.CSLimit("", "cnt_limit", 10))

		val, err := nm.CIncrBy("", "cnt_limit", 8)
		require.NoError(t, err)
		assert.Equal(t, int64(8), val)

		_, err = nm.CIncrBy("", "cnt_limit", 3)
		require.ErrorIs(t, err, ErrLimitExceeded)

		val, err = nm.CGet("", "cnt_limit")
		require.NoError(t, err)
		assert.Equal(t, int64(8), val)
	})
}

func Test_CDecrBy(t *testing.T) {
	t.Run("non-existent counter returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()

		_, err := nm.CDecrBy("", "missing_cnt", 1)
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()

		_, err := nm.CDecrBy("missing_ns", "cnt", 1)
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("valid decrement", func(t *testing.T) {
		nm := newTestManager()
		val, err := nm.CIncrBy("", "cnt_decr", 10)
		require.NoError(t, err)
		assert.Equal(t, int64(10), val)

		val, err = nm.CDecrBy("", "cnt_decr", 4)
		require.NoError(t, err)
		assert.Equal(t, int64(6), val)
	})

	t.Run("decrement below zero returns ErrLimitExceeded and preserves value", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.CIncrBy("", "cnt_decr", 6)
		require.NoError(t, err)

		_, err = nm.CDecrBy("", "cnt_decr", 7)
		require.ErrorIs(t, err, ErrLimitExceeded)

		val, err := nm.CGet("", "cnt_decr")
		require.NoError(t, err)
		assert.Equal(t, int64(6), val)
	})

	t.Run("decrement in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))

		_, err := nm.CIncrBy("custom_ns", "cnt", 15)
		require.NoError(t, err)

		val, err := nm.CDecrBy("custom_ns", "cnt", 5)
		require.NoError(t, err)
		assert.Equal(t, int64(10), val)
	})
}

func Test_CSLimit_and_CGLimit(t *testing.T) {
	t.Run("auto-initializes counter with limit", func(t *testing.T) {
		nm := newTestManager()

		require.NoError(t, nm.CSLimit("", "cnt", 100))

		lim, err := nm.CGLimit("", "cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(100), lim)

		val, err := nm.CGet("", "cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(0), val)
	})

	t.Run("missing counter returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()

		_, err := nm.CGLimit("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("missing namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()

		err := nm.CSLimit("missing_ns", "cnt", 50)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.CGLimit("missing_ns", "cnt")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("set and get limit in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("ns1"))

		require.NoError(t, nm.CSLimit("ns1", "cnt", 50))
		lim, err := nm.CGLimit("ns1", "cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(50), lim)
	})
}

func Test_CGet(t *testing.T) {
	t.Run("missing counter returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()

		_, err := nm.CGet("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("missing namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()

		_, err := nm.CGet("missing_ns", "cnt")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("get value in default namespace", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.CIncrBy("", "cnt", 7)
		require.NoError(t, err)

		val, err := nm.CGet("", "cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(7), val)
	})

	t.Run("get value in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("ns1"))
		_, err := nm.CIncrBy("ns1", "cnt", 42)
		require.NoError(t, err)

		val, err := nm.CGet("ns1", "cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(42), val)
	})
}

func Test_CDel(t *testing.T) {
	t.Run("delete counter in default namespace", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.CIncrBy("", "cnt", 10)
		require.NoError(t, err)

		require.NoError(t, nm.CDel("", "cnt"))

		_, err = nm.CGet("", "cnt")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("delete in missing namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()

		err := nm.CDel("missing_ns", "cnt")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("delete counter in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("ns1"))
		_, err := nm.CIncrBy("ns1", "cnt", 10)
		require.NoError(t, err)

		require.NoError(t, nm.CDel("ns1", "cnt"))
		_, err = nm.CGet("ns1", "cnt")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})
}

func Test_CExpire_and_CTTL(t *testing.T) {
	t.Run("missing counter returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()

		err := nm.CExpire("", "missing", 10)
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.CTTL("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("missing namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()

		err := nm.CExpire("missing_ns", "cnt", 10)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.CTTL("missing_ns", "cnt")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("counter without TTL returns -1", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.CIncrBy("", "cnt", 5)
		require.NoError(t, err)

		ttl, err := nm.CTTL("", "cnt")
		require.NoError(t, err)
		assert.Equal(t, int64(-1), ttl)
	})

	t.Run("sets valid TTL", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.CIncrBy("", "cnt", 5)
		require.NoError(t, err)

		require.NoError(t, nm.CExpire("", "cnt", 60))
		ttl, err := nm.CTTL("", "cnt")
		require.NoError(t, err)
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(60))
	})

	t.Run("sets TTL in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("ns1"))
		_, err := nm.CIncrBy("ns1", "cnt", 5)
		require.NoError(t, err)

		require.NoError(t, nm.CExpire("ns1", "cnt", 100))
		ttl, err := nm.CTTL("ns1", "cnt")
		require.NoError(t, err)
		assert.Greater(t, ttl, int64(0))
	})

	t.Run("expired counter returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.CIncrBy("", "exp_cnt", 1)
		require.NoError(t, err)
		require.NoError(t, nm.CExpire("", "exp_cnt", 1))

		time.Sleep(2100 * time.Millisecond)
		_, err = nm.CGet("", "exp_cnt")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.CTTL("", "exp_cnt")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})
}
