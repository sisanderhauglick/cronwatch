package metrics

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker for a job.
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

// CircuitEntry holds the breaker state for a single job.
type CircuitEntry struct {
	Job          string       `json:"job"`
	State        CircuitState `json:"state"`
	Failures     int          `json:"failures"`
	LastFailure  time.Time    `json:"last_failure"`
	OpenedAt     time.Time    `json:"opened_at,omitempty"`
	ResetAfter   time.Duration `json:"reset_after_seconds"`
}

// CircuitBreaker tracks per-job failure counts and opens the circuit
// when a threshold is exceeded, suppressing alerts until reset.
type CircuitBreaker struct {
	mu        sync.Mutex
	entries   map[string]*CircuitEntry
	threshold int
	resetAfter time.Duration
	now       func() time.Time
}

// NewCircuitBreaker creates a CircuitBreaker that opens after threshold
// consecutive failures and resets after resetAfter duration.
func NewCircuitBreaker(threshold int, resetAfter time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		entries:    make(map[string]*CircuitEntry),
		threshold:  threshold,
		resetAfter: resetAfter,
		now:        time.Now,
	}
}

// RecordFailure increments the failure count for a job and may open the circuit.
func (cb *CircuitBreaker) RecordFailure(job string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.getOrCreate(job)
	e.Failures++
	e.LastFailure = cb.now()
	if e.State == CircuitClosed && e.Failures >= cb.threshold {
		e.State = CircuitOpen
		e.OpenedAt = cb.now()
	}
}

// RecordSuccess resets the failure count and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess(job string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.getOrCreate(job)
	e.Failures = 0
	e.State = CircuitClosed
}

// IsOpen returns true when the circuit is open and the reset window has not elapsed.
// If the reset window has elapsed the circuit transitions to half-open.
func (cb *CircuitBreaker) IsOpen(job string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e, ok := cb.entries[job]
	if !ok {
		return false
	}
	if e.State == CircuitOpen && cb.now().Sub(e.OpenedAt) >= cb.resetAfter {
		e.State = CircuitHalfOpen
	}
	return e.State == CircuitOpen
}

// All returns a snapshot of all circuit entries.
func (cb *CircuitBreaker) All() []CircuitEntry {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	out := make([]CircuitEntry, 0, len(cb.entries))
	for _, e := range cb.entries {
		out = append(out, *e)
	}
	return out
}

func (cb *CircuitBreaker) getOrCreate(job string) *CircuitEntry {
	if e, ok := cb.entries[job]; ok {
		return e
	}
	e := &CircuitEntry{Job: job, State: CircuitClosed, ResetAfter: cb.resetAfter}
	cb.entries[job] = e
	return e
}
