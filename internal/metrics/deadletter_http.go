package metrics

import (
	"encoding/json"
	"net/http"
)

// DeadLetterHandler serves HTTP endpoints for inspecting and managing the dead-letter queue.
type DeadLetterHandler struct {
	queue *DeadLetterQueue
}

// NewDeadLetterHandler creates an HTTP handler backed by the given queue.
func NewDeadLetterHandler(q *DeadLetterQueue) *DeadLetterHandler {
	return &DeadLetterHandler{queue: q}
}

// ServeHTTP handles GET (list) and DELETE (remove by job+reason query params).
func (h *DeadLetterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		entries := h.queue.All()
		if entries == nil {
			entries = []DeadLetterEntry{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)

	case http.MethodDelete:
		job := r.URL.Query().Get("job")
		reason := r.URL.Query().Get("reason")
		if job == "" || reason == "" {
			http.Error(w, "job and reason query params required", http.StatusBadRequest)
			return
		}
		if h.queue.Remove(job, reason) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, "entry not found", http.StatusNotFound)
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
