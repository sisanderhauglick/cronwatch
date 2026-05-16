package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// OncallHandler serves GET (list all shifts) and POST (add a shift).
type OncallHandler struct {
	manager *OncallManager
}

// NewOncallHandler returns an http.Handler backed by the given manager.
func NewOncallHandler(m *OncallManager) http.Handler {
	return &OncallHandler{manager: m}
}

func (h *OncallHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OncallHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	shifts := h.manager.All()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shifts); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}

func (h *OncallHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string `json:"name"`
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	start, err := time.Parse(time.RFC3339, body.Start)
	if err != nil {
		http.Error(w, "invalid start time", http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, body.End)
	if err != nil {
		http.Error(w, "invalid end time", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	h.manager.AddShift(OncallShift{Name: body.Name, Start: start, End: end})
	w.WriteHeader(http.StatusCreated)
}
