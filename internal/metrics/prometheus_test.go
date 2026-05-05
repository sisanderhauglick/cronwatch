package metrics

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWritePrometheus_Empty(t *testing.T) {
	r := New()
	var buf bytes.Buffer
	if err := r.WritePrometheus(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "cronwatch_job_seen_total") {
		t.Error("expected seen_total header in output")
	}
	if !strings.Contains(out, "cronwatch_job_missed_total") {
		t.Error("expected missed_total header in output")
	}
}

func TestWritePrometheus_WithData(t *testing.T) {
	r := New()
	r.RecordSeen("backup")
	r.RecordSeen("backup")
	r.RecordMissed("backup")
	r.RecordFailed("sync")

	var buf bytes.Buffer
	if err := r.WritePrometheus(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, `cronwatch_job_seen_total{job="backup"} 2`) {
		t.Errorf("expected seen count 2 for backup, got:\n%s", out)
	}
	if !strings.Contains(out, `cronwatch_job_missed_total{job="backup"} 1`) {
		t.Errorf("expected missed count 1 for backup, got:\n%s", out)
	}
	if !strings.Contains(out, `cronwatch_job_failed_total{job="sync"} 1`) {
		t.Errorf("expected failed count 1 for sync, got:\n%s", out)
	}
}

func TestPrometheusHandler_ContentType(t *testing.T) {
	r := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	PrometheusHandler(r)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain content-type, got %q", ct)
	}
}

func TestPrometheusHandler_Body(t *testing.T) {
	r := New()
	r.RecordSeen("myjob")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	PrometheusHandler(r)(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "myjob") {
		t.Errorf("expected myjob in body, got:\n%s", body)
	}
}
