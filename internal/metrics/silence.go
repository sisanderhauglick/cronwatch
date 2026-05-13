package metrics

import (
	"sync"
	"time"
)

// SilenceRule suppresses alerts for a specific job within a time window.
type SilenceRule struct {
	JobName   string
	StartTime time.Time
	EndTime   time.Time
	Reason    string
}

// SilenceManager tracks active silence rules for jobs.
type SilenceManager struct {
	mu      sync.RWMutex
	rules   []SilenceRule
	nowFunc func() time.Time
}

// NewSilenceManager creates a new SilenceManager.
func NewSilenceManager() *SilenceManager {
	return &SilenceManager{
		nowFunc: time.Now,
	}
}

// Add registers a silence rule for a job.
func (s *SilenceManager) Add(rule SilenceRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, rule)
}

// IsSilenced reports whether the given job is currently silenced.
func (s *SilenceManager) IsSilenced(jobName string) (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.nowFunc()
	for _, r := range s.rules {
		if r.JobName == jobName && !now.Before(r.StartTime) && now.Before(r.EndTime) {
			return true, r.Reason
		}
	}
	return false, ""
}

// Prune removes expired silence rules.
func (s *SilenceManager) Prune() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFunc()
	active := s.rules[:0]
	for _, r := range s.rules {
		if now.Before(r.EndTime) {
			active = append(active, r)
		}
	}
	s.rules = active
}

// List returns a copy of all current silence rules.
func (s *SilenceManager) List() []SilenceRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SilenceRule, len(s.rules))
	copy(out, s.rules)
	return out
}
