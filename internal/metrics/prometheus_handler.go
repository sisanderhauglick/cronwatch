package metrics

import (
	"net/http"
)

// PrometheusHandler returns an http.HandlerFunc that serves metrics
// in Prometheus text exposition format.
func PrometheusHandler(r *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if err := r.WritePrometheus(w); err != nil {
			http.Error(w, "failed to write metrics", http.StatusInternalServerError)
		}
	}
}
