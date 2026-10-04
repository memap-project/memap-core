package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BInit(t *testing.T) {
	t.Run("initialize buffer", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "buf1", 5, 0))

		capVal, err := nm.BCap("", "buf1")
		require.NoError(t, err)
		assert.Equal(t, int64(5), capVal)
	})

	t.Run("duplicate buffer initialization returns ErrKeyAlreadyExists", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "buf1", 5, 0))

		err := nm.BInit("", "buf1", 10, 0)
		require.ErrorIs(t, err, ErrKeyAlreadyExists)
	})

	t.Run("non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BInit("missing_ns", "buf1", 5, 0)
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("initialize buffer in custom namespace", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.Create("custom_ns"))
		require.NoError(t, nm.BInit("custom_ns", "buf1", 8, 0))

		capVal, err := nm.BCap("custom_ns", "buf1")
		require.NoError(t, err)
		assert.Equal(t, int64(8), capVal)
	})
}

func Test_BPush_and_BPop(t *testing.T) {
	t.Run("push and pop items in FIFO order", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 3, 0))

		require.NoError(t, nm.BPush("", "rb", "item1"))
		require.NoError(t, nm.BPush("", "rb", "item2"))

		val, err := nm.BPop("", "rb")
		require.NoError(t, err)
		assert.Equal(t, "item1", val)

		val, err = nm.BPop("", "rb")
		require.NoError(t, err)
		assert.Equal(t, "item2", val)
	})

	t.Run("pop from empty buffer returns ErrBufferEmpty", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 3, 0))

		_, err := nm.BPop("", "rb")
		require.ErrorIs(t, err, ErrBufferEmpty)
	})

	t.Run("overflow overwrites oldest item when buffer is full", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 3, 0))

		require.NoError(t, nm.BPush("", "rb", "v1"))
		require.NoError(t, nm.BPush("", "rb", "v2"))
		require.NoError(t, nm.BPush("", "rb", "v3"))
		require.NoError(t, nm.BPush("", "rb", "v4")) // overwrites v1

		val, err := nm.BPop("", "rb")
		require.NoError(t, err)
		assert.Equal(t, "v2", val)
	})

	t.Run("push and pop on non-existent buffer returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BPush("", "missing_buf", "v")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.BPop("", "missing_buf")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("push and pop in non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BPush("missing_ns", "rb", "v")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.BPop("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_BAt_and_BSlice(t *testing.T) {
	t.Run("access elements by index", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 5, 0))
		require.NoError(t, nm.BPush("", "rb", "first"))
		require.NoError(t, nm.BPush("", "rb", "second"))
		require.NoError(t, nm.BPush("", "rb", "third"))

		val, err := nm.BAt("", "rb", 0)
		require.NoError(t, err)
		assert.Equal(t, "first", val)

		val, err = nm.BAt("", "rb", 2)
		require.NoError(t, err)
		assert.Equal(t, "third", val)
	})

	t.Run("index out of bounds returns ErrIndexOutOfBounds", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 5, 0))
		require.NoError(t, nm.BPush("", "rb", "first"))

		_, err := nm.BAt("", "rb", 3)
		require.ErrorIs(t, err, ErrIndexOutOfBounds)

		_, err = nm.BAt("", "rb", -1)
		require.ErrorIs(t, err, ErrIndexOutOfBounds)
	})

	t.Run("slice returns all elements in logical order", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 5, 0))
		require.NoError(t, nm.BPush("", "rb", "first"))
		require.NoError(t, nm.BPush("", "rb", "second"))
		require.NoError(t, nm.BPush("", "rb", "third"))

		slice, err := nm.BSlice("", "rb")
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "second", "third"}, slice)
	})

	t.Run("operations on non-existent buffer return ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.BAt("", "missing", 0)
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.BSlice("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("operations in non-existent namespace return ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.BAt("missing_ns", "rb", 0)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.BSlice("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_BPeek_and_BBack(t *testing.T) {
	t.Run("peek and back on empty buffer return ErrBufferEmpty", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 3, 0))

		_, err := nm.BPeek("", "rb")
		require.ErrorIs(t, err, ErrBufferEmpty)

		_, err = nm.BBack("", "rb")
		require.ErrorIs(t, err, ErrBufferEmpty)
	})

	t.Run("peek and back return head and tail without removing", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 3, 0))

		require.NoError(t, nm.BPush("", "rb", "head_item"))
		require.NoError(t, nm.BPush("", "rb", "tail_item"))

		peekVal, err := nm.BPeek("", "rb")
		require.NoError(t, err)
		assert.Equal(t, "head_item", peekVal)

		backVal, err := nm.BBack("", "rb")
		require.NoError(t, err)
		assert.Equal(t, "tail_item", backVal)

		// Length is preserved
		lenVal, err := nm.BLen("", "rb")
		require.NoError(t, err)
		assert.Equal(t, int64(2), lenVal)
	})

	t.Run("peek and back on non-existent buffer return ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.BPeek("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.BBack("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("peek and back in non-existent namespace return ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.BPeek("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.BBack("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_BCap_BLen_BReset(t *testing.T) {
	t.Run("capacity, length, and reset", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 10, 0))

		capVal, err := nm.BCap("", "rb")
		require.NoError(t, err)
		assert.Equal(t, int64(10), capVal)

		lenVal, err := nm.BLen("", "rb")
		require.NoError(t, err)
		assert.Equal(t, int64(0), lenVal)

		require.NoError(t, nm.BPush("", "rb", "item"))
		lenVal, err = nm.BLen("", "rb")
		require.NoError(t, err)
		assert.Equal(t, int64(1), lenVal)

		require.NoError(t, nm.BReset("", "rb"))
		lenVal, err = nm.BLen("", "rb")
		require.NoError(t, err)
		assert.Equal(t, int64(0), lenVal)

		_, err = nm.BPop("", "rb")
		require.ErrorIs(t, err, ErrBufferEmpty)
	})

	t.Run("operations on non-existent buffer return ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		_, err := nm.BCap("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.BLen("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("operations in non-existent namespace return ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BReset("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.BCap("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.BLen("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_BDel(t *testing.T) {
	t.Run("delete existing buffer", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 5, 0))
		require.NoError(t, nm.BDel("", "rb"))

		_, err := nm.BCap("", "rb")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("delete non-existent buffer is a no-op", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BDel("", "missing"))
	})

	t.Run("delete in non-existent namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BDel("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})
}

