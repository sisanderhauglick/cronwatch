package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestTrendAnalyzer(t *testing.T) (*TrendAnalyzer, *Collector) {
	t.Helper()
	reg := New()
	col := NewCollector(reg, DefaultRetentionPolicy())
	agg := NewAggregator(col)
	return NewTrendAnalyzer(agg, time.Hour, 0.05), col
}

func TestTrendAnalyzer_EmptyCollector(t *testing.T) {
	analyzer, _ := newTestTrendAnalyzer(t)
	trends := analyzer.Analyze(time.Now())
	if len(trends) != 0 {
		t.Fatalf("expected 0 trends, got %d", len(trends))
	}
}

func TestTrendAnalyzer_StableWhenNoPrevious(t *testing.T) {
	analyzer, col := newTestTrendAnalyzer(t)
	now := time.Now().UTC()

	// Record a snapshot only in the current window.
	snap := Snapshot{
		CapturedAt: now.Add(-10 * time.Minute),
		Jobs: map[string]JobStats{
			"backup": {Seen: 5, Failed: 0},
		},
	}
	col.store(snap)

	trends := analyzer.Analyze(now)
	if len(trends) != 1 {
		t.Fatalf("expected 1 trend, got %d", len(trends))
	}
	if trends[0].Direction != TrendStable {
		t.Errorf("expected stable, got %s", trends[0].Direction)
	}
}

func TestTrendAnalyzer_DetectsDegradation(t *testing.T) {
	analyzer, col := newTestTrendAnalyzer(t)
	now := time.Now().UTC()

	// Previous window: 100% success.
	col.store(Snapshot{
		CapturedAt: now.Add(-90 * time.Minute),
		Jobs: map[string]JobStats{"backup": {Seen: 4, Failed: 0}},
	})
	// Current window: 50% success.
	col.store(Snapshot{
		CapturedAt: now.Add(-10 * time.Minute),
		Jobs: map[string]JobStats{"backup": {Seen: 4, Failed: 2}},
	})

	trends := analyzer.Analyze(now)
	if len(trends) != 1 {
		t.Fatalf("expected 1 trend, got %d", len(trends))
	}
	if trends[0].Direction != TrendDown {
		t.Errorf("expected down, got %s", trends[0].Direction)
	}
	if trends[0].Delta >= 0 {
		t.Errorf("expected negative delta, got %f", trends[0].Delta)
	}
}

func TestTrendHandler_ContentTypeAndBody(t *testing.T) {
	analyzer, _ := newTestTrendAnalyzer(t)
	h := TrendHandler(analyzer)

	req := httptest.NewRequest(http.MethodGet, "/metrics/trends", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var result []JobTrend
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}
