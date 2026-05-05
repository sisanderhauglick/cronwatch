package alert

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogNotifier_Send_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	ln := NewLogNotifier(&buf, "INFO")

	a := makeAlert()
	if err := ln.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, a.JobName) {
		t.Errorf("expected job name %q in output, got: %s", a.JobName, out)
	}
	if !strings.Contains(out, string(a.Level)) {
		t.Errorf("expected level %q in output, got: %s", a.Level, out)
	}
	if !strings.Contains(out, a.Message) {
		t.Errorf("expected message %q in output, got: %s", a.Message, out)
	}
}

func TestLogNotifier_Send_ContainsPrefix(t *testing.T) {
	var buf bytes.Buffer
	ln := NewLogNotifier(&buf, "MYPREFIX")

	if err := ln.Send(makeAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "MYPREFIX") {
		t.Errorf("expected prefix in output, got: %s", buf.String())
	}
}

func TestLogNotifier_Send_DefaultsToStderr(t *testing.T) {
	// Just ensure no panic when w is nil; output goes to stderr.
	ln := NewLogNotifier(nil, "TEST")
	if ln.out == nil {
		t.Error("expected non-nil writer")
	}
}

func TestLogNotifier_Send_WriterError(t *testing.T) {
	ln := NewLogNotifier(&errorWriter{}, "ERR")
	if err := ln.Send(makeAlert()); err == nil {
		t.Error("expected error from failing writer")
	}
}

// errorWriter always returns an error.
type errorWriter struct{}

func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
