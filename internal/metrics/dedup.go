package metrics

import (
	"sync"
	"time"
)

// DedupEntry tracks the last alert sent for a job+reason pair.
type DedupEntry struct {
	JobName  string
	Reason   string
	LastSent time.Time
}

// DedupManager suppresses duplicate alerts within a configurable window.
type DedupManager struct {
	mu      sync.Mutex
	window  time.Duration
	entries map[string]DedupEntry
}

// NewDedupManager creates a DedupManager with the given dedup window.
func NewDedupManager(window time.Duration) *DedupManager {
	return &DedupManager{
		window:  window,
		entries: make(map[string]DedupEntry),
	}
}

func dedupKey(job, reason string) string {
	return job + "::" + reason
}

// IsDuplicate returns true if an identical alert was sent within the window.
// If not a duplicate, it records the current time and returns false.
func (d *DedupManager) IsDuplicate(job, reason string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := dedupKey(job, reason)
	if entry, ok := d.entries[key]; ok {
		if now.Sub(entry.LastSent) < d.window {
			return true
		}
	}
	d.entries[key] = DedupEntry{JobName: job, Reason: reason, LastSent: now}
	return false
}

// Reset clears the dedup state for a specific job+reason pair.
func (d *DedupManager) Reset(job, reason string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, dedupKey(job, reason))
}

// Prune removes entries older than the dedup window.
func (d *DedupManager) Prune(now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, e := range d.entries {
		if now.Sub(e.LastSent) >= d.window {
			delete(d.entries, k)
		}
	}
}

// Entries returns a snapshot of all current dedup entries.
func (d *DedupManager) Entries() []DedupEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DedupEntry, 0, len(d.entries))
	for _, e := range d.entries {
		out = append(out, e)
	}
	return out
}
