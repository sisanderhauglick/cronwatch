package metrics

import (
	"sync"
	"time"
)

// EscalationLevel represents the severity of an alert escalation.
type EscalationLevel int

const (
	LevelNone EscalationLevel = iota
	LevelWarn
	LevelCritical
)

func (l EscalationLevel) String() string {
	switch l {
	case LevelWarn:
		return "warn"
	case LevelCritical:
		return "critical"
	default:
		return "none"
	}
}

// EscalationPolicy defines thresholds for escalating alerts.
type EscalationPolicy struct {
	// WarnAfter is how long a job must be failing/missed before warn level.
	WarnAfter time.Duration
	// CriticalAfter is how long before escalating to critical.
	CriticalAfter time.Duration
}

// EscalationState tracks per-job escalation state.
type EscalationState struct {
	Level     EscalationLevel
	Since     time.Time
	JobName   string
	LastCheck time.Time
}

// EscalationManager evaluates escalation levels for jobs.
type EscalationManager struct {
	mu     sync.Mutex
	policy EscalationPolicy
	states map[string]*EscalationState
	now    func() time.Time
}

// NewEscalationManager creates a new EscalationManager with the given policy.
func NewEscalationManager(policy EscalationPolicy) *EscalationManager {
	return &EscalationManager{
		policy: policy,
		states: make(map[string]*EscalationState),
		now:    time.Now,
	}
}

// Evaluate updates and returns the escalation level for a job given its health.
// unhealthy should be true when the job is failed or missed.
func (m *EscalationManager) Evaluate(jobName string, unhealthy bool) EscalationState {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	state, ok := m.states[jobName]
	if !ok {
		state = &EscalationState{JobName: jobName}
		m.states[jobName] = state
	}

	if !unhealthy {
		state.Level = LevelNone
		state.Since = time.Time{}
		state.LastCheck = now
		return *state
	}

	if state.Since.IsZero() {
		state.Since = now
	}

	duration := now.Sub(state.Since)
	switch {
	case duration >= m.policy.CriticalAfter:
		state.Level = LevelCritical
	case duration >= m.policy.WarnAfter:
		state.Level = LevelWarn
	default:
		state.Level = LevelNone
	}
	state.LastCheck = now
	return *state
}

// All returns a snapshot of all current escalation states.
func (m *EscalationManager) All() []EscalationState {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]EscalationState, 0, len(m.states))
	for _, s := range m.states {
		out = append(out, *s)
	}
	return out
}
