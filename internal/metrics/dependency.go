package metrics

import (
	"math"
	"sort"
	"time"
)

// DependencyEdge represents a detected dependency between two jobs.
type DependencyEdge struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Correlation float64 `json:"correlation"`
}

// DependencyAnalyzer infers likely job dependencies from failure co-occurrence.
type DependencyAnalyzer struct {
	collector *Collector
	window    time.Duration
	threshold float64
}

// NewDependencyAnalyzer creates a DependencyAnalyzer.
// threshold is the minimum Pearson correlation to report an edge (0..1).
func NewDependencyAnalyzer(c *Collector, window time.Duration, threshold float64) *DependencyAnalyzer {
	if threshold <= 0 {
		threshold = 0.6
	}
	return &DependencyAnalyzer{collector: c, window: window, threshold: threshold}
}

// Analyze returns job pairs whose failure timeseries are correlated above the threshold.
func (d *DependencyAnalyzer) Analyze(now time.Time) []DependencyEdge {
	cutoff := now.Add(-d.window)
	snaps := d.collector.All()

	// collect per-job failure vectors keyed by snapshot time bucket (minute)
	type bucket = map[int64]float64
	jobVecs := map[string]bucket{}

	for _, s := range snaps {
		if s.Time.Before(cutoff) {
			continue
		}
		tb := s.Time.Truncate(time.Minute).Unix()
		for job, m := range s.Jobs {
			if jobVecs[job] == nil {
				jobVecs[job] = bucket{}
			}
			if m.Failed > 0 || m.Missed > 0 {
				jobVecs[job][tb] += float64(m.Failed + m.Missed)
			} else {
				if _, ok := jobVecs[job][tb]; !ok {
					jobVecs[job][tb] = 0
				}
			}
		}
	}

	jobs := make([]string, 0, len(jobVecs))
	for j := range jobVecs {
		jobs = append(jobs, j)
	}
	sort.Strings(jobs)

	var edges []DependencyEdge
	for i := 0; i < len(jobs); i++ {
		for j := i + 1; j < len(jobs); j++ {
			r := pearsonBuckets(jobVecs[jobs[i]], jobVecs[jobs[j]])
			if !math.IsNaN(r) && r >= d.threshold {
				edges = append(edges, DependencyEdge{
					From:        jobs[i],
					To:          jobs[j],
					Correlation: math.Round(r*1000) / 1000,
				})
			}
		}
	}
	return edges
}

func pearsonBuckets(a, b map[int64]float64) float64 {
	// union of time buckets
	keys := map[int64]struct{}{}
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	if len(keys) < 2 {
		return math.NaN()
	}
	var xs, ys []float64
	for k := range keys {
		xs = append(xs, a[k])
		ys = append(ys, b[k])
	}
	return pearson(xs, ys)
}
