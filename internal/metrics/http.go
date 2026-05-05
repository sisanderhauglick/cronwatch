package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

// Handler returns an http.Handler that exposes job metrics as JSON.
// It is intended to be mounted at a path such as /metrics.
func Handler(r *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		snap := r.Snapshot()
		sort.Slice(snap, func(i, j int) bool {
			return snap[i].Name < snap[j].Name
		})

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(snap); err != nil {
			http.Error(w, fmt.Sprintf("encode error: %v", err), http.StatusInternalServerError)
		}
	})
}
