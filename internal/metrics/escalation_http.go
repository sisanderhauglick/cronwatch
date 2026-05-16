package metrics

import (
	"encoding/json"
	"net/http"
)

type escalationResponse struct {
	Job   string `json:"job"`
	Level string `json:"level"`
	Since string `json:"since,omitempty"`
}

// EscalationHandler returns an HTTP handler that reports current escalation
// states for all tracked jobs.
func EscalationHandler(mgr *EscalationManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		states := mgr.All()
		resp := make([]escalationResponse, 0, len(states))
		for _, s := range states {
			if s.Level == LevelNone {
				continue
			}
			er := escalationResponse{
				Job:   s.JobName,
				Level: s.Level.String(),
			}
			if !s.Since.IsZero() {
				er.Since = s.Since.UTC().Format("2006-01-02T15:04:05Z")
			}
			resp = append(resp, er)
		}
		if resp == nil {
			resp = []escalationResponse{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
