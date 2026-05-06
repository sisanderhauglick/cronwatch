package metrics

import (
	"math"
	"time"
)

// TrendDirection indicates whether a metric is improving, degrading, or stable.
type TrendDirection string

const (
	TrendUp     TrendDirection = "up"
	TrendDown   TrendDirection = "down"
	TrendStable TrendDirection = "stable"
)

// JobTrend holds a simple trend analysis for a single job.
type JobTrend struct {
	JobName        string         `json:"job_name"`
	Direction      TrendDirection `json:"direction"`
	AvgSuccessRate float64        `json:"avg_success_rate"`
	Delta          float64        `json:"delta"` // latest window rate minus previous window rate
}

// TrendAnalyzer computes success-rate trends across two time windows.
type TrendAnalyzer struct {
	agg       *Aggregator
	window    time.Duration
	threshold float64 // minimum absolute delta to consider non-stable
}

// NewTrendAnalyzer creates a TrendAnalyzer using the given aggregator.
// window is the size of each comparison window; threshold is the minimum
// delta (0–1) required to report a directional change.
func NewTrendAnalyzer(agg *Aggregator, window time.Duration, threshold float64) *TrendAnalyzer {
	return &TrendAnalyzer{agg: agg, window: window, threshold: threshold}
}

// Analyze returns a JobTrend for every job visible in the collector.
// It compares the most recent window against the preceding window.
func (t *TrendAnalyzer) Analyze(now time.Time) []JobTrend {
	current := t.agg.Summarize(now, t.window)
	previous := t.agg.Summarize(now.Add(-t.window), t.window)

	prevMap := make(map[string]float64, len(previous))
	for _, s := range previous {
		prevMap[s.JobName] = s.SuccessRate
	}

	trends := make([]JobTrend, 0, len(current))
	for _, s := range current {
		delta := s.SuccessRate - prevMap[s.JobName]
		dir := TrendStable
		if math.Abs(delta) >= t.threshold {
			if delta > 0 {
				dir = TrendUp
			} else {
				dir = TrendDown
			}
		}
		trends = append(trends, JobTrend{
			JobName:        s.JobName,
			Direction:      dir,
			AvgSuccessRate: s.SuccessRate,
			Delta:          delta,
		})
	}
	return trends
}
