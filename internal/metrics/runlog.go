package metrics

import (
	"net/http"
	"sort"
	"sync"
	"time"
)

// RunEntry records a single execution attempt for a cron job.
type RunEntry struct {
	Job       string    `json:"job"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"` // "ok", "failed", "missed"
	Message   string    `json:"message,omitempty"`
}

// RunLog stores a bounded, in-memory log of recent job run entries.
type RunLog struct {
	mu      sync.Mutex
	entries []RunEntry
	maxSize int
}

// NewRunLog creates a RunLog that retains at most maxSize entries.
func NewRunLog(maxSize int) *RunLog {
	if maxSize <= 0 {
		maxSize = 500
	}
	return &RunLog{maxSize: maxSize}
}

// Record appends a new entry, evicting the oldest if over capacity.
func (r *RunLog) Record(entry RunEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
	if len(r.entries) > r.maxSize {
		r.entries = r.entries[len(r.entries)-r.maxSize:]
	}
}

// Query returns entries filtered by job name and/or time range.
// Pass an empty job string to include all jobs.
func (r *RunLog) Query(job string, from, to time.Time) []RunEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []RunEntry
	for _, e := range r.entries {
		if job != "" && e.Job != job {
			continue
		}
		if !from.IsZero() && e.StartedAt.Before(from) {
			continue
		}
		if !to.IsZero() && e.StartedAt.After(to) {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.Before(out[j].StartedAt)
	})
	return out
}

// RunLogHandler returns an HTTP handler that serves recent run log entries as JSON.
func RunLogHandler(rl *RunLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		entries := rl.Query(job, time.Time{}, time.Time{})
		if entries == nil {
			entries = []RunEntry{}
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, entries)
	}
}
