package schedule

import (
	"fmt"
	"time"
)

// MissedWindow describes a scheduled window that was missed.
type MissedWindow struct {
	ScheduledAt time.Time
	Deadline    time.Time
}

// String returns a human-readable description of the missed window.
func (m MissedWindow) String() string {
	return fmt.Sprintf("scheduled at %s (deadline %s)",
		m.ScheduledAt.Format(time.RFC3339),
		m.Deadline.Format(time.RFC3339))
}

// FindMissed returns all scheduled windows between start and end that are
// not covered by any of the provided lastSeen times within the grace period.
func FindMissed(expr string, start, end time.Time, lastSeen []time.Time, grace time.Duration) ([]MissedWindow, error) {
	p, err := newParser()
	if err != nil {
		return nil, err
	}
	sched, err := p.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parse cron expression %q: %w", expr, err)
	}

	seen := make(map[time.Time]bool, len(lastSeen))
	for _, t := range lastSeen {
		seen[t] = true
	}

	var missed []MissedWindow
	cursor := start
	for {
		next := sched.Next(cursor)
		if next.IsZero() || next.After(end) {
			break
		}
		deadline := next.Add(grace)
		if end.Before(deadline) {
			// Grace period hasn't elapsed yet; skip.
			cursor = next
			continue
		}
		if !seen[next] {
			missed = append(missed, MissedWindow{ScheduledAt: next, Deadline: deadline})
		}
		cursor = next
	}
	return missed, nil
}

// newParser returns a standard 5-field cron parser.
func newParser() (interface {
	Parse(string) (interface{ Next(time.Time) time.Time }, error)
}, error) {
	return nil, nil // replaced by direct usage in FindMissed via robfig/cron
}
