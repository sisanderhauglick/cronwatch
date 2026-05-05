package metrics

import (
	"fmt"
	"io"
	"strings"
)

// WritePrometheus writes all metrics in Prometheus text exposition format
// to the provided writer.
func (r *Registry) WritePrometheus(w io.Writer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lines := []string{
		"# HELP cronwatch_job_seen_total Total number of successful job check-ins",
		"# TYPE cronwatch_job_seen_total counter",
	}
	for name, s := range r.jobs {
		lines = append(lines,
			fmt.Sprintf(`cronwatch_job_seen_total{job=%q} %d`, name, s.SeenCount),
		)
	}

	lines = append(lines,
		"# HELP cronwatch_job_missed_total Total number of missed job executions",
		"# TYPE cronwatch_job_missed_total counter",
	)
	for name, s := range r.jobs {
		lines = append(lines,
			fmt.Sprintf(`cronwatch_job_missed_total{job=%q} %d`, name, s.MissedCount),
		)
	}

	lines = append(lines,
		"# HELP cronwatch_job_failed_total Total number of failed job executions",
		"# TYPE cronwatch_job_failed_total counter",
	)
	for name, s := range r.jobs {
		lines = append(lines,
			fmt.Sprintf(`cronwatch_job_failed_total{job=%q} %d`, name, s.FailedCount),
		)
	}

	_, err := fmt.Fprintln(w, strings.Join(lines, "\n"))
	return err
}
