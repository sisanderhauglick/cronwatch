package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestRollupAnalyzer() (*RollupAnalyzer, *Collector) {
	c := NewCollector(DefaultRetentionPolicy())
	return NewRollupAnalyzer(c), c
}

func TestRollup_EmptyCollector(t *testing.T) {
	a, _ := newTestRollupAnalyzer()
	now := time.Now()
	buckets := a.Rollup(now.Add(-time.Hour), now, "1h")
	if len(buckets) != 0 {
		t.Fatalf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestRollup_AggregatesWithinWindow(t *testing.T) {
	a, c := newTestRollupAnalyzer()
	now := time.Now()

	c.Collect(Snapshot{Job: "backup", Timestamp: now.Add(-30 * time.Minute), Seen: 2, Failed: 1, Missed: 0})
	c.Collect(Snapshot{Job: "backup", Timestamp: now.Add(-10 * time.Minute), Seen: 1, Failed: 0, Missed: 0})

	buckets := a.Rollup(now.Add(-time.Hour), now, "1h")
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	b := buckets[0]
	if b.Runs != 3 {
		t.Errorf("expected runs=3, got %d", b.Runs)
	}
	if b.Failed != 1 {
		t.Errorf("expected failed=1, got %d", b.Failed)
	}
}

func TestRollup_ExcludesOutsideWindow(t *testing.T) {
	a, c := newTestRollupAnalyzer()
	now := time.Now()

	c.Collect(Snapshot{Job: "backup", Timestamp: now.Add(-2 * time.Hour), Seen: 5, Failed: 2})

	buckets := a.Rollup(now.Add(-time.Hour), now, "1h")
	if len(buckets) != 0 {
		t.Fatalf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestRollup_SuccessRate(t *testing.T) {
	a, c := newTestRollupAnalyzer()
	now := time.Now()

	c.Collect(Snapshot{Job: "job", Timestamp: now.Add(-5 * time.Minute), Seen: 4, Failed: 2, Missed: 0})

	buckets := a.Rollup(now.Add(-time.Hour), now, "1h")
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket")
	}
	if buckets[0].Success != 0.5 {
		t.Errorf("expected success_rate=0.5, got %f", buckets[0].Success)
	}
}

func TestRollupHandler_ContentTypeAndBody(t *testing.T) {
	a, c := newTestRollupAnalyzer()
	now := time.Now()
	c.Collect(Snapshot{Job: "j", Timestamp: now.Add(-10 * time.Minute), Seen: 1})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/rollup?hours=1", nil)
	RollupHandler(a)(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected content-type: %s", ct)
	}
	var buckets []RollupBucket
	if err := json.NewDecoder(rec.Body).Decode(&buckets); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(buckets) == 0 {
		t.Error("expected at least one bucket")
	}
}
