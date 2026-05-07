package metrics

import (
	"math"
	"time"
)

// SLATarget defines the expected uptime percentage for a job.
type SLATarget struct {
	JobName   string
	TargetPct float64 // e.g. 99.5 means 99.5%
}

// SLAResult holds the evaluation outcome for a single job.
type SLAResult struct {
	JobName    string
	TargetPct  float64
	ActualPct  float64
	Breaching  bool
	MarginPct  float64 // ActualPct - TargetPct (negative when breaching)
}

// SLAEvaluator computes SLA compliance using uptime data.
type SLAEvaluator struct {
	tracker *UptimeTracker
	window  time.Duration
}

// NewSLAEvaluator creates an evaluator backed by the given UptimeTracker.
func NewSLAEvaluator(tracker *UptimeTracker, window time.Duration) *SLAEvaluator {
	return &SLAEvaluator{tracker: tracker, window: window}
}

// Evaluate checks each target against actual uptime and returns results.
func (e *SLAEvaluator) Evaluate(targets []SLATarget) []SLAResult {
	now := time.Now()
	results := make([]SLAResult, 0, len(targets))
	for _, t := range targets {
		actual := e.tracker.Uptime(t.JobName, now, e.window)
		actualPct := math.Round(actual*10000) / 100 // two decimal places
		margin := math.Round((actualPct-t.TargetPct)*100) / 100
		results = append(results, SLAResult{
			JobName:   t.JobName,
			TargetPct: t.TargetPct,
			ActualPct: actualPct,
			Breaching: actualPct < t.TargetPct,
			MarginPct: margin,
		})
	}
	return results
}
