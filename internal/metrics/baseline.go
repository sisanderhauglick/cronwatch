package metrics

import (
	"math"
	"time"
)

// BaselineSummary holds the computed baseline stats for a single job.
type BaselineSummary struct {
	Job          string  `json:"job"`
	AvgRunsPerHr float64 `json:"avg_runs_per_hour"`
	StdDev       float64 `json:"std_dev"`
	SampleCount  int     `json:"sample_count"`
}

// BaselineAnalyzer computes per-job run-rate baselines from historical snapshots.
type BaselineAnalyzer struct {
	collector *Collector
	window    time.Duration
}

// NewBaselineAnalyzer returns a BaselineAnalyzer using the given collector and
// lookback window.
func NewBaselineAnalyzer(c *Collector, window time.Duration) *BaselineAnalyzer {
	return &BaselineAnalyzer{collector: c, window: window}
}

// Analyze returns a BaselineSummary for every job that has snapshots within
// the lookback window.
func (b *BaselineAnalyzer) Analyze(now time.Time) []BaselineSummary {
	cutoff := now.Add(-b.window)

	// Bucket run counts by job per hour slot.
	type slot struct {
		hour  int64 // unix hour
		count float64
	}
	hourly := map[string]map[int64]float64{}

	for _, snap := range b.collector.All() {
		if snap.Timestamp.Before(cutoff) {
			continue
		}
		for job, m := range snap.Jobs {
			if _, ok := hourly[job]; !ok {
				hourly[job] = map[int64]float64{}
			}
			h := snap.Timestamp.Unix() / 3600
			hourly[job][h] += float64(m.SeenCount)
		}
	}

	results := make([]BaselineSummary, 0, len(hourly))
	for job, slots := range hourly {
		if len(slots) == 0 {
			continue
		}
		values := make([]float64, 0, len(slots))
		sum := 0.0
		for _, v := range slots {
			values = append(values, v)
			sum += v
		}
		avg := sum / float64(len(values))
		variance := 0.0
		for _, v := range values {
			d := v - avg
			variance += d * d
		}
		if len(values) > 1 {
			variance /= float64(len(values) - 1)
		}
		results = append(results, BaselineSummary{
			Job:          job,
			AvgRunsPerHr: avg,
			StdDev:       math.Sqrt(variance),
			SampleCount:  len(values),
		})
	}
	return results
}
