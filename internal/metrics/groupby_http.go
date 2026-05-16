package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// GroupByHandler returns an HTTP handler that serves group-by aggregations.
// Callers pass ?key=<tagKey> and optionally ?window=<duration>.
func GroupByHandler(a *GroupByAnalyzer, defaultWindow time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, `{"error":"missing query param: key"}`, http.StatusBadRequest)
			return
		}

		win := defaultWindow
		if raw := r.URL.Query().Get("window"); raw != "" {
			if d, err := time.ParseDuration(raw); err == nil && d > 0 {
				win = d
			}
		}

		// Temporarily override window for this request.
		a.mu.Lock()
		prev := a.window
		a.window = win
		a.mu.Unlock()

		results := a.Summarize(key)

		a.mu.Lock()
		a.window = prev
		a.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
