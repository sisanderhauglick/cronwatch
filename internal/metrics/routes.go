package metrics

import (
	"net/http"
	"time"
)

// RegisterRoutes mounts all metrics sub-handlers onto mux.
func RegisterRoutes(
	mux *http.ServeMux,
	reg *Registry,
	collector *Collector,
	tracker *AlertTracker,
	limiter *AlertRateLimiter,
) {
	health := NewHealthEvaluator(collector, reg, 2*time.Hour)
	agg := NewAggregator(collector, time.Hour)
	trend := NewTrendAnalyzer(collector, time.Hour)
	anomaly := NewAnomalyDetector(collector, time.Hour)
	forecast := NewForecastAnalyzer(collector, time.Hour)
	heatmap := NewHeatmapAnalyzer(collector, 24*time.Hour)
	corr := NewCorrelationAnalyzer(collector, time.Hour)
	baseline := NewBaselineAnalyzer(collector, 7*24*time.Hour)
	uptime := NewUptimeTracker(collector, 24*time.Hour)
	sla := NewSLAEvaluator(collector, 24*time.Hour, 0.99)
	latency := NewLatencyTracker(collector, time.Hour)
	replay := NewReplayAnalyzer(tracker, time.Hour)
	digest := NewDigestAnalyzer(collector, health, time.Hour)

	mux.Handle("/metrics/jobs", Handler(reg))
	mux.Handle("/metrics/prometheus", PrometheusHandler(reg))
	mux.Handle("/metrics/summary", SummaryHandler(agg))
	mux.Handle("/metrics/dashboard", DashboardHandler(collector, reg))
	mux.Handle("/metrics/trend", TrendHandler(trend))
	mux.Handle("/metrics/anomaly", AnomalyHandler(anomaly))
	mux.Handle("/metrics/forecast", &forecastHTTPHandler{analyzer: forecast})
	mux.Handle("/metrics/heatmap", &heatmapHTTPHandler{analyzer: heatmap})
	mux.Handle("/metrics/correlation", &correlationHTTPHandler{analyzer: corr})
	mux.Handle("/metrics/baseline", &baselineHTTPHandler{analyzer: baseline})
	mux.Handle("/metrics/health", HealthHandler(health))
	mux.Handle("/metrics/uptime", UptimeHandler(uptime))
	mux.Handle("/metrics/sla", NewSLAHandler(sla))
	mux.Handle("/metrics/latency", LatencyHandler(latency))
	mux.Handle("/metrics/alerts", alertTrackerHandler(tracker))
	mux.Handle("/metrics/ratelimit", RateLimitHandler(limiter))
	mux.Handle("/metrics/replay", ReplayHandler(replay))
	mux.Handle("/metrics/digest", NewDigestHandler(digest))
}

func alertTrackerHandler(t *AlertTracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		events := t.Recent(50)
		writeJSON(w, events)
	})
}
