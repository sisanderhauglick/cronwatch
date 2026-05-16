package metrics

import (
	"sync"
	"time"
)

// StaleEntry describes a job whose last snapshot is older than the staleness threshold.
type StaleEntry struct {
	Job         string        `json:"job"`
	LastSeen    time.Time     `json:"last_seen"`
	Staleness   time.Duration `json:"staleness_seconds"`
}

// StaleDetector identifies jobs that have not reported within a configurable window.
type StaleDetector struct {
	mu        sync.Mutex
	collector *Collector
	threshold time.Duration
}

// NewStaleDetector creates a StaleDetector with the given staleness threshold.
func NewStaleDetector(c *Collector, threshold time.Duration) *StaleDetector {
	return &StaleDetector{
		collector: c,
		threshold: threshold,
	}
}

// Detect returns all jobs whose most recent snapshot is older than the threshold.
func (s *StaleDetector) Detect(now time.Time) []StaleEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	var entries []StaleEntry

	for _, job := range s.collector.Jobs() {
		snap, ok := s.collector.Latest(job)
		if !ok {
			continue
		}
		age := now.Sub(snap.Timestamp)
		if age > s.threshold {
			entries = append(entries, StaleEntry{
				Job:       job,
				LastSeen:  snap.Timestamp,
				Staleness: age,
			})
		}
	}

	return entries
}

// SetThreshold updates the staleness threshold.
func (s *StaleDetector) SetThreshold(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.threshold = d
}
