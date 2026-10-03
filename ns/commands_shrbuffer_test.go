package ns_test

import (
	"testing"
	"time"

	. "github.com/memap-project/memap-core/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BInit(t *testing.T) {
	nm := newTestManager()

	// Initialize buffer
	require.NoError(t, nm.BInit("", "buf1", 5, 0))

	capVal, err := nm.BCap("", "buf1")
	require.NoError(t, err)
	assert.Equal(t, int64(5), capVal)

	// BInit on existing buffer returns ErrKeyAlreadyExists
	err = nm.BInit("", "buf1", 10, 0)
	require.ErrorIs(t, err, ErrKeyAlreadyExists)

	// Non-existent namespace
	err = nm.BInit("missing_ns", "buf1", 5, 0)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Custom namespace
	require.NoError(t, nm.Create("custom_ns"))
	require.NoError(t, nm.BInit("custom_ns", "buf1", 8, 0))

	capVal, err = nm.BCap("custom_ns", "buf1")
	require.NoError(t, err)
	assert.Equal(t, int64(8), capVal)
}

func Test_BPush_and_BPop(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.BInit("", "rb", 3, 0))

	// Push items
	require.NoError(t, nm.BPush("", "rb", "item1"))
	require.NoError(t, nm.BPush("", "rb", "item2"))

	val, err := nm.BPop("", "rb")
	require.NoError(t, err)
	assert.Equal(t, "item1", val)

	val, err = nm.BPop("", "rb")
	require.NoError(t, err)
	assert.Equal(t, "item2", val)

	// Pop from empty buffer
	_, err = nm.BPop("", "rb")
	require.ErrorIs(t, err, ErrBufferEmpty)

	// Overwrite behavior when full (capacity 3)
	require.NoError(t, nm.BPush("", "rb", "v1"))
	require.NoError(t, nm.BPush("", "rb", "v2"))
	require.NoError(t, nm.BPush("", "rb", "v3"))
	require.NoError(t, nm.BPush("", "rb", "v4")) // overwrites v1

	val, err = nm.BPop("", "rb")
	require.NoError(t, err)
	assert.Equal(t, "v2", val)

	// Non-existent buffer
	err = nm.BPush("", "missing_buf", "v")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.BPop("", "missing_buf")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.BPush("missing_ns", "rb", "v")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.BPop("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_BAt_and_BSlice(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.BInit("", "rb", 5, 0))
	require.NoError(t, nm.BPush("", "rb", "first"))
	require.NoError(t, nm.BPush("", "rb", "second"))
	require.NoError(t, nm.BPush("", "rb", "third"))

	// BAt valid indices
	val, err := nm.BAt("", "rb", 0)
	require.NoError(t, err)
	assert.Equal(t, "first", val)

	val, err = nm.BAt("", "rb", 2)
	require.NoError(t, err)
	assert.Equal(t, "third", val)

	// BAt index out of bounds
	_, err = nm.BAt("", "rb", 3)
	require.ErrorIs(t, err, ErrIndexOutOfBounds)

	_, err = nm.BAt("", "rb", -1)
	require.ErrorIs(t, err, ErrIndexOutOfBounds)

	// BSlice
	slice, err := nm.BSlice("", "rb")
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second", "third"}, slice)

	// Non-existent buffer
	_, err = nm.BAt("", "missing", 0)
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.BSlice("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	_, err = nm.BAt("missing_ns", "rb", 0)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.BSlice("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_BPeek_and_BBack(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.BInit("", "rb", 3, 0))

	// Empty buffer
	_, err := nm.BPeek("", "rb")
	require.ErrorIs(t, err, ErrBufferEmpty)

	_, err = nm.BBack("", "rb")
	require.ErrorIs(t, err, ErrBufferEmpty)

	require.NoError(t, nm.BPush("", "rb", "head_item"))
	require.NoError(t, nm.BPush("", "rb", "tail_item"))

	peekVal, err := nm.BPeek("", "rb")
	require.NoError(t, err)
	assert.Equal(t, "head_item", peekVal)

	backVal, err := nm.BBack("", "rb")
	require.NoError(t, err)
	assert.Equal(t, "tail_item", backVal)

	// Peek and Back do not remove elements
	lenVal, err := nm.BLen("", "rb")
	require.NoError(t, err)
	assert.Equal(t, int64(2), lenVal)

	// Non-existent buffer
	_, err = nm.BPeek("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.BBack("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	_, err = nm.BPeek("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.BBack("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_BCap_BLen_BReset(t *testing.T) {
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

	// Non-existent buffer
	_, err = nm.BCap("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.BLen("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.BReset("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.BCap("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.BLen("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_BDel(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.BInit("", "rb", 5, 0))
	require.NoError(t, nm.BDel("", "rb"))

	_, err := nm.BCap("", "rb")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// BDel on non-existent buffer is a no-op
	require.NoError(t, nm.BDel("", "missing"))

	// Non-existent namespace
	err = nm.BDel("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)
}

func Test_BExpire_and_BTTL(t *testing.T) {
	nm := newTestManager()

	require.NoError(t, nm.BInit("", "rb", 5, 0))

	// Without expiration returns -1
	ttl, err := nm.BTTL("", "rb")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), ttl)

	// Set expiration
	require.NoError(t, nm.BExpire("", "rb", 60))
	ttl, err = nm.BTTL("", "rb")
	require.NoError(t, err)
	assert.Greater(t, ttl, int64(0))
	assert.LessOrEqual(t, ttl, int64(60))

	// Missing buffer
	err = nm.BExpire("", "missing", 10)
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.BTTL("", "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	// Non-existent namespace
	err = nm.BExpire("missing_ns", "rb", 10)
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	_, err = nm.BTTL("missing_ns", "rb")
	require.ErrorIs(t, err, ErrNamespaceNotFound)

	// Expiration behavior
	require.NoError(t, nm.BInit("", "exp_rb", 5, 1))
	time.Sleep(2100 * time.Millisecond)

	_, err = nm.BCap("", "exp_rb")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = nm.BTTL("", "exp_rb")
	require.ErrorIs(t, err, ErrKeyNotFound)
}
