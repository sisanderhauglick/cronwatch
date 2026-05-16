package metrics

import (
	"encoding/json"
	"net/http"
)

// CircuitHandler serves the current state of all circuit breakers as JSON.
// GET /metrics/circuits  — returns []CircuitEntry
// DELETE /metrics/circuits/{job} — resets the circuit for a job
func CircuitHandler(cb *CircuitBreaker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			entries := cb.All()
			if entries == nil {
				entries = []CircuitEntry{}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(entries)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, "missing job query param", http.StatusBadRequest)
				return
			}
			cb.RecordSuccess(job) // reset by recording a synthetic success
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
