package clean

import (
	"context"
	"time"
)

// Cleaner executes a periodic cleanup function on a fixed interval until stopped or canceled.
type Cleaner struct {
	interval  int // in seconds
	cleanFunc func()
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewCleaner creates a new Cleaner with the given interval and cleanup function.
// If parentCtx is nil, context.Background() is used.
func NewCleaner(parentCtx context.Context, interval int, cleanFunc func()) *Cleaner {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &Cleaner{
		interval:  interval,
		cleanFunc: cleanFunc,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start runs the cleaner background loop in a new goroutine.
func (c *Cleaner) Start() {
	go c.run()
}

// Stop stops the background cleaner loop.
func (c *Cleaner) Stop() {
	c.cancel()
}

// run executes cleanFunc periodically on every ticker interval until context is canceled.
func (c *Cleaner) run() {
	ticker := time.NewTicker(time.Duration(c.interval) * time.Second)
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
