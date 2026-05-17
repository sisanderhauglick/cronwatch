package metrics

import (
	"sync"
	"time"
)

// WatchdogEntry tracks the last heartbeat for a job.
type WatchdogEntry struct {
	Job       string
	LastBeat  time.Time
	Timeout   time.Duration
	Triggered bool
}

// WatchdogManager monitors heartbeats and reports jobs that have gone silent.
type WatchdogManager struct {
	mu      sync.Mutex
	entries map[string]*WatchdogEntry
	now     func() time.Time
}

// NewWatchdogManager creates a new WatchdogManager.
func NewWatchdogManager() *WatchdogManager {
	return &WatchdogManager{
		entries: make(map[string]*WatchdogEntry),
		now:     time.Now,
	}
}

// Register registers a job with a heartbeat timeout.
func (w *WatchdogManager) Register(job string, timeout time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.entries[job]; !ok {
		w.entries[job] = &WatchdogEntry{
			Job:     job,
			Timeout: timeout,
		}
	}
}

// Beat records a heartbeat for the given job.
func (w *WatchdogManager) Beat(job string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	e, ok := w.entries[job]
	if !ok {
		return
	}
	e.LastBeat = w.now()
	e.Triggered = false
}

// Expired returns all jobs whose heartbeat timeout has elapsed and marks them triggered.
func (w *WatchdogManager) Expired() []WatchdogEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.now()
	var out []WatchdogEntry
	for _, e := range w.entries {
		if e.LastBeat.IsZero() {
			continue
		}
		if !e.Triggered && now.Sub(e.LastBeat) > e.Timeout {
			e.Triggered = true
			out = append(out, *e)
		}
	}
	return out
}

// All returns a snapshot of all registered entries.
func (w *WatchdogManager) All() []WatchdogEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]WatchdogEntry, 0, len(w.entries))
	for _, e := range w.entries {
		out = append(out, *e)
	}
	return out
}
