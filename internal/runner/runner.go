package runner

import (
	"log"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
	"github.com/yourorg/cronwatch/internal/config"
	"github.com/yourorg/cronwatch/internal/state"
)

// Runner periodically checks all configured jobs for missed runs and
// dispatches alerts via the provided dispatcher.
type Runner struct {
	cfg        *config.Config
	updater    *state.Updater
	dispatcher *alert.Dispatcher
	interval   time.Duration
	stopCh     chan struct{}
}

// New creates a new Runner.
func New(cfg *config.Config, updater *state.Updater, dispatcher *alert.Dispatcher, interval time.Duration) *Runner {
	return &Runner{
		cfg:        cfg,
		updater:    updater,
		dispatcher: dispatcher,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// Start begins the check loop in the current goroutine. Call Stop to exit.
func (r *Runner) Start() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	log.Printf("runner: starting, check interval=%s", r.interval)
	for {
		select {
		case <-ticker.C:
			r.tick(time.Now())
		case <-r.stopCh:
			log.Println("runner: stopped")
			return
		}
	}
}

// Stop signals the runner to exit after the current tick completes.
func (r *Runner) Stop() {
	close(r.stopCh)
}

// tick runs a single check cycle at the given wall-clock time.
func (r *Runner) tick(now time.Time) {
	for _, job := range r.cfg.Jobs {
		alerts, err := r.updater.CheckMissed(job, now)
		if err != nil {
			log.Printf("runner: check missed error for job %q: %v", job.Name, err)
			continue
		}
		for _, a := range alerts {
			if err := r.dispatcher.Send(a); err != nil {
				log.Printf("runner: alert dispatch error for job %q: %v", job.Name, err)
			}
		}
	}
}
