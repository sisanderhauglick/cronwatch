package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestEscalationManager() (*EscalationManager, *time.Time) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mgr := NewEscalationManager(EscalationPolicy{
		WarnAfter:     5 * time.Minute,
		CriticalAfter: 15 * time.Minute,
	})
	mgr.now = func() time.Time { return now }
	return mgr, &now
}

func TestEscalation_NoneWhenHealthy(t *testing.T) {
	mgr, _ := newTestEscalationManager()
	s := mgr.Evaluate("backup", false)
	if s.Level != LevelNone {
		t.Fatalf("expected none, got %s", s.Level)
	}
}

func TestEscalation_NoneBeforeWarnThreshold(t *testing.T) {
	mgr, now := newTestEscalationManager()
	mgr.Evaluate("backup", true)
	*now = now.Add(2 * time.Minute)
	s := mgr.Evaluate("backup", true)
	if s.Level != LevelNone {
		t.Fatalf("expected none within grace, got %s", s.Level)
	}
}

func TestEscalation_WarnAfterThreshold(t *testing.T) {
	mgr, now := newTestEscalationManager()
	mgr.Evaluate("backup", true)
	*now = now.Add(6 * time.Minute)
	s := mgr.Evaluate("backup", true)
	if s.Level != LevelWarn {
		t.Fatalf("expected warn, got %s", s.Level)
	}
}

func TestEscalation_CriticalAfterThreshold(t *testing.T) {
	mgr, now := newTestEscalationManager()
	mgr.Evaluate("backup", true)
	*now = now.Add(20 * time.Minute)
	s := mgr.Evaluate("backup", true)
	if s.Level != LevelCritical {
		t.Fatalf("expected critical, got %s", s.Level)
	}
}

func TestEscalation_ResetsOnHealthy(t *testing.T) {
	mgr, now := newTestEscalationManager()
	mgr.Evaluate("backup", true)
	*now = now.Add(20 * time.Minute)
	mgr.Evaluate("backup", true)
	s := mgr.Evaluate("backup", false)
	if s.Level != LevelNone {
		t.Fatalf("expected reset to none, got %s", s.Level)
	}
	if !s.Since.IsZero() {
		t.Fatal("expected Since to be zero after reset")
	}
}

func TestEscalationHandler_ContentType(t *testing.T) {
	mgr, _ := newTestEscalationManager()
	h := EscalationHandler(mgr)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/escalation", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestEscalationHandler_EmptyWhenAllHealthy(t *testing.T) {
	mgr, _ := newTestEscalationManager()
	mgr.Evaluate("job1", false)
	h := EscalationHandler(mgr)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/escalation", nil))
	var out []escalationResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d entries", len(out))
	}
}

func TestEscalationHandler_ReturnsEscalatedJobs(t *testing.T) {
	mgr, now := newTestEscalationManager()
	mgr.Evaluate("job1", true)
	*now = now.Add(10 * time.Minute)
	mgr.Evaluate("job1", true)
	h := EscalationHandler(mgr)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/escalation", nil))
	var out []escalationResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Job != "job1" || out[0].Level != "warn" {
		t.Fatalf("unexpected entry: %+v", out[0])
	}
}
