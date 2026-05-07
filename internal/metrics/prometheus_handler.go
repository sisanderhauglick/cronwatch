package metrics

import (
	"net/http"
	"net/http/httptest"
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

// PrometheusSnapshot returns the current metrics as a string in Prometheus
// text exposition format. It is intended for testing and debugging.
func PrometheusSnapshot(r *Registry) (string, error) {
	rec := httptest.NewRecorder()
	PrometheusHandler(r)(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		return "", fmt.Errorf("metrics handler returned status %d", rec.Code)
	}
	return rec.Body.String(), nil
}
