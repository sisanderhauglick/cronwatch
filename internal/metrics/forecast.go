package metrics

import (
	"math"
	"time"
)

// ForecastResult holds a predicted next-failure probability for a job.
type ForecastResult struct {
	JobName         string    `json:"job_name"`
	FailureProb     float64   `json:"failure_probability"`
	SampleWindow    int       `json:"sample_window"`
	ComputedAt      time.Time `json:"computed_at"`
}

// ForecastAnalyzer predicts failure likelihood based on recent snapshot history.
type ForecastAnalyzer struct {
	collector *Collector
	window    time.Duration
}

// NewForecastAnalyzer creates a ForecastAnalyzer using the given collector and
// look-back window.
func NewForecastAnalyzer(c *Collector, window time.Duration) *ForecastAnalyzer {
	return &ForecastAnalyzer{collector: c, window: window}
}

// Predict returns a ForecastResult for each known job. The failure probability
// is the exponentially-weighted ratio of failed runs within the window, giving
// more weight to recent snapshots.
func (f *ForecastAnalyzer) Predict(now time.Time) []ForecastResult {
	cutoff := now.Add(-f.window)
	all := f.collector.All()

	// Gather job names
	jobSet := map[string]struct{}{}
	for _, snap := range all {
		for name := range snap.Jobs {
			jobSet[name] = struct{}{}
		}
	}

	results := make([]ForecastResult, 0, len(jobSet))
	for job := range jobSet {
		var weightedFail, weightedTotal float64
		samples := 0
		for i, snap := range all {
			if snap.CollectedAt.Before(cutoff) {
				continue
			}
			s, ok := snap.Jobs[job]
			if !ok {
				continue
			}
			// Weight increases linearly with recency (index position).
			w := math.Log1p(float64(i + 1))
			weightedFail += w * float64(s.Failed)
			weightedTotal += w * float64(s.Seen+s.Failed)
			samples++
		}
		prob := 0.0
		if weightedTotal > 0 {
			prob = math.Round((weightedFail/weightedTotal)*1000) / 1000
		}
		results = append(results, ForecastResult{
			JobName:      job,
			FailureProb:  prob,
			SampleWindow: samples,
			ComputedAt:   now,
		})
	}
	return results
}
