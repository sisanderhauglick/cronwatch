package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type uptimeResponse struct {
	Job       string  `json:"job"`
	Since     string  `json:"since"`
	UptimePct float64 `json:"uptime_pct"`
}

// UptimeHandler returns an HTTP handler that reports per-job uptime percentages.
func UptimeHandler(tracker *UptimeTracker, window time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := tracker.Compute(window)

		responses := make([]uptimeResponse, 0, len(results))
		for _, res := range results {
			responses = append(responses, uptimeResponse{
				Job:       res.Job,
				Since:     res.Since.UTC().Format(time.RFC3339),
				UptimePct: res.UptimePct,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
