package metrics

import "time"

// RetentionPolicy defines how long snapshots are kept in the collector.
type RetentionPolicy struct {
	// MaxAge is the maximum age of a snapshot before it is pruned.
	MaxAge time.Duration
	// MaxSnapshots is the maximum number of snapshots to retain per job.
	// Zero means unlimited (until MaxAge applies).
	MaxSnapshots int
}

// DefaultRetentionPolicy returns a sensible default retention policy.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MaxAge:       24 * time.Hour,
		MaxSnapshots: 100,
	}
}

// apply filters snapshots according to the policy.
// It expects snapshots sorted ascending by time (oldest first).
func (p RetentionPolicy) apply(snapshots []Snapshot) []Snapshot {
	now := time.Now()

	// Filter by age first.
	var filtered []Snapshot
	for _, s := range snapshots {
		if p.MaxAge <= 0 || now.Sub(s.CollectedAt) <= p.MaxAge {
			filtered = append(filtered, s)
		}
	}

	// Trim to MaxSnapshots keeping the most recent.
	if p.MaxSnapshots > 0 && len(filtered) > p.MaxSnapshots {
		filtered = filtered[len(filtered)-p.MaxSnapshots:]
	}

	return filtered
}
