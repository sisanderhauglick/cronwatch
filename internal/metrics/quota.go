package metrics

import (
	"sync"
	"time"
)

// QuotaPolicy defines the maximum number of alerts allowed per job within a window.
type QuotaPolicy struct {
	MaxAlerts int
	Window    time.Duration
}

// quotaEntry tracks alert counts for a single job.
type quotaEntry struct {
	count     int
	windowEnd time.Time
}

// QuotaManager enforces per-job alert quotas over a rolling time window.
type QuotaManager struct {
	mu      sync.Mutex
	policy  QuotaPolicy
	entries map[string]*quotaEntry
	now     func() time.Time
}

// NewQuotaManager creates a QuotaManager with the given policy.
func NewQuotaManager(policy QuotaPolicy) *QuotaManager {
	if policy.MaxAlerts <= 0 {
		policy.MaxAlerts = 10
	}
	if policy.Window <= 0 {
		policy.Window = time.Hour
	}
	return &QuotaManager{
		policy:  policy,
		entries: make(map[string]*quotaEntry),
		now:     time.Now,
	}
}

// Allow returns true if the job is within its alert quota and increments the counter.
// Returns false if the quota has been exhausted for the current window.
func (q *QuotaManager) Allow(jobName string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	e, ok := q.entries[jobName]
	if !ok || now.After(e.windowEnd) {
		q.entries[jobName] = &quotaEntry{
			count:     1,
			windowEnd: now.Add(q.policy.Window),
		}
		return true
	}
	if e.count >= q.policy.MaxAlerts {
		return false
	}
	e.count++
	return true
}

// Stats returns the current alert count and window end for a job.
// Returns zero values if no entry exists.
func (q *QuotaManager) Stats(jobName string) (count int, windowEnd time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.entries[jobName]
	if !ok {
		return 0, time.Time{}
	}
	now := q.now()
	if now.After(e.windowEnd) {
		return 0, time.Time{}
	}
	return e.count, e.windowEnd
}

// Reset clears the quota entry for a job.
func (q *QuotaManager) Reset(jobName string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.entries, jobName)
}
