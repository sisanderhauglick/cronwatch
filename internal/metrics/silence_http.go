package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type silenceRequest struct {
	JobName  string `json:"job_name"`
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
}

type silenceResponse struct {
	JobName string    `json:"job_name"`
	Until   time.Time `json:"until"`
	Reason  string    `json:"reason"`
}

// SilenceHandler returns an HTTP handler for managing silence rules.
// POST adds a new rule; GET lists active rules.
func SilenceHandler(mgr *SilenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req silenceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
				return
			}
			dur, err := time.ParseDuration(req.Duration)
			if err != nil || dur <= 0 {
				http.Error(w, `{"error":"invalid duration"}`, http.StatusBadRequest)
				return
			}
			now := time.Now()
			rule := SilenceRule{
				JobName:   req.JobName,
				StartTime: now,
				EndTime:   now.Add(dur),
				Reason:    req.Reason,
			}
			mgr.Add(rule)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(silenceResponse{
				JobName: rule.JobName,
				Until:   rule.EndTime,
				Reason:  rule.Reason,
			})
		case http.MethodGet:
			mgr.Prune()
			rules := mgr.List()
			out := make([]silenceResponse, 0, len(rules))
			for _, r := range rules {
				out = append(out, silenceResponse{
					JobName: r.JobName,
					Until:   r.EndTime,
					Reason:  r.Reason,
				})
			}
			json.NewEncoder(w).Encode(out)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}
