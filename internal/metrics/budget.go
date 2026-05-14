package metrics

import (
	"sync"
	"time"
)

// ErrorBudget represents the remaining error budget for a job.
type ErrorBudget struct {
	Job           string  `json:"job"`
	SLOPercent    float64 `json:"slo_percent"`
	AllowedErrors int     `json:"allowed_errors"`
	ActualErrors  int     `json:"actual_errors"`
	Remaining     int     `json:"remaining"`
	Exhausted     bool    `json:"exhausted"`
}

// BudgetAnalyzer computes error budgets based on SLO targets and observed failures.
type BudgetAnalyzer struct {
	mu        sync.Mutex
	collector *Collector
	window    time.Duration
	sloByJob  map[string]float64
}

// NewBudgetAnalyzer creates a BudgetAnalyzer with the given collector, window, and per-job SLOs.
func NewBudgetAnalyzer(c *Collector, window time.Duration, sloByJob map[string]float64) *BudgetAnalyzer {
	return &BudgetAnalyzer{
		collector: c,
		window:    window,
		sloByJob:  sloByJob,
	}
}

// Analyze returns error budget status for all tracked jobs.
func (b *BudgetAnalyzer) Analyze(now time.Time) []ErrorBudget {
	b.mu.Lock()
	defer b.mu.Unlock()

	cutoff := now.Add(-b.window)
	snapshots := b.collector.All()

	// Aggregate per-job counts within the window.
	totalRuns := map[string]int{}
	failRuns := map[string]int{}
	for _, snap := range snapshots {
		if snap.Time.Before(cutoff) {
			continue
		}
		for job, stats := range snap.Jobs {
			totalRuns[job] += stats.Seen
			failRuns[job] += stats.Failed + stats.Missed
		}
	}

	var results []ErrorBudget
	for job, slo := range b.sloByJob {
		total := totalRuns[job]
		fails := failRuns[job]
		allowed := 0
		if total > 0 {
			allowed = int(float64(total) * (1.0 - slo/100.0))
		}
		remaining := allowed - fails
		results = append(results, ErrorBudget{
			Job:           job,
			SLOPercent:    slo,
			AllowedErrors: allowed,
			ActualErrors:  fails,
			Remaining:     remaining,
			Exhausted:     remaining < 0,
		})
	}
	return results
}
