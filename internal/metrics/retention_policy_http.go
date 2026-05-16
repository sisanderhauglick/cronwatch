package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// RetentionPolicyHandler serves and updates the active retention policy over HTTP.
type RetentionPolicyHandler struct {
	policy *RetentionPolicy
}

// NewRetentionPolicyHandler creates a handler backed by the given policy.
func NewRetentionPolicyHandler(p *RetentionPolicy) *RetentionPolicyHandler {
	return &RetentionPolicyHandler{policy: p}
}

func (h *RetentionPolicyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w)
	case http.MethodPut:
		h.handlePut(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type retentionPolicyResponse struct {
	MaxAge      string `json:"max_age"`
	MaxSnapshots int   `json:"max_snapshots"`
}

func (h *RetentionPolicyHandler) handleGet(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	resp := retentionPolicyResponse{
		MaxAge:       h.policy.MaxAge.String(),
		MaxSnapshots: h.policy.MaxSnapshots,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RetentionPolicyHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if v := q.Get("max_age"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			http.Error(w, "invalid max_age: "+err.Error(), http.StatusBadRequest)
			return
		}
		h.policy.MaxAge = d
	}

	if v := q.Get("max_snapshots"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			http.Error(w, "invalid max_snapshots", http.StatusBadRequest)
			return
		}
		h.policy.MaxSnapshots = n
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h.handleGet(w)
}
