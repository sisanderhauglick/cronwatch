package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestRunLog() *RunLog {
	return NewRunLog(10)
}

func TestRunLog_EmptyQuery(t *testing.T) {
	rl := newTestRunLog()
	got := rl.Query("", time.Time{}, time.Time{})
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}

func TestRunLog_RecordAndQuery(t *testing.T) {
	rl := newTestRunLog()
	now := time.Now()
	rl.Record(RunEntry{Job: "backup", StartedAt: now, Status: "ok"})
	rl.Record(RunEntry{Job: "cleanup", StartedAt: now.Add(time.Minute), Status: "failed"})

	all := rl.Query("", time.Time{}, time.Time{})
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	filtered := rl.Query("backup", time.Time{}, time.Time{})
	if len(filtered) != 1 || filtered[0].Job != "backup" {
		t.Fatalf("expected 1 backup entry, got %+v", filtered)
	}
}

func TestRunLog_EvictsOldest(t *testing.T) {
	rl := NewRunLog(3)
	now := time.Now()
	for i := 0; i < 5; i++ {
		rl.Record(RunEntry{Job: "job", StartedAt: now.Add(time.Duration(i) * time.Minute), Status: "ok"})
	}
	all := rl.Query("", time.Time{}, time.Time{})
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
}

func TestRunLog_TimeRangeFilter(t *testing.T) {
	rl := newTestRunLog()
	base := time.Now().Truncate(time.Second)
	rl.Record(RunEntry{Job: "j", StartedAt: base, Status: "ok"})
	rl.Record(RunEntry{Job: "j", StartedAt: base.Add(2 * time.Hour), Status: "ok"})
	rl.Record(RunEntry{Job: "j", StartedAt: base.Add(4 * time.Hour), Status: "ok"})

	got := rl.Query("", base.Add(time.Hour), base.Add(3*time.Hour))
	if len(got) != 1 {
		t.Fatalf("expected 1 entry in range, got %d", len(got))
	}
}

func TestRunLogHandler_ContentTypeAndBody(t *testing.T) {
	rl := newTestRunLog()
	now := time.Now()
	rl.Record(RunEntry{Job: "sync", StartedAt: now, Status: "ok", Message: "done"})

	handler := RunLogHandler(rl)
	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	var entries []RunEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 || entries[0].Job != "sync" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestRunLogHandler_FiltersByJob(t *testing.T) {
	rl := newTestRunLog()
	now := time.Now()
	rl.Record(RunEntry{Job: "a", StartedAt: now, Status: "ok"})
	rl.Record(RunEntry{Job: "b", StartedAt: now, Status: "ok"})

	handler := RunLogHandler(rl)
	req := httptest.NewRequest(http.MethodGet, "/runlog?job=a", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	var entries []RunEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 || entries[0].Job != "a" {
		t.Fatalf("expected only job a, got %+v", entries)
	}
}
