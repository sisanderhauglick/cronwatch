package alert

import (
	"fmt"
	"log"
	"time"
)

// Dispatcher routes alerts to one or more Notifiers.
type Dispatcher struct {
	notifiers []Notifier
	logger    *log.Logger
}

// NewDispatcher creates a Dispatcher with the given notifiers.
func NewDispatcher(logger *log.Logger, notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{notifiers: notifiers, logger: logger}
}

// Missed constructs and dispatches a "missed" alert for the named job.
func (d *Dispatcher) Missed(jobName string) {
	d.dispatch(Alert{
		JobName:   jobName,
		Level:     LevelMissed,
		Message:   fmt.Sprintf("cron job %q missed its scheduled run", jobName),
		Timestamp: time.Now().UTC(),
	})
}

// Failed constructs and dispatches a "failed" alert for the named job.
func (d *Dispatcher) Failed(jobName string, reason string) {
	d.dispatch(Alert{
		JobName:   jobName,
		Level:     LevelFailed,
		Message:   fmt.Sprintf("cron job %q failed: %s", jobName, reason),
		Timestamp: time.Now().UTC(),
	})
}

func (d *Dispatcher) dispatch(a Alert) {
	for _, n := range d.notifiers {
		if err := n.Send(a); err != nil {
			d.logger.Printf("alert dispatcher: send error: %v", err)
		}
	}
}
