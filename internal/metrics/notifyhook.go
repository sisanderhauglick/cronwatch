package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// NotifyHookEntry records a single outbound alert notification attempt.
type NotifyHookEntry struct {
	JobName   string    `json:"job_name"`
	Reason    string    `json:"reason"`
	Target    string    `json:"target"` // e.g. "webhook", "email", "log"
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NotifyHookLog stores recent notification delivery records.
type NotifyHookLog struct {
	mu      sync.Mutex
	entries []NotifyHookEntry
	maxSize int
}

// NewNotifyHookLog creates a NotifyHookLog with the given capacity.
func NewNotifyHookLog(maxSize int) *NotifyHookLog {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &NotifyHookLog{maxSize: maxSize}
}

// Record appends a delivery attempt to the log.
func (n *NotifyHookLog) Record(entry NotifyHookEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.entries = append(n.entries, entry)
	if len(n.entries) > n.maxSize {
		n.entries = n.entries[len(n.entries)-n.maxSize:]
	}
}

// All returns a copy of all stored entries, optionally filtered by job name.
func (n *NotifyHookLog) All(job string) []NotifyHookEntry {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]NotifyHookEntry, 0, len(n.entries))
	for _, e := range n.entries {
		if job == "" || e.JobName == job {
			out = append(out, e)
		}
	}
	return out
}

// NotifyHookHandler serves recent notification delivery records as JSON.
func NotifyHookHandler(log *NotifyHookLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		entries := log.All(job)
		if entries == nil {
			entries = []NotifyHookEntry{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	}
}
