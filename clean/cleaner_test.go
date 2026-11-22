package clean

import (
	"context"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleaner_InvalidInputsDoNotPanic(t *testing.T) {
	cases := []struct {
		name     string
		interval int
		fn       func()
	}{
		{"zero interval", 0, func() {}},
		{"negative interval", -1, func() {}},
		{"min int interval", math.MinInt, func() {}},
		{"overflow interval", math.MaxInt, func() {}},
		{"nil cleanFunc", 1, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var nilCtx context.Context
			c := NewCleaner(nilCtx, tc.interval, tc.fn)
			assert.NotPanics(t, func() {
				c.Start()
				c.Start() // idempotent
				c.Stop()
				c.Stop() // idempotent
			})
		})
	}

	t.Run("nil receiver is safe", func(t *testing.T) {
		var c *Cleaner
		assert.NotPanics(t, func() {
			c.Start()
			c.Stop()
		})
	})
}

func TestCleaner_StartIdempotentAndStopWaits(t *testing.T) {
	var runs atomic.Int32
	c := NewCleaner(context.Background(), 1, func() {
		runs.Add(1)
	})

	c.Start()
	c.Start() // must not spawn a second goroutine

	assert.Eventually(t, func() bool {
		return runs.Load() >= 1
	}, 2500*time.Millisecond, 50*time.Millisecond)

	c.Stop()
	afterStop := runs.Load()
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, afterStop, runs.Load())
}
