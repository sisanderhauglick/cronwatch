package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type rateLimitStatus struct {
	Job         string    `json:"job"`
	LastAlerted time.Time `json:"last_alerted"`
	AlertCount  int       `json:"alert_count"`
	CooldownSec float64   `json:"cooldown_seconds"`
	Suppressed  bool      `json:"suppressed"`
}

// RateLimitHandler returns an HTTP handler that exposes rate limiter state
// for all known jobs derived from the provided registry.
func RateLimitHandler(reg *Registry, limiter *AlertRateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		snap := reg.Snapshot()

		results := make([]rateLimitStatus, 0, len(snap))
		for job := range snap {
			entry, ok := limiter.Stats(job)
			if !ok {
				results = append(results, rateLimitStatus{
					Job:         job,
					CooldownSec: limiter.cooldown.Seconds(),
					Suppressed:  false,
				})
				continue
			}
			suppressed := now.Sub(entry.LastAlerted) < limiter.cooldown
			results = append(results, rateLimitStatus{
				Job:         job,
				LastAlerted: entry.LastAlerted,
				AlertCount:  entry.Count,
				CooldownSec: limiter.cooldown.Seconds(),
				Suppressed:  suppressed,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
