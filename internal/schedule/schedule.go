package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// NextRun returns the next scheduled run time after 'from' for the given cron expression.
func NextRun(expr string, from time.Time) (time.Time, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := p.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron expression %q: %w", expr, err)
	}
	return sched.Next(from), nil
}

// PrevRun returns the most recent scheduled run time at or before 'at' for the given cron expression.
// It approximates by stepping back in time from 'at'.
func PrevRun(expr string, at time.Time) (time.Time, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := p.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron expression %q: %w", expr, err)
	}

	// Walk back up to 366 days in 1-minute steps to find the last scheduled time.
	const maxSteps = 366 * 24 * 60
	cursor := at.Add(-time.Minute)
	for i := 0; i < maxSteps; i++ {
		next := sched.Next(cursor)
		if !next.After(at) {
			return next, nil
		}
		cursor = cursor.Add(-time.Minute)
	}
	return time.Time{}, fmt.Errorf("could not determine previous run for expression %q", expr)
}

// IsDue reports whether a job with the given expression was due between start and end.
func IsDue(expr string, start, end time.Time) (bool, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := p.Parse(expr)
	if err != nil {
		return false, fmt.Errorf("parse cron expression %q: %w", expr, err)
	}
	next := sched.Next(start)
	return !next.After(end), nil
}
