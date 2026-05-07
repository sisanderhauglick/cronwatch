package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// SLAHandler serves SLA compliance results as JSON.
// It expects a list of SLATarget values configured at construction time.
type SLAHandler struct {
	evaluator *SLAEvaluator
	targets   []SLATarget
}

// NewSLAHandler creates an HTTP handler for SLA evaluation.
func NewSLAHandler(tracker *UptimeTracker, window time.Duration, targets []SLATarget) *SLAHandler {
	return &SLAHandler{
		evaluator: NewSLAEvaluator(tracker, window),
		targets:   targets,
	}
}

func (h *SLAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	results := h.evaluator.Evaluate(h.targets)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
