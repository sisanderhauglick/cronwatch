// Package metrics tracks runtime counters for cronwatch job execution.
package metrics

import (
	"sync"
	"time"
)

// JobStats holds execution statistics for a single cron job.
type JobStats struct {
	Name        string
	SeenCount   int64
	MissedCount int64
	FailedCount int64
	LastSeen    time.Time
	LastMissed  time.Time
	LastFailed  time.Time
}

// Registry stores metrics for all monitored jobs.
type Registry struct {
	mu   sync.RWMutex
	jobs map[string]*JobStats
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{jobs: make(map[string]*JobStats)}
}

func (r *Registry) ensure(name string) *JobStats {
	if s, ok := r.jobs[name]; ok {
		return s
	}
	s := &JobStats{Name: name}
	r.jobs[name] = s
	return s
}

// RecordSeen increments the seen counter for the named job.
func (r *Registry) RecordSeen(name string, at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.ensure(name)
	s.SeenCount++
	s.LastSeen = at
}

// RecordMissed increments the missed counter for the named job.
func (r *Registry) RecordMissed(name string, at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.ensure(name)
	s.MissedCount++
	s.LastMissed = at
}

// RecordFailed increments the failed counter for the named job.
func (r *Registry) RecordFailed(name string, at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.ensure(name)
	s.FailedCount++
	s.LastFailed = at
}

// Snapshot returns a copy of stats for all jobs.
func (r *Registry) Snapshot() []JobStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]JobStats, 0, len(r.jobs))
	for _, s := range r.jobs {
		out = append(out, *s)
	}
	return out
}

// Get returns stats for a single job and whether it exists.
func (r *Registry) Get(name string) (JobStats, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.jobs[name]
	if !ok {
		return JobStats{}, false
	}
	return *s, true
}
