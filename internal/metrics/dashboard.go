package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// DashboardSummary holds a high-level overview of all monitored jobs.
type DashboardSummary struct {
	GeneratedAt  time.Time        `json:"generated_at"`
	TotalJobs    int              `json:"total_jobs"`
	Healthy      int              `json:"healthy"`
	Degraded     int              `json:"degraded"`
	Jobs         []JobHealthEntry `json:"jobs"`
}

// JobHealthEntry describes the health of a single job.
type JobHealthEntry struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"` // "ok", "missed", "failed"
	SuccessRate float64 `json:"success_rate"`
	SeenTotal   int64   `json:"seen_total"`
	MissedTotal int64   `json:"missed_total"`
	FailedTotal int64   `json:"failed_total"`
}

// DashboardHandler returns an HTTP handler that renders a JSON dashboard
// summary derived from the provided Registry.
func DashboardHandler(r *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		snap := r.Snapshot()

		entries := make([]JobHealthEntry, 0, len(snap))
		healthy := 0

		for name, s := range snap {
			total := s.Seen + s.Failed
			var rate float64
			if total > 0 {
				rate = float64(s.Seen) / float64(total)
			}

			status := "ok"
			if s.Failed > 0 {
				status = "failed"
			} else if s.Missed > 0 {
				status = "missed"
			}

			if status == "ok" {
				healthy++
			}

			entries = append(entries, JobHealthEntry{
				Name:        name,
				Status:      status,
				SuccessRate: rate,
				SeenTotal:   s.Seen,
				MissedTotal: s.Missed,
				FailedTotal: s.Failed,
			})
		}

		summary := DashboardSummary{
			GeneratedAt: time.Now().UTC(),
			TotalJobs:   len(entries),
			Healthy:     healthy,
			Degraded:    len(entries) - healthy,
			Jobs:        entries,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}
