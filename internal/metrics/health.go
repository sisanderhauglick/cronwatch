package metrics

import "time"

// HealthStatus represents the overall health of a monitored job.
type HealthStatus string

const (
	HealthOK      HealthStatus = "ok"
	HealthDegraded HealthStatus = "degraded"
	HealthDown    HealthStatus = "down"
	HealthUnknown HealthStatus = "unknown"
)

// JobHealth summarises the current health of a single job.
type JobHealth struct {
	JobName    string       `json:"job_name"`
	Status     HealthStatus `json:"status"`
	LastSeen   *time.Time   `json:"last_seen,omitempty"`
	MissedRuns int          `json:"missed_runs"`
	FailedRuns int          `json:"failed_runs"`
	Message    string       `json:"message,omitempty"`
}

// HealthEvaluator derives a HealthStatus from a job's current metrics.
type HealthEvaluator struct {
	MissedThreshold int
	FailedThreshold int
	StaleAfter      time.Duration
}

// NewHealthEvaluator returns an evaluator with sensible defaults.
func NewHealthEvaluator() *HealthEvaluator {
	return &HealthEvaluator{
		MissedThreshold: 1,
		FailedThreshold: 1,
		StaleAfter:      24 * time.Hour,
	}
}

// Evaluate returns a JobHealth for the given snapshot and current time.
func (e *HealthEvaluator) Evaluate(snap Snapshot, now time.Time) JobHealth {
	h := JobHealth{
		JobName:    snap.JobName,
		MissedRuns: snap.Missed,
		FailedRuns: snap.Failed,
	}

	if !snap.LastSeen.IsZero() {
		t := snap.LastSeen
		h.LastSeen = &t
	}

	switch {
	case snap.Missed >= e.MissedThreshold:
		h.Status = HealthDown
		h.Message = "job has missed scheduled runs"
	case snap.Failed >= e.FailedThreshold:
		h.Status = HealthDegraded
		h.Message = "job has recent failures"
	case h.LastSeen != nil && now.Sub(*h.LastSeen) > e.StaleAfter:
		h.Status = HealthDegraded
		h.Message = "job has not been seen recently"
	case h.LastSeen == nil:
		h.Status = HealthUnknown
		h.Message = "no data recorded yet"
	default:
		h.Status = HealthOK
	}

	return h
}
