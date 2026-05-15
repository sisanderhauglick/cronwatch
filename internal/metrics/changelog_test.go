package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestChangeLog() *ChangeLog { return NewChangeLog(10) }

func TestChangeLog_NoEventOnFirstSeen(t *testing.T) {
	cl := newTestChangeLog()
	cl.Record("backup", "ok", time.Now())
	// first observation has no prior state — still recorded as a transition from ""
	events := cl.Recent(10)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].From != "" {
		t.Errorf("expected empty From, got %q", events[0].From)
	}
}

func TestChangeLog_RecordsTransition(t *testing.T) {
	cl := newTestChangeLog()
	now := time.Now()
	cl.Record("backup", "ok", now)
	cl.Record("backup", "missed", now.Add(time.Minute))

	events := cl.Recent(10)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	// Recent returns newest first
	if events[0].To != "missed" {
		t.Errorf("expected newest event To=missed, got %q", events[0].To)
	}
	if events[0].From != "ok" {
		t.Errorf("expected From=ok, got %q", events[0].From)
	}
}

func TestChangeLog_NoDuplicateStatus(t *testing.T) {
	cl := newTestChangeLog()
	now := time.Now()
	cl.Record("job", "ok", now)
	cl.Record("job", "ok", now.Add(time.Second))

	if got := len(cl.Recent(10)); got != 1 {
		t.Errorf("expected 1 event (no duplicate), got %d", got)
	}
}

func TestChangeLog_EvictsOldest(t *testing.T) {
	cl := NewChangeLog(3)
	now := time.Now()
	statuses := []string{"ok", "missed", "ok", "failed", "ok"}
	for i, s := range statuses {
		cl.Record("job", s, now.Add(time.Duration(i)*time.Minute))
	}
	if got := len(cl.Recent(10)); got != 3 {
		t.Errorf("expected 3 events after eviction, got %d", got)
	}
}

func TestChangelogHandler_ContentType(t *testing.T) {
	cl := newTestChangeLog()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/changelog", nil)
	ChangelogHandler(cl)(rec, req)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestChangelogHandler_EmptyReturnsEmptySlice(t *testing.T) {
	cl := newTestChangeLog()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/changelog", nil)
	ChangelogHandler(cl)(rec, req)

	var out []ChangeEvent
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty slice, got %d items", len(out))
	}
}

func TestChangelogHandler_LimitParam(t *testing.T) {
	cl := newTestChangeLog()
	now := time.Now()
	statuses := []string{"ok", "missed", "ok", "failed", "ok"}
	for i, s := range statuses {
		cl.Record("job", s, now.Add(time.Duration(i)*time.Minute))
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/changelog?limit=2", nil)
	ChangelogHandler(cl)(rec, req)

	var out []ChangeEvent
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 events with limit=2, got %d", len(out))
	}
}
