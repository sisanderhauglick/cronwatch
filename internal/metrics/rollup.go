package metrics

import (
	"sync"
	"time"
)

// RollupBucket holds aggregated stats for a fixed time window.
type RollupBucket struct {
	Job     string    `json:"job"`
	Window  string    `json:"window"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	Runs    int       `json:"runs"`
	Failed  int       `json:"failed"`
	Missed  int       `json:"missed"`
	Success float64   `json:"success_rate"`
}

// RollupAnalyzer computes hourly and daily rollups from a Collector.
type RollupAnalyzer struct {
	mu        sync.Mutex
	collector *Collector
}

// NewRollupAnalyzer returns a RollupAnalyzer backed by the given Collector.
func NewRollupAnalyzer(c *Collector) *RollupAnalyzer {
	return &RollupAnalyzer{collector: c}
}

// Rollup aggregates snapshots within [from, to) into per-job buckets.
func (r *RollupAnalyzer) Rollup(from, to time.Time, label string) []RollupBucket {
	r.mu.Lock()
	defer r.mu.Unlock()

	type acc struct {
		runs, failed, missed int
	}
	jobs := map[string]*acc{}

	for _, snap := range r.collector.All() {
		if snap.Timestamp.Before(from) || !snap.Timestamp.Before(to) {
			continue
		}
		a, ok := jobs[snap.Job]
		if !ok {
			a = &acc{}
			jobs[snap.Job] = a
		}
		a.runs += snap.Seen
		a.failed += snap.Failed
		a.missed += snap.Missed
	}

	buckets := make([]RollupBucket, 0, len(jobs))
	for job, a := range jobs {
		rate := 0.0
		if a.runs > 0 {
			rate = float64(a.runs-a.failed-a.missed) / float64(a.runs)
		}
		buckets = append(buckets, RollupBucket{
			Job:     job,
			Window:  label,
			From:    from,
			To:      to,
			Runs:    a.runs,
			Failed:  a.failed,
			Missed:  a.missed,
			Success: rate,
		})
	}
	return buckets
}

// HourlyRollup returns rollup buckets for the past n hours.
func (r *RollupAnalyzer) HourlyRollup(n int, now time.Time) []RollupBucket {
	var out []RollupBucket
	for i := n - 1; i >= 0; i-- {
		to := now.Add(-time.Duration(i) * time.Hour).Truncate(time.Hour).Add(time.Hour)
		from := to.Add(-time.Hour)
		out = append(out, r.Rollup(from, to, "1h")...)
	}
	return out
}
