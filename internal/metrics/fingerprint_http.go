package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// FingerprintHandler returns an HTTP handler for querying fingerprint records.
func FingerprintHandler(store *FingerprintStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			all := store.All(time.Now())
			if job != "" {
				filtered := all[:0]
				for _, rec := range all {
					if rec.Job == job {
						filtered = append(filtered, rec)
					}
				}
				all = filtered
			}
			if all == nil {
				all = []FingerprintRecord{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(all)

		case http.MethodDelete:
			hash := r.URL.Query().Get("hash")
			if hash == "" {
				http.Error(w, "missing hash", http.StatusBadRequest)
				return
			}
			store.mu.Lock()
			delete(store.entries, hash)
			store.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
