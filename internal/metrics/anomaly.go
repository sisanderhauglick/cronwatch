package metrics

import (
	"math"
	"time"
)

// AnomalyReport describes an unusual pattern detected for a job.
type AnomalyReport struct {
	Job       string
	Kind      string  // "latency_spike", "failure_burst", "silence"
	Score     float64 // higher = more anomalous
	DetectedAt time.Time
}

// AnomalyDetector inspects collector snapshots for unusual patterns.
type AnomalyDetector struct {
	collector  *Collector
	window     time.Duration
	latency    *LatencyTracker
	spikeRatio float64 // e.g. 2.0 means 2× median triggers spike
}

// NewAnomalyDetector creates a detector using the given collector.
// spikeRatio is the multiplier above the median p50 that triggers a latency spike.
func NewAnomalyDetector(c *Collector, lt *LatencyTracker, window time.Duration, spikeRatio float64) *AnomalyDetector {
	if spikeRatio <= 0 {
		spikeRatio = 2.0
	}
	return &AnomalyDetector{
		collector:  c,
		window:     window,
		latency:    lt,
		spikeRatio: spikeRatio,
	}
}

// Detect returns anomaly reports for all jobs within the configured window.
func (a *AnomalyDetector) Detect(now time.Time) []AnomalyReport {
	snapshots := a.collector.All()
	var reports []AnomalyReport

	for job, snaps := range snapshots {
		var recent []Snapshot
		for _, s := range snaps {
			if now.Sub(s.CollectedAt) <= a.window {
				recent = append(recent, s)
			}
		}
		if len(recent) == 0 {
			continue
		}

		// Failure burst: majority of recent runs failed
		failed := 0
		for _, s := range recent {
			if s.Failed > 0 {
				failed++
			}
		}
		if len(recent) > 0 && float64(failed)/float64(len(recent)) >= 0.5 {
			reports = append(reports, AnomalyReport{
				Job:        job,
				Kind:       "failure_burst",
				Score:      float64(failed) / float64(len(recent)),
				DetectedAt: now,
			})
		}

		// Latency spike: current p50 > spikeRatio * historical median
		p := a.latency.Percentiles(job, now)
		if p.P50 > 0 {
			historical := a.historicalMedian(job, now)
			if historical > 0 && p.P50 > a.spikeRatio*historical {
				score := p.P50 / historical
				reports = append(reports, AnomalyReport{
					Job:        job,
					Kind:       "latency_spike",
					Score:      math.Round(score*100) / 100,
					DetectedAt: now,
				})
			}
		}
	}
	return reports
}

// historicalMedian returns the average p50 across all snapshots older than the window.
func (a *AnomalyDetector) historicalMedian(job string, now time.Time) float64 {
	snaps := a.collector.All()[job]
	var total float64
	var count int
	for _, s := range snaps {
		if now.Sub(s.CollectedAt) > a.window {
			total += s.AvgLatencyMs
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}
