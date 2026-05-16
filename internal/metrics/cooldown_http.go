package metrics

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// NewCooldownHandler returns an HTTP handler for the cooldown manager.
// GET  /metrics/cooldown        — list all entries
// POST /metrics/cooldown/{job}  — activate cooldown (body: {"duration":"30s"})
// DELETE /metrics/cooldown/{job} — reset cooldown
func NewCooldownHandler(mgr *CooldownManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimPrefix(r.URL.Path, "/")

		switch r.Method {
		case http.MethodGet:
			entries := mgr.All()
			w.Header().Set("Content-Type", "application/json")
			if entries == nil {
				entries = []CooldownEntry{}
			}
			_ = json.NewEncoder(w).Encode(entries)

		case http.MethodPost:
			if job == "" {
				http.Error(w, "job name required", http.StatusBadRequest)
				return
			}
			var body struct {
				Duration string `json:"duration"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(body.Duration)
			if err != nil || d <= 0 {
				http.Error(w, "invalid duration", http.StatusBadRequest)
				return
			}
			mgr.SetCooldown(job, d)
			mgr.Activate(job)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, "job name required", http.StatusBadRequest)
				return
			}
			mgr.Reset(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
