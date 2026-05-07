package metrics

import (
	"sync"
	"time"
)

// UptimeTracker tracks when each job was first seen and computes uptime percentage
// based on expected vs actual successful runs within a window.
type UptimeTracker struct {
	mu       sync.Mutex
	firstSeen map[string]time.Time
	collector *Collector
}

// NewUptimeTracker creates a new UptimeTracker backed by the given Collector.
func NewUptimeTracker(c *Collector) *UptimeTracker {
	return &UptimeTracker{
		firstSeen: make(map[string]time.Time),
		collector: c,
	}
}

// RecordSeen marks the first time a job was observed.
func (u *UptimeTracker) RecordSeen(job string, at time.Time) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if _, ok := u.firstSeen[job]; !ok {
		u.firstSeen[job] = at
	}
}

// UptimeResult holds uptime information for a single job.
type UptimeResult struct {
	Job      string
	Since    time.Time
	UptimePct float64
}

// Compute returns uptime results for all known jobs within the given window.
// Uptime is defined as seen_runs / (seen_runs + missed_runs) * 100.
func (u *UptimeTracker) Compute(window time.Duration) []UptimeResult {
	u.mu.Lock()
	defer u.mu.Unlock()

	snapshots := u.collector.All()
	now := time.Now()
	cutoff := now.Add(-window)

	type counts struct {
		seen   int
		missed int
	}
	jobCounts := make(map[string]*counts)

	for _, snap := range snapshots {
		if snap.Timestamp.Before(cutoff) {
			continue
		}
		for job, stat := range snap.Stats {
			if _, ok := jobCounts[job]; !ok {
				jobCounts[job] = &counts{}
			}
			jobCounts[job].seen += stat.SeenCount
			jobCounts[job].missed += stat.MissedCount
		}
	}

	var results []UptimeResult
	for job, since := range u.firstSeen {
		c := jobCounts[job]
		var pct float64
		if c != nil {
			total := c.seen + c.missed
			if total > 0 {
				pct = float64(c.seen) / float64(total) * 100.0
			} else {
				pct = 100.0
			}
		} else {
			pct = 100.0
		}
		results = append(results, UptimeResult{
			Job:       job,
			Since:     since,
			UptimePct: pct,
		})
	}
	return results
}
