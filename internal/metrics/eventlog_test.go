package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestEventLog() *EventLog {
	return NewEventLog(5)
}

func TestEventLog_EmptyAll(t *testing.T) {
	el := newTestEventLog()
	if got := el.All(); len(got) != 0 {
		t.Fatalf("expected empty, got %d entries", len(got))
	}
}

func TestEventLog_RecordAndAll(t *testing.T) {
	el := newTestEventLog()
	el.Record("backup", SeverityInfo, "started")
	el.Record("backup", SeverityWarn, "slow")
	all := el.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Job != "backup" || all[0].Severity != SeverityInfo {
		t.Errorf("unexpected first entry: %+v", all[0])
	}
}

func TestEventLog_EvictsOldest(t *testing.T) {
	el := NewEventLog(3)
	for i := 0; i < 5; i++ {
		el.Record("job", SeverityInfo, "msg")
	}
	if got := len(el.All()); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestEventLog_FiltersByJob(t *testing.T) {
	el := newTestEventLog()
	el.Record("alpha", SeverityInfo, "ok")
	el.Record("beta", SeverityError, "fail")
	el.Record("alpha", SeverityWarn, "slow")
	got := el.FilterByJob("alpha")
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestEventLog_FiltersBySeverity(t *testing.T) {
	el := newTestEventLog()
	el.Record("job", SeverityInfo, "a")
	el.Record("job", SeverityError, "b")
	el.Record("job", SeverityError, "c")
	got := el.FilterBySeverity(SeverityError)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestEventLogHandler_ContentType(t *testing.T) {
	el := newTestEventLog()
	h := EventLogHandler(el)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/events", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestEventLogHandler_FiltersByJob(t *testing.T) {
	el := newTestEventLog()
	el.Record("alpha", SeverityInfo, "ok")
	el.Record("beta", SeverityWarn, "slow")
	h := EventLogHandler(el)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/events?job=alpha", nil))
	var out []EventEntry
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Job != "alpha" {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestEventLogHandler_EmptyReturnsEmptySlice(t *testing.T) {
	el := newTestEventLog()
	h := EventLogHandler(el)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/events", nil))
	var out []EventEntry
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out == nil || len(out) != 0 {
		t.Errorf("expected empty slice, got %v", out)
	}
}
