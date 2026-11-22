package clean

import (
	"context"
	"math"
	"sync"
	"time"
)

const maxIntervalSec int64 = math.MaxInt64 / int64(time.Second)

// Cleaner executes a periodic cleanup function on a fixed interval until stopped or canceled.
type Cleaner struct {
	interval  time.Duration
	cleanFunc func()
	ctx       context.Context
	cancel    context.CancelFunc
	startOnce sync.Once
	wg        sync.WaitGroup
}

// NewCleaner creates a new Cleaner with the given interval (in seconds) and cleanup function.
// If parentCtx is nil, context.Background() is used.
// If interval is non-positive or overflows time.Duration, or cleanFunc is nil, Start becomes a safe no-op.
func NewCleaner(parentCtx context.Context, interval int, cleanFunc func()) *Cleaner {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	var dur time.Duration
	if interval > 0 && int64(interval) <= maxIntervalSec {
		dur = time.Duration(interval) * time.Second
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &Cleaner{
		interval:  dur,
		cleanFunc: cleanFunc,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start runs the cleaner background loop in a new goroutine.
// Calling Start multiple times or with an invalid interval / nil cleanFunc is safe and idempotent.
func (c *Cleaner) Start() {
	if c == nil || c.interval <= 0 || c.cleanFunc == nil {
		return
	}
	c.startOnce.Do(func() {
		c.wg.Add(1)
		go c.run()
	})
}

// Stop stops the background cleaner loop and waits for any in-flight cleanup to finish.
func (c *Cleaner) Stop() {
	if c == nil {
		return
	}
	c.cancel()
	c.wg.Wait()
}

// run executes cleanFunc periodically on every ticker interval until context is canceled.
func (c *Cleaner) run() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanFunc()
		case <-c.ctx.Done():
			return
		}
	}
}
