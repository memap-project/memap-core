package vals

import (
	"math"
	"sync/atomic"
	"time"
)

// processStart anchors monotonic timestamps for all expiration calculations,
// making TTL tracking immune to wall-clock (NTP) adjustments.
var processStart = time.Now()

// MonotonicNow returns monotonic nanoseconds elapsed since process start (always >= 1).
func MonotonicNow() int64 {
	ns := time.Since(processStart).Nanoseconds()
	if ns <= 0 {
		return 1
	}
	return ns
}

// IsValidTTL reports whether ttl (in seconds) is strictly positive and will not
// overflow int64 nanoseconds when added to the current monotonic timestamp.
func IsValidTTL(ttl int64) bool {
	if ttl <= 0 {
		return false
	}
	now := MonotonicNow()
	return ttl <= (math.MaxInt64-now)/int64(time.Second)
}

// expiry provides lock-free monotonic expiration tracking shared by all value types.
// A zero expiresAt value indicates that the entry never expires.
type expiry struct {
	expiresAt atomic.Int64 // monotonic nanoseconds since processStart; 0 = no expiration
}

// IsExpired reports whether the entry has an expiration deadline that has passed.
func (e *expiry) IsExpired() bool {
	return e.IsExpiredAt(MonotonicNow())
}

// IsExpiredAt reports whether the entry is expired at the given monotonic timestamp (ns).
func (e *expiry) IsExpiredAt(nowNs int64) bool {
	exp := e.expiresAt.Load()
	if exp == 0 {
		return false
	}
	return nowNs >= exp
}

// Expire atomically sets a new TTL (in seconds) relative to now.
// Returns false if ttl is non-positive, overflows int64, or the entry is already expired.
func (e *expiry) Expire(ttl int64) bool {
	return e.ExpireAt(MonotonicNow(), ttl)
}

// ExpireAt atomically sets a new TTL (in seconds) relative to nowNs.
// Returns false if ttl <= 0, overflows int64, or the entry is already expired at nowNs.
func (e *expiry) ExpireAt(nowNs, ttl int64) bool {
	if ttl <= 0 || ttl > (math.MaxInt64-nowNs)/int64(time.Second) {
		return false
	}
	newExp := nowNs + ttl*int64(time.Second)
	for {
		cur := e.expiresAt.Load()
		if cur != 0 && nowNs >= cur {
			return false
		}
		if e.expiresAt.CompareAndSwap(cur, newExp) {
			return true
		}
	}
}

// TTL returns the remaining time-to-live in whole seconds (rounded up so any
// unexpired key with a TTL returns >= 1).
// Returns -1 if the entry has no expiration.
// Returns -2 if the entry is expired.
func (e *expiry) TTL() int64 {
	return e.TTLAt(MonotonicNow())
}

// TTLAt returns the remaining time-to-live in whole seconds at the given monotonic timestamp.
// Returns -1 if the entry has no expiration, -2 if expired, or >= 1 if active with TTL.
func (e *expiry) TTLAt(nowNs int64) int64 {
	exp := e.expiresAt.Load()
	if exp == 0 {
		return -1
	}
	remNs := exp - nowNs
	if remNs <= 0 {
		return -2
	}
	sec := remNs / int64(time.Second)
	if remNs%int64(time.Second) > 0 {
		sec++
	}
	return sec
}
