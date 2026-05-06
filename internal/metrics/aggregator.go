package metrics

import "time"

// JobSummary holds aggregated statistics for a single job over a time window.
type JobSummary struct {
	JobName     string
	Seen        int64
	Missed      int64
	Failed      int64
	SuccessRate float64
	WindowStart time.Time
	WindowEnd   time.Time
}

// Aggregator computes summary statistics from a slice of snapshots.
type Aggregator struct {
	collector *Collector
}

// NewAggregator returns an Aggregator backed by the given Collector.
func NewAggregator(c *Collector) *Aggregator {
	return &Aggregator{collector: c}
}

// Summarize returns aggregated JobSummary values for all jobs within
// the half-open interval [from, to).
func (a *Aggregator) Summarize(from, to time.Time) []JobSummary {
	snapshots := a.collector.All()

	type acc struct {
		seen   int64
		missed int64
		failed int64
	}
	totals := make(map[string]*acc)

	for _, snap := range snapshots {
		if snap.Timestamp.Before(from) || !snap.Timestamp.Before(to) {
			continue
		}
		for name, stat := range snap.Jobs {
			if _, ok := totals[name]; !ok {
				totals[name] = &acc{}
			}
			totals[name].seen += stat.Seen
			totals[name].missed += stat.Missed
			totals[name].failed += stat.Failed
		}
	}

	summaries := make([]JobSummary, 0, len(totals))
	for name, t := range totals {
		rate := 0.0
		if total := t.seen + t.missed + t.failed; total > 0 {
			rate = float64(t.seen) / float64(total)
		}
		summaries = append(summaries, JobSummary{
			JobName:     name,
			Seen:        t.seen,
			Missed:      t.missed,
			Failed:      t.failed,
			SuccessRate: rate,
			WindowStart: from,
			WindowEnd:   to,
		})
	}
	return summaries
}
