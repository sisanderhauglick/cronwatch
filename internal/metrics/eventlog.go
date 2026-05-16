package metrics

import (
	"sync"
	"time"
)

// EventSeverity represents the severity level of a logged event.
type EventSeverity string

const (
	SeverityInfo  EventSeverity = "info"
	SeverityWarn  EventSeverity = "warn"
	SeverityError EventSeverity = "error"
)

// EventEntry is a single structured event record.
type EventEntry struct {
	Timestamp time.Time     `json:"timestamp"`
	Job       string        `json:"job"`
	Severity  EventSeverity `json:"severity"`
	Message   string        `json:"message"`
}

// EventLog stores recent structured events with a configurable capacity.
type EventLog struct {
	mu       sync.Mutex
	entries  []EventEntry
	maxSize  int
}

// NewEventLog creates an EventLog with the given maximum capacity.
func NewEventLog(maxSize int) *EventLog {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &EventLog{maxSize: maxSize}
}

// Record appends a new event, evicting the oldest entry when at capacity.
func (e *EventLog) Record(job string, severity EventSeverity, message string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry := EventEntry{
		Timestamp: time.Now().UTC(),
		Job:       job,
		Severity:  severity,
		Message:   message,
	}
	if len(e.entries) >= e.maxSize {
		e.entries = e.entries[1:]
	}
	e.entries = append(e.entries, entry)
}

// All returns a copy of all stored events.
func (e *EventLog) All() []EventEntry {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]EventEntry, len(e.entries))
	copy(out, e.entries)
	return out
}

// FilterByJob returns events matching the given job name.
func (e *EventLog) FilterByJob(job string) []EventEntry {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []EventEntry
	for _, ev := range e.entries {
		if ev.Job == job {
			out = append(out, ev)
		}
	}
	return out
}

// FilterBySeverity returns events matching the given severity.
func (e *EventLog) FilterBySeverity(sev EventSeverity) []EventEntry {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []EventEntry
	for _, ev := range e.entries {
		if ev.Severity == sev {
			out = append(out, ev)
		}
	}
	return out
}
