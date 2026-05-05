package state

import (
	"log"
	"time"

	"github.com/user/cronwatch/internal/schedule"
)

// MissedEvent describes a cron job whose scheduled run was not observed.
type MissedEvent struct {
	JobName      string
	ScheduledAt  time.Time
}

// Updater reconciles missed-run detection with persistent state.
type Updater struct {
	store    *Store
	grace    time.Duration
	logger   *log.Logger
}

// NewUpdater creates an Updater backed by the given store.
func NewUpdater(store *Store, grace time.Duration, logger *log.Logger) *Updater {
	return &Updater{store: store, grace: grace, logger: logger}
}

// RecordSeen marks a job as successfully observed at the given time.
func (u *Updater) RecordSeen(name string, at time.Time) error {
	return u.store.Set(name, JobState{LastSeen: at, LastStatus: "ok"})
}

// RecordFailed marks a job run as failed at the given time.
func (u *Updater) RecordFailed(name string, at time.Time) error {
	return u.store.Set(name, JobState{LastSeen: at, LastStatus: "failed"})
}

// CheckMissed inspects jobs defined by expr and returns any missed events
// relative to now, using the store's last-seen time as the reference.
func (u *Updater) CheckMissed(name, expr string, now time.Time) ([]MissedEvent, error) {
	st, _ := u.store.Get(name)
	lastSeen := st.LastSeen

	missed, err := schedule.FindMissed(expr, lastSeen, now, u.grace)
	if err != nil {
		return nil, err
	}

	events := make([]MissedEvent, 0, len(missed))
	for _, t := range missed {
		u.logger.Printf("[cronwatch] missed run detected: job=%s scheduled_at=%s", name, t.Format(time.RFC3339))
		events = append(events, MissedEvent{JobName: name, ScheduledAt: t})
	}
	return events, nil
}
