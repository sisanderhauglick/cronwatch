package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse is the JSON envelope returned by HealthHandler.
type HealthResponse struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Jobs        []JobHealth  `json:"jobs"`
	Overall     HealthStatus `json:"overall"`
}

// HealthHandler serves a JSON summary of every job's current health.
func HealthHandler(reg *Registry, eval *HealthEvaluator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		jobs := reg.Jobs()
		health := make([]JobHealth, 0, len(jobs))

		overall := HealthOK
		for _, name := range jobs {
			snap, ok := reg.Get(name)
			if !ok {
				continue
			}
			h := eval.Evaluate(Snapshot{
				JobName:  name,
				Seen:     snap.Seen,
				Missed:   snap.Missed,
				Failed:   snap.Failed,
				LastSeen: snap.LastSeen,
			}, now)
			health = append(health, h)

			if h.Status == HealthDown {
				overall = HealthDown
			} else if h.Status == HealthDegraded && overall != HealthDown {
				overall = HealthDegraded
			}
		}

		if len(health) == 0 {
			overall = HealthUnknown
		}

		resp := HealthResponse{
			GeneratedAt: now,
			Jobs:        health,
			Overall:     overall,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
