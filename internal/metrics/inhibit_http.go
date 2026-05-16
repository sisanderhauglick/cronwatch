package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type inhibitCheckResponse struct {
	Job       string `json:"job"`
	Inhibited bool   `json:"inhibited"`
}

type inhibitRuleRequest struct {
	SourceJob string `json:"source_job"`
	TargetJob string `json:"target_job"`
}

// InhibitHandler exposes inhibit check and rule management over HTTP.
type InhibitHandler struct {
	manager *InhibitManager
}

// NewInhibitHandler creates a handler backed by the given InhibitManager.
func NewInhibitHandler(m *InhibitManager) *InhibitHandler {
	return &InhibitHandler{manager: m}
}

// ServeHTTP routes GET (check) and POST (add rule) requests.
//
//	GET  /metrics/inhibit?job=<name>  — check if job is inhibited
//	POST /metrics/inhibit              — add a new rule (JSON body)
func (h *InhibitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		job := r.URL.Query().Get("job")
		inhibited := h.manager.IsInhibited(job, time.Now())
		json.NewEncoder(w).Encode(inhibitCheckResponse{Job: job, Inhibited: inhibited}) //nolint:errcheck

	case http.MethodPost:
		var req inhibitRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		if req.SourceJob == "" || req.TargetJob == "" {
			http.Error(w, `{"error":"source_job and target_job required"}`, http.StatusBadRequest)
			return
		}
		h.manager.AddRule(InhibitRule{SourceJob: req.SourceJob, TargetJob: req.TargetJob})
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"}) //nolint:errcheck

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