func Test_BExpire_and_BTTL(t *testing.T) {
	t.Run("buffer without TTL returns -1", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 5, 0))

		ttl, err := nm.BTTL("", "rb")
		require.NoError(t, err)
		assert.Equal(t, int64(-1), ttl)
	})

	t.Run("set valid expiration TTL", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "rb", 5, 0))

		require.NoError(t, nm.BExpire("", "rb", 60))
		ttl, err := nm.BTTL("", "rb")
		require.NoError(t, err)
		assert.Greater(t, ttl, int64(0))
		assert.LessOrEqual(t, ttl, int64(60))
	})

	t.Run("missing buffer returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BExpire("", "missing", 10)
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.BTTL("", "missing")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("missing namespace returns ErrNamespaceNotFound", func(t *testing.T) {
		nm := newTestManager()
		err := nm.BExpire("missing_ns", "rb", 10)
		require.ErrorIs(t, err, ErrNamespaceNotFound)

		_, err = nm.BTTL("missing_ns", "rb")
		require.ErrorIs(t, err, ErrNamespaceNotFound)
	})

	t.Run("expired buffer returns ErrKeyNotFound", func(t *testing.T) {
		nm := newTestManager()
		require.NoError(t, nm.BInit("", "exp_rb", 5, 1))

		time.Sleep(2100 * time.Millisecond)

		_, err := nm.BCap("", "exp_rb")
		require.ErrorIs(t, err, ErrKeyNotFound)

		_, err = nm.BTTL("", "exp_rb")
		require.ErrorIs(t, err, ErrKeyNotFound)
	})
}
