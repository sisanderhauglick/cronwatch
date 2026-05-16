package metrics

import (
	"sync"
	"time"
)

// SuppressionRule defines a time window during which alerts for a job are suppressed.
type SuppressionRule struct {
	JobName   string
	Start     time.Time
	End       time.Time
	Reason    string
}

// SuppressionManager tracks active suppression windows for jobs.
type SuppressionManager struct {
	mu    sync.RWMutex
	rules []SuppressionRule
}

// NewSuppressionManager returns an empty SuppressionManager.
func NewSuppressionManager() *SuppressionManager {
	return &SuppressionManager{}
}

// Add registers a new suppression rule.
func (s *SuppressionManager) Add(rule SuppressionRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, rule)
}

// IsSuppressed reports whether alerts for jobName should be suppressed at t.
func (s *SuppressionManager) IsSuppressed(jobName string, t time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rules {
		if r.JobName == jobName && !t.Before(r.Start) && t.Before(r.End) {
			return true
		}
	}
	return false
}

// Prune removes rules whose End time is before now.
func (s *SuppressionManager) Prune(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	active := s.rules[:0]
	for _, r := range s.rules {
		if !r.End.Before(now) {
			active = append(active, r)
		}
	}
	s.rules = active
}

// Active returns a copy of all current suppression rules.
func (s *SuppressionManager) Active() []SuppressionRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SuppressionRule, len(s.rules))
	copy(out, s.rules)
	return out
}
