package metrics

import "net/http"

// RegisterRoutes mounts all metrics-related HTTP handlers onto mux.
func RegisterRoutes(mux *http.ServeMux, reg *Registry, col *Collector, agg *Aggregator, trend *TrendAnalyzer, tracker *AlertTracker) {
	mux.Handle("/metrics", Handler(reg))
	mux.Handle("/metrics/prometheus", PrometheusHandler(reg))
	mux.Handle("/metrics/summary", SummaryHandler(agg))
	mux.Handle("/metrics/dashboard", DashboardHandler(reg, agg))
	mux.Handle("/metrics/trend", TrendHandler(trend))
	mux.Handle("/metrics/alerts", alertTrackerHandler(tracker))
	mux.Handle("/metrics/health", HealthHandler(reg, NewHealthEvaluator()))
}

// alertTrackerHandler exposes recent alert events as JSON.
func alertTrackerHandler(tracker *AlertTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events := tracker.Recent(50)
		writeJSON(w, events)
	}
}
