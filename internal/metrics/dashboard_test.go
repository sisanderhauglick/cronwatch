package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestDashboardHandler(t *testing.T) (*Registry, http.HandlerFunc) {
	t.Helper()
	r := New()
	return r, DashboardHandler(r)
}

func TestDashboardHandler_EmptyRegistry(t *testing.T) {
	_, h := newTestDashboardHandler(t)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var summary DashboardSummary
	if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if summary.TotalJobs != 0 {
		t.Errorf("expected 0 jobs, got %d", summary.TotalJobs)
	}
}

func TestDashboardHandler_ContentType(t *testing.T) {
	_, h := newTestDashboardHandler(t)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestDashboardHandler_HealthCounts(t *testing.T) {
	r, h := newTestDashboardHandler(t)

	r.RecordSeen("job-a")
	r.RecordSeen("job-a")
	r.RecordMissed("job-b")
	r.RecordFailed("job-c")

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	var summary DashboardSummary
	if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if summary.TotalJobs != 3 {
		t.Errorf("expected 3 total jobs, got %d", summary.TotalJobs)
	}
	if summary.Healthy != 1 {
		t.Errorf("expected 1 healthy, got %d", summary.Healthy)
	}
	if summary.Degraded != 2 {
		t.Errorf("expected 2 degraded, got %d", summary.Degraded)
	}
}

func TestDashboardHandler_SuccessRate(t *testing.T) {
	r, h := newTestDashboardHandler(t)

	r.RecordSeen("job-x")
	r.RecordSeen("job-x")
	r.RecordFailed("job-x")

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	var summary DashboardSummary
	if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(summary.Jobs) != 1 {
		t.Fatalf("expected 1 job entry, got %d", len(summary.Jobs))
	}

	entry := summary.Jobs[0]
	want := 2.0 / 3.0
	if entry.SuccessRate < want-0.001 || entry.SuccessRate > want+0.001 {
		t.Errorf("expected success rate ~%.4f, got %.4f", want, entry.SuccessRate)
	}
}
