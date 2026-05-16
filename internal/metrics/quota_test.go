package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestQuotaManager(max int, window time.Duration) *QuotaManager {
	qm := NewQuotaManager(QuotaPolicy{MaxAlerts: max, Window: window})
	fixed := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	qm.now = func() time.Time { return fixed }
	return qm
}

func TestQuota_FirstAlwaysAllowed(t *testing.T) {
	qm := newTestQuotaManager(3, time.Hour)
	if !qm.Allow("backup") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestQuota_ExhaustedAfterMax(t *testing.T) {
	qm := newTestQuotaManager(2, time.Hour)
	qm.Allow("backup")
	qm.Allow("backup")
	if qm.Allow("backup") {
		t.Fatal("expected third alert to be denied")
	}
}

func TestQuota_ResetsAfterWindow(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	qm := NewQuotaManager(QuotaPolicy{MaxAlerts: 1, Window: time.Hour})
	qm.now = func() time.Time { return base }

	qm.Allow("backup")
	if qm.Allow("backup") {
		t.Fatal("expected second call to be denied within window")
	}

	// Advance past window
	qm.now = func() time.Time { return base.Add(2 * time.Hour) }
	if !qm.Allow("backup") {
		t.Fatal("expected allow after window reset")
	}
}

func TestQuota_Reset(t *testing.T) {
	qm := newTestQuotaManager(1, time.Hour)
	qm.Allow("backup")
	qm.Reset("backup")
	if !qm.Allow("backup") {
		t.Fatal("expected allow after manual reset")
	}
}

func TestQuota_StatsReturnsCount(t *testing.T) {
	qm := newTestQuotaManager(5, time.Hour)
	qm.Allow("nightly")
	qm.Allow("nightly")
	count, end := qm.Stats("nightly")
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	if end.IsZero() {
		t.Fatal("expected non-zero window end")
	}
}

func TestQuotaHandler_ContentType(t *testing.T) {
	qm := newTestQuotaManager(5, time.Hour)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/quota", nil)
	QuotaHandler(qm)(rec, req)
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestQuotaHandler_ReturnsActiveEntries(t *testing.T) {
	qm := newTestQuotaManager(5, time.Hour)
	qm.Allow("job-a")
	qm.Allow("job-a")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/quota", nil)
	QuotaHandler(qm)(rec, req)

	var results []quotaStatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", results[0].Count)
	}
}
