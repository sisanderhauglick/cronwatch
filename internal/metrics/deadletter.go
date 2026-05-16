package metrics

import (
	"sync"
	"time"
)

// DeadLetterEntry represents an alert that failed to deliver and was queued.
type DeadLetterEntry struct {
	Job       string    `json:"job"`
	Reason    string    `json:"reason"`
	Payload   string    `json:"payload"`
	FailedAt  time.Time `json:"failed_at"`
	Attempts  int       `json:"attempts"`
}

// DeadLetterQueue stores alerts that could not be delivered.
type DeadLetterQueue struct {
	mu      sync.Mutex
	entries []DeadLetterEntry
	maxSize int
}

// NewDeadLetterQueue creates a new DeadLetterQueue with the given capacity.
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &DeadLetterQueue{maxSize: maxSize}
}

// Push adds a failed alert entry to the queue, evicting the oldest if full.
func (q *DeadLetterQueue) Push(job, reason, payload string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	entry := DeadLetterEntry{
		Job:      job,
		Reason:   reason,
		Payload:  payload,
		FailedAt: time.Now(),
		Attempts: 1,
	}
	for i, e := range q.entries {
		if e.Job == job && e.Reason == reason {
			q.entries[i].Attempts++
			q.entries[i].FailedAt = entry.FailedAt
			return
		}
	}
	if len(q.entries) >= q.maxSize {
		q.entries = q.entries[1:]
	}
	q.entries = append(q.entries, entry)
}

// All returns a copy of all current dead-letter entries.
func (q *DeadLetterQueue) All() []DeadLetterEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]DeadLetterEntry, len(q.entries))
	copy(out, q.entries)
	return out
}

// Remove deletes the entry matching the given job and reason.
func (q *DeadLetterQueue) Remove(job, reason string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, e := range q.entries {
		if e.Job == job && e.Reason == reason {
			q.entries = append(q.entries[:i], q.entries[i+1:]...)
			return true
		}
	}
	return false
}

// Len returns the number of entries currently in the queue.
func (q *DeadLetterQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}
