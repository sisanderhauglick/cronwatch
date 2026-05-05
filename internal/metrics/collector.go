package metrics

import (
	"sync"
	"time"
)

// Collector gathers periodic snapshots from a Registry and retains them
// according to a RetentionPolicy.
type Collector struct {
	mu        sync.Mutex
	registry  *Registry
	snapshots []Snapshot
	policy    RetentionPolicy
}

// NewCollector creates a Collector backed by the given Registry.
func NewCollector(r *Registry, policy RetentionPolicy) *Collector {
	return &Collector{
		registry: r,
		policy:   policy,
	}
}

// Collect takes a snapshot of the current registry state and stores it.
func (c *Collector) Collect() {
	snap := Snapshot{
		CollectedAt: time.Now(),
		Jobs:        c.registry.Snapshot(),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshots = append(c.snapshots, snap)
	c.snapshots = c.policy.apply(c.snapshots)
}

// Latest returns the most recently collected snapshot, or false if none.
func (c *Collector) Latest() (Snapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.snapshots) == 0 {
		return Snapshot{}, false
	}
	return c.snapshots[len(c.snapshots)-1], true
}

// All returns a copy of all retained snapshots.
func (c *Collector) All() []Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]Snapshot, len(c.snapshots))
	copy(result, c.snapshots)
	return result
}

// Snapshot holds a point-in-time view of all job metrics.
type Snapshot struct {
	CollectedAt time.Time
	Jobs        map[string]JobStats
}
