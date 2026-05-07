package metrics

import (
	"math"
	"time"
)

// CorrelationResult holds the computed correlation between two jobs.
type CorrelationResult struct {
	JobA        string  `json:"job_a"`
	JobB        string  `json:"job_b"`
	Correlation float64 `json:"correlation"`
}

// CorrelationAnalyzer computes failure correlations between job pairs.
type CorrelationAnalyzer struct {
	collector *Collector
	window    time.Duration
}

// NewCorrelationAnalyzer returns a CorrelationAnalyzer using the given collector and window.
func NewCorrelationAnalyzer(c *Collector, window time.Duration) *CorrelationAnalyzer {
	return &CorrelationAnalyzer{collector: c, window: window}
}

// Analyze returns pairwise failure correlations for all jobs within the time window.
func (ca *CorrelationAnalyzer) Analyze(now time.Time) []CorrelationResult {
	cutoff := now.Add(-ca.window)
	snaps := ca.collector.All()

	// Build per-job binary failure series keyed by snapshot index.
	jobSeries := make(map[string][]float64)
	var ordered []Snapshot
	for _, s := range snaps {
		if s.Timestamp.After(cutoff) {
			ordered = append(ordered, s)
		}
	}
	if len(ordered) == 0 {
		return nil
	}

	for i, s := range ordered {
		for job, stats := range s.Jobs {
			if _, ok := jobSeries[job]; !ok {
				jobSeries[job] = make([]float64, len(ordered))
			}
			if stats.Failed > 0 || stats.Missed > 0 {
				jobSeries[job][i] = 1.0
			}
		}
	}

	jobs := make([]string, 0, len(jobSeries))
	for j := range jobSeries {
		jobs = append(jobs, j)
	}

	var results []CorrelationResult
	for i := 0; i < len(jobs); i++ {
		for j := i + 1; j < len(jobs); j++ {
			r := pearson(jobSeries[jobs[i]], jobSeries[jobs[j]])
			if math.IsNaN(r) {
				continue
			}
			results = append(results, CorrelationResult{
				JobA:        jobs[i],
				JobB:        jobs[j],
				Correlation: math.Round(r*1000) / 1000,
			})
		}
	}
	return results
}

func pearson(a, b []float64) float64 {
	n := float64(len(a))
	if n == 0 {
		return math.NaN()
	}
	var sumA, sumB, sumAB, sumA2, sumB2 float64
	for i := range a {
		sumA += a[i]
		sumB += b[i]
		sumAB += a[i] * b[i]
		sumA2 += a[i] * a[i]
		sumB2 += b[i] * b[i]
	}
	num := n*sumAB - sumA*sumB
	den := math.Sqrt((n*sumA2 - sumA*sumA) * (n*sumB2 - sumB*sumB))
	if den == 0 {
		return math.NaN()
	}
	return num / den
}
