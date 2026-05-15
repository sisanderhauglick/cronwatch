package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultChangelogLimit = 50

// ChangelogHandler returns an HTTP handler that exposes recent status-change
// events from the given ChangeLog as JSON.
func ChangelogHandler(cl *ChangeLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := defaultChangelogLimit
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil && v > 0 {
				limit = v
			}
		}

		events := cl.Recent(limit)
		if events == nil {
			events = []ChangeEvent{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events) //nolint:errcheck
	}
}
