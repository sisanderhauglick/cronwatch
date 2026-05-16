package metrics

import (
	"sync"
	"time"
)

// CheckpointEntry records the last successful completion time for a job.
type CheckpointEntry struct {
	Job       string    `json:"job"`
	LastOK    time.Time `json:"last_ok"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckpointStore tracks the last known-good run time for each monitored job.
type CheckpointStore struct {
	mu      sync.RWMutex
	entries map[string]CheckpointEntry
}

// NewCheckpointStore returns an initialised CheckpointStore.
func NewCheckpointStore() *CheckpointStore {
	return &CheckpointStore{
		entries: make(map[string]CheckpointEntry),
	}
}

// Record marks job as having completed successfully at t.
func (c *CheckpointStore) Record(job string, t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[job] = CheckpointEntry{
		Job:       job,
		LastOK:    t,
		UpdatedAt: time.Now(),
	}
}

// Get returns the checkpoint for job and whether it exists.
func (c *CheckpointStore) Get(job string) (CheckpointEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[job]
	return e, ok
}

// All returns a snapshot of every checkpoint entry.
func (c *CheckpointStore) All() []CheckpointEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]CheckpointEntry, 0, len(c.entries))
	for _, e := range c.entries {
		out = append(out, e)
	}
	return out
}

// StaleBefore returns jobs whose last successful run is older than cutoff.
func (c *CheckpointStore) StaleBefore(cutoff time.Time) []CheckpointEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []CheckpointEntry
	for _, e := range c.entries {
		if e.LastOK.Before(cutoff) {
			out = append(out, e)
		}
	}
	return out
}
