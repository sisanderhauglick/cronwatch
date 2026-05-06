package metrics

import (
	"sync"
	"time"
)

// AlertEvent records a single alert that was dispatched.
type AlertEvent struct {
	JobName   string
	Reason    string // "missed" or "failed"
	SentAt    time.Time
	Notifiers []string // names of notifiers that were triggered
}

// AlertTracker keeps an in-memory log of recent alert events so the
// dashboard and API can surface them without querying external systems.
type AlertTracker struct {
	mu     sync.RWMutex
	events []AlertEvent
	maxLen int
}

// NewAlertTracker returns an AlertTracker that retains up to maxLen events.
func NewAlertTracker(maxLen int) *AlertTracker {
	if maxLen <= 0 {
		maxLen = 200
	}
	return &AlertTracker{maxLen: maxLen}
}

// Record appends a new alert event, evicting the oldest if capacity is reached.
func (t *AlertTracker) Record(ev AlertEvent) {
	if ev.SentAt.IsZero() {
		ev.SentAt = time.Now()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.events) >= t.maxLen {
		t.events = t.events[1:]
	}
	t.events = append(t.events, ev)
}

// Recent returns up to n most-recent alert events, newest first.
func (t *AlertTracker) Recent(n int) []AlertEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if n <= 0 || n > len(t.events) {
		n = len(t.events)
	}
	out := make([]AlertEvent, n)
	base := len(t.events) - n
	for i := 0; i < n; i++ {
		out[i] = t.events[base+n-1-i] // reverse order
	}
	return out
}

// CountByJob returns how many alerts have been recorded for each job name.
func (t *AlertTracker) CountByJob() map[string]int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	counts := make(map[string]int, len(t.events))
	for _, ev := range t.events {
		counts[ev.JobName]++
	}
	return counts
}
