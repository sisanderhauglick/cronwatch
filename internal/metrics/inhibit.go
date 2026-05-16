package metrics

import (
	"sync"
	"time"
)

// InhibitRule suppresses alerts for a target job when a source job is failing.
type InhibitRule struct {
	SourceJob string
	TargetJob string
}

// InhibitManager prevents alert noise by suppressing dependent job alerts
// when a upstream source job is already in a failed or missed state.
type InhibitManager struct {
	mu      sync.RWMutex
	rules   []InhibitRule
	registry *Registry
	window  time.Duration
}

// NewInhibitManager creates an InhibitManager using the given registry and
// look-back window to evaluate source job health.
func NewInhibitManager(r *Registry, rules []InhibitRule, window time.Duration) *InhibitManager {
	return &InhibitManager{
		rules:    rules,
		registry: r,
		window:   window,
	}
}

// IsInhibited returns true if any inhibit rule suppresses alerts for targetJob
// because its source job is currently unhealthy.
func (im *InhibitManager) IsInhibited(targetJob string, now time.Time) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	for _, rule := range im.rules {
		if rule.TargetJob != targetJob {
			continue
		}
		snap, ok := im.registry.Get(rule.SourceJob)
		if !ok {
			continue
		}
		if now.Sub(snap.LastSeen) <= im.window {
			if snap.MissedCount > 0 || snap.FailedCount > 0 {
				return true
			}
		}
	}
	return false
}

// AddRule appends a new inhibit rule at runtime.
func (im *InhibitManager) AddRule(rule InhibitRule) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.rules = append(im.rules, rule)
}

// Rules returns a snapshot of the current rule set.
func (im *InhibitManager) Rules() []InhibitRule {
	im.mu.RLock()
	defer im.mu.RUnlock()
	out := make([]InhibitRule, len(im.rules))
	copy(out, im.rules)
	return out
}
