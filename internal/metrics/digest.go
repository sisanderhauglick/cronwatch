package metrics

import (
	"fmt"
	"sort"
	"time"
)

// DigestEntry summarises activity for a single job over a window.
type DigestEntry struct {
	JobName     string        `json:"job_name"`
	TotalRuns   int           `json:"total_runs"`
	Failures    int           `json:"failures"`
	Missed      int           `json:"missed"`
	SuccessRate float64       `json:"success_rate"`
	AvgLatency  time.Duration `json:"avg_latency_ms"`
	Health      string        `json:"health"`
}

// DigestReport is the full report produced by DigestAnalyzer.
type DigestReport struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Window      time.Duration  `json:"window"`
	Entries     []DigestEntry  `json:"entries"`
}

// DigestAnalyzer builds periodic digest reports from collected snapshots.
type DigestAnalyzer struct {
	collector *Collector
	health    *HealthEvaluator
	window    time.Duration
	now       func() time.Time
}

// NewDigestAnalyzer returns a DigestAnalyzer using the provided collector.
func NewDigestAnalyzer(c *Collector, h *HealthEvaluator, window time.Duration) *DigestAnalyzer {
	return &DigestAnalyzer{
		collector: c,
		health:    h,
		window:    window,
		now:       time.Now,
	}
}

// Summarize builds a DigestReport covering the configured window.
func (d *DigestAnalyzer) Summarize() DigestReport {
	cutoff := d.now().Add(-d.window)
	all := d.collector.All()

	type agg struct {
		runs    int
		fail    int
		missed  int
		latency int64
	}
	aggs := map[string]*agg{}

	for _, snap := range all {
		if snap.Timestamp.Before(cutoff) {
			continue
		}
		for job, s := range snap.Jobs {
			a := aggs[job]
			if a == nil {
				a = &agg{}
				aggs[job] = a
			}
			a.runs += s.Seen
			a.fail += s.Failed
			a.missed += s.Missed
			a.latency += s.TotalLatencyMs
		}
	}

	entries := make([]DigestEntry, 0, len(aggs))
	for job, a := range aggs {
		sr := 0.0
		if a.runs > 0 {
			sr = float64(a.runs-a.fail-a.missed) / float64(a.runs)
		}
		avg := time.Duration(0)
		if a.runs > 0 {
			avg = time.Duration(a.latency/int64(a.runs)) * time.Millisecond
		}
		h := d.health.Evaluate(job)
		entries = append(entries, DigestEntry{
			JobName:     job,
			TotalRuns:   a.runs,
			Failures:    a.fail,
			Missed:      a.missed,
			SuccessRate: sr,
			AvgLatency:  avg,
			Health:      fmt.Sprintf("%s", h.Status),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].JobName < entries[j].JobName
	})
	return DigestReport{
		GeneratedAt: d.now(),
		Window:      d.window,
		Entries:     entries,
	}
}
