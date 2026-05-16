package metrics

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ThrottleHandler exposes throttle state and allows resets via HTTP.
//
//	GET  /metrics/throttle        → list all entries
//	GET  /metrics/throttle/{job}  → single entry
//	DELETE /metrics/throttle/{job} → reset entry
type ThrottleHandler struct {
	tm *ThrottleManager
}

// NewThrottleHandler creates a new ThrottleHandler.
func NewThrottleHandler(tm *ThrottleManager) *ThrottleHandler {
	return &ThrottleHandler{tm: tm}
}

func (h *ThrottleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Strip leading slash and split to get optional job segment.
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	jobName := ""
	if len(parts) == 2 {
		jobName = parts[1]
	}

	switch r.Method {
	case http.MethodDelete:
		if jobName == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}
		h.tm.Reset(jobName)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		if jobName != "" {
			e, ok := h.tm.Stats(jobName)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(e)
			return
		}
		entries := h.tm.All()
		if entries == nil {
			entries = []ThrottleEntry{}
		}
		json.NewEncoder(w).Encode(entries)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
