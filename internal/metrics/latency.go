package metrics

import (
	"sync"
	"time"
)

// LatencyRecord holds timing information for a single job execution.
type LatencyRecord struct {
	Job       string
	Duration  time.Duration
	RecordedAt time.Time
}

// LatencyTracker tracks execution durations for cron jobs.
type LatencyTracker struct {
	mu      sync.RWMutex
	records map[string][]LatencyRecord
	window  time.Duration
}

// NewLatencyTracker creates a LatencyTracker that retains records within window.
func NewLatencyTracker(window time.Duration) *LatencyTracker {
	return &LatencyTracker{
		records: make(map[string][]LatencyRecord),
		window:  window,
	}
}

// Record stores a duration sample for the given job.
func (t *LatencyTracker) Record(job string, d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.records[job] = append(t.records[job], LatencyRecord{
		Job:        job,
		Duration:   d,
		RecordedAt: now,
	})
	t.prune(job, now)
}

// Stats returns p50, p95, and p99 latencies for a job over the retention window.
// Returns zero values if no data exists.
func (t *LatencyTracker) Stats(job string) (p50, p95, p99 time.Duration) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	recs := t.records[job]
	if len(recs) == 0 {
		return 0, 0, 0
	}
	durations := make([]time.Duration, len(recs))
	for i, r := range recs {
		durations[i] = r.Duration
	}
	sortDurations(durations)
	return percentile(durations, 50), percentile(durations, 95), percentile(durations, 99)
}

// Jobs returns all job names that have latency data.
func (t *LatencyTracker) Jobs() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]string, 0, len(t.records))
	for k := range t.records {
		out = append(out, k)
	}
	return out
}

func (t *LatencyTracker) prune(job string, now time.Time) {
	cutoff := now.Add(-t.window)
	recs := t.records[job]
	i := 0
	for i < len(recs) && recs[i].RecordedAt.Before(cutoff) {
		i++
	}
	t.records[job] = recs[i:]
}

func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j] < d[j-1]; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
