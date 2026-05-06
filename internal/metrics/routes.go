package metrics

import "net/http"

// RegisterRoutes attaches all metrics-related HTTP handlers to the given
// ServeMux under the /metrics/* prefix.
//
//   GET /metrics/json        – per-job counters snapshot
//   GET /metrics/prometheus  – Prometheus text exposition
//   GET /metrics/summary     – windowed aggregation summary
//   GET /metrics/dashboard   – high-level health dashboard
func RegisterRoutes(mux *http.ServeMux, r *Registry, c *Collector, a *Aggregator) {
	mux.Handle("/metrics/json", Handler(r))
	mux.Handle("/metrics/prometheus", PrometheusHandler(r))
	mux.Handle("/metrics/summary", SummaryHandler(a))
	mux.Handle("/metrics/dashboard", DashboardHandler(r))
}
