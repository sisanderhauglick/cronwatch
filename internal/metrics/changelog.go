package metrics

import (
	"sync"
	"time"
)

// ChangeEvent records a status transition for a job.
type ChangeEvent struct {
	Job       string    `json:"job"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ChangeLog tracks status transitions for monitored jobs.
type ChangeLog struct {
	mu     sync.Mutex
	events []ChangeEvent
	max    int
	last   map[string]string
}

// NewChangeLog creates a ChangeLog that retains at most maxEvents entries.
func NewChangeLog(maxEvents int) *ChangeLog {
	if maxEvents <= 0 {
		maxEvents = 200
	}
	return &ChangeLog{
		max:  maxEvents,
		last: make(map[string]string),
	}
}

// Record observes a job's current status and appends a ChangeEvent if the
// status differs from the previously recorded one.
func (c *ChangeLog) Record(job, status string, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prev, ok := c.last[job]
	if ok && prev == status {
		return
	}
	from := ""
	if ok {
		from = prev
	}
	c.last[job] = status
	c.events = append(c.events, ChangeEvent{
		Job:        job,
		From:       from,
		To:         status,
		OccurredAt: at,
	})
	if len(c.events) > c.max {
		c.events = c.events[len(c.events)-c.max:]
	}
}

// Recent returns up to n most recent change events, newest first.
func (c *ChangeLog) Recent(n int) []ChangeEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n <= 0 || len(c.events) == 0 {
		return nil
	}
	start := len(c.events) - n
	if start < 0 {
		start = 0
	}
	slice := c.events[start:]
	out := make([]ChangeEvent, len(slice))
	for i, e := range slice {
		out[len(slice)-1-i] = e
	}
	return out
}
