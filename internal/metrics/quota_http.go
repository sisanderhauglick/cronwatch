package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type quotaStatsResponse struct {
	Job       string    `json:"job"`
	Count     int       `json:"count"`
	WindowEnd time.Time `json:"window_end,omitempty"`
	Exhausted bool      `json:"exhausted"`
	MaxAlerts int       `json:"max_alerts"`
}

// QuotaHandler returns an HTTP handler that reports per-job quota stats.
func QuotaHandler(qm *QuotaManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")

		var results []quotaStatsResponse

		if job != "" {
			count, windowEnd := qm.Stats(job)
			results = append(results, quotaStatsResponse{
				Job:       job,
				Count:     count,
				WindowEnd: windowEnd,
				Exhausted: count >= qm.policy.MaxAlerts && !windowEnd.IsZero(),
				MaxAlerts: qm.policy.MaxAlerts,
			})
		} else {
			qm.mu.Lock()
			now := qm.now()
			for name, e := range qm.entries {
				if now.After(e.windowEnd) {
					continue
				}
				results = append(results, quotaStatsResponse{
					Job:       name,
					Count:     e.count,
					WindowEnd: e.windowEnd,
					Exhausted: e.count >= qm.policy.MaxAlerts,
					MaxAlerts: qm.policy.MaxAlerts,
				})
			}
			qm.mu.Unlock()
		}

		if results == nil {
			results = []quotaStatsResponse{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
