package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestNotifyHookLog() *NotifyHookLog {
	return NewNotifyHookLog(10)
}

func TestNotifyHookLog_EmptyAll(t *testing.T) {
	log := newTestNotifyHookLog()
	if got := log.All(""); len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}

func TestNotifyHookLog_RecordAndAll(t *testing.T) {
	log := newTestNotifyHookLog()
	log.Record(NotifyHookEntry{JobName: "backup", Reason: "missed", Target: "webhook", Success: true})
	log.Record(NotifyHookEntry{JobName: "cleanup", Reason: "failed", Target: "email", Success: false, Error: "timeout"})

	all := log.All("")
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestNotifyHookLog_FiltersByJob(t *testing.T) {
	log := newTestNotifyHookLog()
	log.Record(NotifyHookEntry{JobName: "backup", Target: "log", Success: true})
	log.Record(NotifyHookEntry{JobName: "cleanup", Target: "log", Success: true})

	got := log.All("backup")
	if len(got) != 1 || got[0].JobName != "backup" {
		t.Fatalf("expected 1 backup entry, got %+v", got)
	}
}

func TestNotifyHookLog_EvictsOldest(t *testing.T) {
	log := NewNotifyHookLog(3)
	for i := 0; i < 5; i++ {
		log.Record(NotifyHookEntry{JobName: "job", Target: "log", Success: true})
	}
	if got := log.All(""); len(got) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(got))
	}
}

func TestNotifyHookLog_TimestampAutoSet(t *testing.T) {
	log := newTestNotifyHookLog()
	before := time.Now()
	log.Record(NotifyHookEntry{JobName: "j", Target: "log", Success: true})
	after := time.Now()

	e := log.All("")[0]
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Fatalf("unexpected timestamp: %v", e.Timestamp)
	}
}

func TestNotifyHookHandler_ContentType(t *testing.T) {
	log := newTestNotifyHookLog()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notify-hook", nil)
	NotifyHookHandler(log)(rec, req)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestNotifyHookHandler_ReturnsEntries(t *testing.T) {
	log := newTestNotifyHookLog()
	log.Record(NotifyHookEntry{JobName: "nightly", Target: "webhook", Success: false, Error: "refused"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notify-hook", nil)
	NotifyHookHandler(log)(rec, req)

	var entries []NotifyHookEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 || entries[0].JobName != "nightly" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}
