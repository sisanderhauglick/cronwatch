package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// SummaryHandler returns an http.Handler that serves aggregated job
// statistics for a configurable look-back window.
//
// Query parameters:
//
//	window — duration string (default: "1h")
func SummaryHandler(agg *Aggregator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		windowStr := r.URL.Query().Get("window")
		if windowStr == "" {
			windowStr = "1h"
		}
		window, err := time.ParseDuration(windowStr)
		if err != nil {
			http.Error(w, "invalid window duration", http.StatusBadRequest)
			return
		}

		now := time.Now()
		summaries := agg.Summarize(now.Add(-window), now)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(summaries); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
