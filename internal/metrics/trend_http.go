package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// TrendHandler serves a JSON array of JobTrend values.
// GET /metrics/trends?window=1h&threshold=0.05
func TrendHandler(analyzer *TrendAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		trends := analyzer.Analyze(now)
		if trends == nil {
			trends = []JobTrend{}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(trends); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
