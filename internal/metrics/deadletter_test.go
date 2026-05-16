package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestDeadLetterQueue() *DeadLetterQueue {
	return NewDeadLetterQueue(5)
}

func TestDeadLetter_EmptyQueue(t *testing.T) {
	q := newTestDeadLetterQueue()
	if got := q.Len(); got != 0 {
		t.Fatalf("expected 0 entries, got %d", got)
	}
}

func TestDeadLetter_PushAndAll(t *testing.T) {
	q := newTestDeadLetterQueue()
	q.Push("backup", "webhook failed", `{"job":"backup"}`)
	entries := q.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Job != "backup" {
		t.Errorf("expected job=backup, got %s", entries[0].Job)
	}
}

func TestDeadLetter_IncreasesAttempts(t *testing.T) {
	q := newTestDeadLetterQueue()
	q.Push("sync", "timeout", "payload")
	q.Push("sync", "timeout", "payload")
	entries := q.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 deduplicated entry, got %d", len(entries))
	}
	if entries[0].Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", entries[0].Attempts)
	}
}

func TestDeadLetter_EvictsOldestWhenFull(t *testing.T) {
	q := NewDeadLetterQueue(3)
	q.Push("a", "r", "")
	q.Push("b", "r", "")
	q.Push("c", "r", "")
	q.Push("d", "r", "")
	entries := q.All()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(entries))
	}
	if entries[0].Job != "b" {
		t.Errorf("expected oldest evicted, first entry should be b, got %s", entries[0].Job)
	}
}

func TestDeadLetter_Remove(t *testing.T) {
	q := newTestDeadLetterQueue()
	q.Push("cleanup", "smtp error", "")
	if !q.Remove("cleanup", "smtp error") {
		t.Fatal("expected Remove to return true")
	}
	if q.Len() != 0 {
		t.Errorf("expected 0 entries after remove, got %d", q.Len())
	}
}

func TestDeadLetterHandler_GetReturnsEntries(t *testing.T) {
	q := newTestDeadLetterQueue()
	q.Push("job1", "failed", "p")
	h := NewDeadLetterHandler(q)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/deadletter", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries []DeadLetterEntry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestDeadLetterHandler_ContentType(t *testing.T) {
	h := NewDeadLetterHandler(newTestDeadLetterQueue())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/deadletter", nil))
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestDeadLetterHandler_DeleteRemovesEntry(t *testing.T) {
	q := newTestDeadLetterQueue()
	q.Push("myjob", "reason", "")
	h := NewDeadLetterHandler(q)
	req := httptest.NewRequest(http.MethodDelete, "/deadletter?job=myjob&reason=reason", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if q.Len() != 0 {
		t.Errorf("expected queue empty after delete")
	}
}
