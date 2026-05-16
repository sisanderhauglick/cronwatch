package metrics

import (
	"encoding/json"
	"net/http"
)

// BackoffHandler returns an HTTP handler exposing current backoff state.
func BackoffHandler(b *BackoffManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			entries := b.All()
			if entries == nil {
				entries = []BackoffEntry{}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(entries)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, "missing job parameter", http.StatusBadRequest)
				return
			}
			b.Reset(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
