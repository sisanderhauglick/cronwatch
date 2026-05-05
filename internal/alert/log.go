package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogNotifier writes alerts to a writer (default: stderr).
type LogNotifier struct {
	out    io.Writer
	prefix string
}

// NewLogNotifier creates a LogNotifier that writes to w.
// If w is nil, os.Stderr is used.
func NewLogNotifier(w io.Writer, prefix string) *LogNotifier {
	if w == nil {
		w = os.Stderr
	}
	return &LogNotifier{out: w, prefix: prefix}
}

// Send writes the alert as a formatted log line.
func (l *LogNotifier) Send(a Alert) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("%s [%s] CRONWATCH %s — job=%q message=%q\n",
		timestamp,
		l.prefix,
		a.Level,
		a.JobName,
		a.Message,
	)
	_, err := fmt.Fprint(l.out, line)
	return err
}
