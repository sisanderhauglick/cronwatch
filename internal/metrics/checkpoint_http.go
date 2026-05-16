package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// CheckpointHandler serves checkpoint data over HTTP.
type CheckpointHandler struct {
	store *CheckpointStore
}

// NewCheckpointHandler returns an HTTP handler backed by store.
func NewCheckpointHandler(store *CheckpointStore) http.Handler {
	return &CheckpointHandler{store: store}
}

func (h *CheckpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, job)
	case http.MethodPost:
		h.handlePost(w, r, job)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CheckpointHandler) handleGet(w http.ResponseWriter, job string) {
	w.Header().Set("Content-Type", "application/json")
	if job != "" {
		e, ok := h.store.Get(job)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(e)
		return
	}
	json.NewEncoder(w).Encode(h.store.All())
}

func (h *CheckpointHandler) handlePost(w http.ResponseWriter, r *http.Request, job string) {
	if job == "" {
		http.Error(w, "job query param required", http.StatusBadRequest)
		return
	}
	var body struct {
		LastOK time.Time `json:"last_ok"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LastOK.IsZero() {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	h.store.Record(job, body.LastOK)
	w.WriteHeader(http.StatusNoContent)
}
