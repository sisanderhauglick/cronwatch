package metrics

import (
	"sync"
	"time"
)

// RateLimitEntry tracks alert suppression state for a single job.
type RateLimitEntry struct {
	LastAlerted time.Time
	Count       int
}

// AlertRateLimiter suppresses repeated alerts for the same job within a
// configurable cooldown window.
type AlertRateLimiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	entries  map[string]*RateLimitEntry
}

// NewAlertRateLimiter creates a new AlertRateLimiter with the given cooldown.
func NewAlertRateLimiter(cooldown time.Duration) *AlertRateLimiter {
	return &AlertRateLimiter{
		cooldown: cooldown,
		entries:  make(map[string]*RateLimitEntry),
	}
}

// Allow returns true if an alert for the given job should be sent.
// It updates internal state when the alert is allowed.
func (r *AlertRateLimiter) Allow(job string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[job]
	if !ok {
		r.entries[job] = &RateLimitEntry{LastAlerted: now, Count: 1}
		return true
	}

	if now.Sub(entry.LastAlerted) >= r.cooldown {
		entry.LastAlerted = now
		entry.Count++
		return true
	}

	return false
}

// Reset clears the rate limit state for a specific job.
func (r *AlertRateLimiter) Reset(job string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, job)
}

// Stats returns a copy of the current entry for a job, and whether it exists.
func (r *AlertRateLimiter) Stats(job string) (RateLimitEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[job]
	if !ok {
		return RateLimitEntry{}, false
	}
	return *entry, true
}
