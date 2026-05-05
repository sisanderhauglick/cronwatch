package metrics

import (
	"sync"
	"time"
)

// Collector periodically snapshots metrics and exposes a rolling window
// of recent job health summaries.
type Collector struct {
	mu       sync.RWMutex
	reg      *Registry
	window   time.Duration
	snapshots []TimedSnapshot
}

// TimedSnapshot pairs a Snapshot with the time it was taken.
type TimedSnapshot struct {
	At       time.Time
	Snapshot map[string]JobStats
}

// NewCollector creates a Collector that retains snapshots within the given
// rolling window duration.
func NewCollector(reg *Registry, window time.Duration) *Collector {
	return &Collector{
		reg:    reg,
		window: window,
	}
}

// Collect takes a new snapshot and prunes entries older than the window.
func (c *Collector) Collect() {
	now := time.Now()
	snap := c.reg.Snapshot()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshots = append(c.snapshots, TimedSnapshot{At: now, Snapshot: snap})
	c.prune(now)
}

// prune removes snapshots outside the rolling window. Must be called with lock held.
func (c *Collector) prune(now time.Time) {
	cutoff := now.Add(-c.window)
	i := 0
	for i < len(c.snapshots) && c.snapshots[i].At.Before(cutoff) {
		i++
	}
	c.snapshots = c.snapshots[i:]
}

// Latest returns the most recent TimedSnapshot, and false if none exist.
func (c *Collector) Latest() (TimedSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.snapshots) == 0 {
		return TimedSnapshot{}, false
	}
	return c.snapshots[len(c.snapshots)-1], true
}

// All returns a copy of all retained snapshots.
func (c *Collector) All() []TimedSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]TimedSnapshot, len(c.snapshots))
	copy(out, c.snapshots)
	return out
}
