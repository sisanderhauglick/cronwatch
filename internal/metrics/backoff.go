package metrics

import (
	"math"
	"sync"
	"time"
)

// BackoffEntry tracks the current backoff state for a job's alert delivery.
type BackoffEntry struct {
	Job       string
	Attempts  int
	NextRetry time.Time
	LastError string
}

// BackoffManager applies exponential backoff to alert retries per job.
type BackoffManager struct {
	mu       sync.Mutex
	entries  map[string]*BackoffEntry
	baseWait time.Duration
	maxWait  time.Duration
}

// NewBackoffManager creates a BackoffManager with the given base and max wait durations.
func NewBackoffManager(base, max time.Duration) *BackoffManager {
	return &BackoffManager{
		entries:  make(map[string]*BackoffEntry),
		baseWait: base,
		maxWait:  max,
	}
}

// Allow returns true if the job is eligible for a retry alert at now.
func (b *BackoffManager) Allow(job string, now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.entries[job]
	if !ok {
		return true
	}
	return now.After(e.NextRetry) || now.Equal(e.NextRetry)
}

// RecordFailure records a failed alert delivery and advances the backoff.
func (b *BackoffManager) RecordFailure(job, reason string, now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.entries[job]
	if !ok {
		e = &BackoffEntry{Job: job}
		b.entries[job] = e
	}
	e.Attempts++
	e.LastError = reason
	wait := time.Duration(float64(b.baseWait) * math.Pow(2, float64(e.Attempts-1)))
	if wait > b.maxWait {
		wait = b.maxWait
	}
	e.NextRetry = now.Add(wait)
}

// Reset clears the backoff state for a job after a successful delivery.
func (b *BackoffManager) Reset(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, job)
}

// All returns a snapshot of all current backoff entries.
func (b *BackoffManager) All() []BackoffEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]BackoffEntry, 0, len(b.entries))
	for _, e := range b.entries {
		out = append(out, *e)
	}
	return out
}
