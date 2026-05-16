package metrics

import (
	"encoding/json"
	"net/http"
)

// PaginatedRunLogHandler returns a paginated view of the RunLog.
// Query params: job (optional filter), page, page_size.
func PaginatedRunLogHandler(rl *RunLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		req := ParsePageRequest(r)

		var entries []RunEntry
		if job != "" {
			entries = rl.Query(job)
		} else {
			entries = rl.All()
		}

		page := Paginate(entries, req)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page) //nolint:errcheck
	}
}
