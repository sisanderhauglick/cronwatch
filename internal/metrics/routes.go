package metrics

import (
	"net/http"
	"time"
)

// RegisterRoutes mounts all metrics HTTP handlers onto mux.
func RegisterRoutes(
	mux *http.ServeMux,
	reg *Registry,
	collector *Collector,
	tracker *AlertTracker,
	silence *SilenceManager,
	runLog *RunLog,
	budget *BudgetAnalyzer,
) {
	mux.Handle("/metrics", Handler(reg))
	mux.Handle("/metrics/prometheus", PrometheusHandler(reg))
	mux.Handle("/metrics/summary", SummaryHandler(NewAggregator(collector, time.Hour)))
	mux.Handle("/metrics/trend", TrendHandler(NewTrendAnalyzer(collector, time.Hour)))
	mux.Handle("/metrics/health", HealthHandler(reg, collector))
	mux.Handle("/metrics/uptime", UptimeHandler(NewUptimeTracker(collector, time.Hour*24)))
	mux.Handle("/metrics/sla", NewSLAHandler(NewSLAEvaluator(collector, nil, time.Hour*24)))
	mux.Handle("/metrics/latency", LatencyHandler(NewLatencyTracker(collector, time.Hour)))
	mux.Handle("/metrics/anomaly", AnomalyHandler(NewAnomalyDetector(collector, time.Hour, 0, 0)))
	mux.Handle("/metrics/forecast", nil) // placeholder
	mux.Handle("/metrics/alerts", alertTrackerHandler(tracker))
	mux.Handle("/metrics/digest", NewDigestHandler(NewDigestAnalyzer(collector, time.Hour*24)))
	mux.Handle("/metrics/replay", ReplayHandler(NewReplayAnalyzer(tracker)))
	mux.Handle("/metrics/silence", SilenceHandler(silence))
	mux.Handle("/metrics/runlog", RunLogHandler(runLog))
	mux.Handle("/metrics/dashboard", DashboardHandler(reg, collector))
	mux.Handle("/metrics/ratelimit", RateLimitHandler(NewAlertRateLimiter(time.Minute)))
	mux.Handle("/metrics/budget", BudgetHandler(budget))
}

func alertTrackerHandler(t *AlertTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events := t.Recent(50)
		if events == nil {
			events = []AlertEvent{}
		}
		writeJSON(w, events)
	}
}
