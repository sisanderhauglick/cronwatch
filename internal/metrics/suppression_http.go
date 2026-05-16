package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type suppressionRuleRequest struct {
	JobName string `json:"job_name"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Reason  string `json:"reason"`
}

// NewSuppressionHandler returns an HTTP handler for managing suppression rules.
// GET  /suppressions  — list active rules
// POST /suppressions  — add a new rule
func NewSuppressionHandler(sm *SuppressionManager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/suppressions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rules := sm.Active()
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(rules); err != nil {
				http.Error(w, "encode error", http.StatusInternalServerError)
			}

		case http.MethodPost:
			var req suppressionRuleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			start, err := time.Parse(time.RFC3339, req.Start)
			if err != nil {
				http.Error(w, "invalid start time", http.StatusBadRequest)
				return
			}
			end, err := time.Parse(time.RFC3339, req.End)
			if err != nil {
				http.Error(w, "invalid end time", http.StatusBadRequest)
				return
			}
			if req.JobName == "" {
				http.Error(w, "job_name required", http.StatusBadRequest)
				return
			}
			sm.Add(SuppressionRule{
				JobName: req.JobName,
				Start:   start,
				End:     end,
				Reason:  req.Reason,
			})
			w.WriteHeader(http.StatusCreated)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
