package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// SummaryHandler returns an HTTP handler that serves aggregated job summaries
// for the given time window using the provided Aggregator.
func SummaryHandler(agg *Aggregator, window time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		summaries := agg.Summarize(window)
		if summaries == nil {
			summaries = []JobSummary{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(summaries); err != nil {
			http.Error(w, "failed to encode summaries", http.StatusInternalServerError)
		}
	})
}
