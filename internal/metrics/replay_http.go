package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// ReplayHandler returns an HTTP handler that serves alert replay data.
// Query params:
//
//	job   - optional job name filter (omit for all jobs)
//	from  - RFC3339 start time (default: 24h ago)
//	to    - RFC3339 end time   (default: now)
func ReplayHandler(ra *ReplayAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		from := now.Add(-24 * time.Hour)
		to := now

		q := r.URL.Query()
		if s := q.Get("from"); s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				from = t
			}
		}
		if s := q.Get("to"); s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				to = t
			}
		}

		var entries []ReplayEntry
		if job := q.Get("job"); job != "" {
			entries = ra.Replay(job, from, to)
		} else {
			entries = ra.ReplayAll(from, to)
		}

		if entries == nil {
			entries = []ReplayEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries) //nolint:errcheck
	}
}
