package metrics

import (
	"sort"
	"time"
)

// ReplayEntry represents a single historical alert event for replay analysis.
type ReplayEntry struct {
	JobName   string    `json:"job_name"`
	Status    string    `json:"status"`
	OccuredAt time.Time `json:"occured_at"`
	Message   string    `json:"message"`
}

// ReplayAnalyzer replays alert history to reconstruct job timelines.
type ReplayAnalyzer struct {
	collector *Collector
	tracker   *AlertTracker
}

// NewReplayAnalyzer creates a ReplayAnalyzer backed by the given collector and tracker.
func NewReplayAnalyzer(c *Collector, t *AlertTracker) *ReplayAnalyzer {
	return &ReplayAnalyzer{collector: c, tracker: t}
}

// Replay returns alert entries for a given job within [from, to], sorted by time.
func (r *ReplayAnalyzer) Replay(jobName string, from, to time.Time) []ReplayEntry {
	events := r.tracker.Recent(256)
	var entries []ReplayEntry
	for _, ev := range events {
		if ev.JobName != jobName {
			continue
		}
		if ev.FiredAt.Before(from) || ev.FiredAt.After(to) {
			continue
		}
		entries = append(entries, ReplayEntry{
			JobName:   ev.JobName,
			Status:    ev.Kind,
			OccuredAt: ev.FiredAt,
			Message:   ev.Message,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].OccuredAt.Before(entries[j].OccuredAt)
	})
	return entries
}

// ReplayAll returns entries for all jobs within [from, to], sorted by time.
func (r *ReplayAnalyzer) ReplayAll(from, to time.Time) []ReplayEntry {
	events := r.tracker.Recent(256)
	var entries []ReplayEntry
	for _, ev := range events {
		if ev.FiredAt.Before(from) || ev.FiredAt.After(to) {
			continue
		}
		entries = append(entries, ReplayEntry{
			JobName:   ev.JobName,
			Status:    ev.Kind,
			OccuredAt: ev.FiredAt,
			Message:   ev.Message,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].OccuredAt.Before(entries[j].OccuredAt)
	})
	return entries
}
