package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestDigestHandler(t *testing.T) (http.Handler, *Collector) {
	t.Helper()
	c := NewCollector(DefaultRetentionPolicy(100, 24*time.Hour))
	reg := New()
	h := NewHealthEvaluator(c, reg, 2*time.Hour)
	a := NewDigestAnalyzer(c, h, time.Hour)
	return NewDigestHandler(a), c
}

func TestDigestHandler_ContentType(t *testing.T) {
	h, _ := newTestDigestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/digest", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestDigestHandler_EmptyReturnsEmptySlice(t *testing.T) {
	h, _ := newTestDigestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/digest", nil))

	var report DigestReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(report.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(report.Entries))
	}
}

func TestDigestHandler_WithData(t *testing.T) {
	h, c := newTestDigestHandler(t)
	now := time.Now()

	c.Collect(Snapshot{
		Timestamp: now.Add(-20 * time.Minute),
		Jobs: map[string]JobStats{
			"etl": {Seen: 2, Failed: 0},
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/digest", nil))

	var report DigestReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	if report.Entries[0].JobName != "etl" {
		t.Errorf("unexpected job name: %s", report.Entries[0].JobName)
	}
}

func TestDigestHandler_WindowOverride(t *testing.T) {
	h, c := newTestDigestHandler(t)
	now := time.Now()

	// snapshot 90 minutes ago — outside default 1h window but inside 2h
	c.Collect(Snapshot{
		Timestamp: now.Add(-90 * time.Minute),
		Jobs: map[string]JobStats{
			"report": {Seen: 1},
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/digest?window=120", nil)
	h.ServeHTTP(rec, req)

	var report DigestReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(report.Entries) != 1 {
		t.Errorf("expected 1 entry with 2h window override, got %d", len(report.Entries))
	}
}
