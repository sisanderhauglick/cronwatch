package metrics

import (
	"encoding/json"
	"net/http"
	"sort"
)

// latencyResponse is the JSON shape returned by LatencyHandler.
type latencyResponse struct {
	Job string  `json:"job"`
	P50 float64 `json:"p50_ms"`
	P95 float64 `json:"p95_ms"`
	P99 float64 `json:"p99_ms"`
}

// LatencyHandler returns an HTTP handler that exposes per-job latency percentiles.
func LatencyHandler(tracker *LatencyTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs := tracker.Jobs()
		sort.Strings(jobs)

		results := make([]latencyResponse, 0, len(jobs))
		for _, job := range jobs {
			p50, p95, p99 := tracker.Stats(job)
			results = append(results, latencyResponse{
				Job: job,
				P50: float64(p50.Milliseconds()),
				P95: float64(p95.Milliseconds()),
				P99: float64(p99.Milliseconds()),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
