package metrics

import (
	"encoding/json"
	"net/http"
)

// EventLogHandler returns an HTTP handler that exposes the event log as JSON.
// Query params:
//   - job=<name>      filter by job name
//   - severity=<lvl>  filter by severity (info|warn|error)
func EventLogHandler(log *EventLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		job := q.Get("job")
		sev := EventSeverity(q.Get("severity"))

		var entries []EventEntry
		switch {
		case job != "" && sev != "":
			all := log.FilterByJob(job)
			for _, e := range all {
				if e.Severity == sev {
					entries = append(entries, e)
				}
			}
		case job != "":
			entries = log.FilterByJob(job)
		case sev != "":
			entries = log.FilterBySeverity(sev)
		default:
			entries = log.All()
		}

		if entries == nil {
			entries = []EventEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}
}
