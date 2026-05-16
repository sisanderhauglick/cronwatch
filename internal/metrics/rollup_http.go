package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// RollupHandler returns an HTTP handler that serves hourly rollup data.
func RollupHandler(a *RollupAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hours := 24
		if v := r.URL.Query().Get("hours"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 168 {
				hours = n
			}
		}

		buckets := a.HourlyRollup(hours, time.Now())
		if buckets == nil {
			buckets = []RollupBucket{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buckets)
	}
}
