package metrics

import (
	"sync"
	"time"
)

// ThrottleEntry tracks per-job alert throttle state.
type ThrottleEntry struct {
	JobName    string
	Suppressed int
	LastSent   time.Time
	NextAllowed time.Time
}

// ThrottleManager limits alert frequency per job using a token-bucket-like
// cooldown. Unlike the rate limiter, it tracks suppression counts and exposes
// per-job state for observability.
type ThrottleManager struct {
	mu       sync.Mutex
	entries  map[string]*ThrottleEntry
	cooldown time.Duration
	now      func() time.Time
}

// NewThrottleManager creates a ThrottleManager with the given cooldown window.
func NewThrottleManager(cooldown time.Duration) *ThrottleManager {
	return &ThrottleManager{
		entries:  make(map[string]*ThrottleEntry),
		cooldown: cooldown,
		now:      time.Now,
	}
}

// Allow returns true if an alert for jobName should be sent now.
// If suppressed, the suppression counter is incremented.
func (t *ThrottleManager) Allow(jobName string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	e, ok := t.entries[jobName]
	if !ok {
		t.entries[jobName] = &ThrottleEntry{
			JobName:     jobName,
			LastSent:    now,
			NextAllowed: now.Add(t.cooldown),
		}
		return true
	}

	if now.Before(e.NextAllowed) {
		e.Suppressed++
		return false
	}

	e.LastSent = now
	e.NextAllowed = now.Add(t.cooldown)
	e.Suppressed = 0
	return true
}

// Stats returns a copy of the ThrottleEntry for a job, and whether it exists.
func (t *ThrottleManager) Stats(jobName string) (ThrottleEntry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[jobName]
	if !ok {
		return ThrottleEntry{}, false
	}
	return *e, true
}

// All returns a snapshot of all throttle entries.
func (t *ThrottleManager) All() []ThrottleEntry {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]ThrottleEntry, 0, len(t.entries))
	for _, e := range t.entries {
		out = append(out, *e)
	}
	return out
}

// Reset clears the throttle state for a job.
func (t *ThrottleManager) Reset(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, jobName)
}
