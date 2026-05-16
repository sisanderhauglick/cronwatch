package metrics

import (
	"sort"
	"sync"
	"time"
)

// GroupByResult holds aggregated stats for a single group value.
type GroupByResult struct {
	Group       string  `json:"group"`
	JobCount    int     `json:"job_count"`
	TotalRuns   int     `json:"total_runs"`
	TotalFailed int     `json:"total_failed"`
	TotalMissed int     `json:"total_missed"`
	SuccessRate float64 `json:"success_rate"`
}

// GroupByAnalyzer aggregates snapshot metrics by a tag key.
type GroupByAnalyzer struct {
	mu        sync.Mutex
	collector *Collector
	tagIndex  *TagIndex
	window    time.Duration
}

// NewGroupByAnalyzer creates a GroupByAnalyzer using the provided collector and tag index.
func NewGroupByAnalyzer(c *Collector, ti *TagIndex, window time.Duration) *GroupByAnalyzer {
	return &GroupByAnalyzer{collector: c, tagIndex: ti, window: window}
}

// Summarize groups all jobs by the given tag key and aggregates their metrics
// across snapshots within the configured window.
func (g *GroupByAnalyzer) Summarize(tagKey string) []GroupByResult {
	g.mu.Lock()
	defer g.mu.Unlock()

	cutoff := time.Now().Add(-g.window)
	snaps := g.collector.All()

	type bucket struct {
		jobs    map[string]struct{}
		runs    int
		failed  int
		missed  int
	}

	groups := map[string]*bucket{}

	for _, snap := range snaps {
		if snap.CollectedAt.Before(cutoff) {
			continue
		}
		for job, stat := range snap.Jobs {
			tags := g.tagIndex.Lookup(job)
			val, ok := tags[tagKey]
			if !ok {
				val = "(untagged)"
			}
			b, exists := groups[val]
			if !exists {
				b = &bucket{jobs: map[string]struct{}{}}
				groups[val] = b
			}
			b.jobs[job] = struct{}{}
			b.runs += stat.SeenCount
			b.failed += stat.FailedCount
			b.missed += stat.MissedCount
		}
	}

	results := make([]GroupByResult, 0, len(groups))
	for grp, b := range groups {
		rate := 1.0
		if b.runs > 0 {
			rate = float64(b.runs-b.failed-b.missed) / float64(b.runs)
			if rate < 0 {
				rate = 0
			}
		}
		results = append(results, GroupByResult{
			Group:       grp,
			JobCount:    len(b.jobs),
			TotalRuns:   b.runs,
			TotalFailed: b.failed,
			TotalMissed: b.missed,
			SuccessRate: rate,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Group < results[j].Group
	})
	return results
}
