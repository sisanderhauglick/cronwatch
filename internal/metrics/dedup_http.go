package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// DedupHandler exposes dedup state and supports DELETE to reset entries.
type DedupHandler struct {
	manager *DedupManager
}

// NewDedupHandler creates an HTTP handler for the DedupManager.
func NewDedupHandler(m *DedupManager) *DedupHandler {
	return &DedupHandler{manager: m}
}

func (h *DedupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DedupHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	entries := h.manager.Entries()
	type row struct {
		Job      string    `json:"job"`
		Reason   string    `json:"reason"`
		LastSent time.Time `json:"last_sent"`
	}
	out := make([]row, 0, len(entries))
	for _, e := range entries {
		out = append(out, row{Job: e.JobName, Reason: e.Reason, LastSent: e.LastSent})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *DedupHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	reason := r.URL.Query().Get("reason")
	if job == "" || reason == "" {
		http.Error(w, "job and reason query params required", http.StatusBadRequest)
		return
	}
	h.manager.Reset(job, reason)
	w.WriteHeader(http.StatusNoContent)
}
