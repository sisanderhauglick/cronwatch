package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestSilenceManager(now time.Time) *SilenceManager {
	sm := NewSilenceManager()
	sm.nowFunc = func() time.Time { return now }
	return sm
}

func TestSilenceManager_NotSilencedByDefault(t *testing.T) {
	sm := newTestSilenceManager(time.Now())
	silenced, _ := sm.IsSilenced("backup")
	if silenced {
		t.Fatal("expected job not to be silenced")
	}
}

func TestSilenceManager_SilencedWithinWindow(t *testing.T) {
	now := time.Now()
	sm := newTestSilenceManager(now)
	sm.Add(SilenceRule{
		JobName:   "backup",
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		Reason:    "maintenance",
	})
	silenced, reason := sm.IsSilenced("backup")
	if !silenced {
		t.Fatal("expected job to be silenced")
	}
	if reason != "maintenance" {
		t.Fatalf("expected reason 'maintenance', got %q", reason)
	}
}

func TestSilenceManager_NotSilencedAfterExpiry(t *testing.T) {
	now := time.Now()
	sm := newTestSilenceManager(now)
	sm.Add(SilenceRule{
		JobName:   "backup",
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now.Add(-time.Minute),
		Reason:    "old window",
	})
	silenced, _ := sm.IsSilenced("backup")
	if silenced {
		t.Fatal("expected job not to be silenced after expiry")
	}
}

func TestSilenceManager_PruneRemovesExpired(t *testing.T) {
	now := time.Now()
	sm := newTestSilenceManager(now)
	sm.Add(SilenceRule{JobName: "a", StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Minute)})
	sm.Add(SilenceRule{JobName: "b", StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour)})
	sm.Prune()
	rules := sm.List()
	if len(rules) != 1 || rules[0].JobName != "b" {
		t.Fatalf("expected 1 active rule for 'b', got %v", rules)
	}
}

func TestSilenceHandler_PostAndGet(t *testing.T) {
	sm := NewSilenceManager()
	handler := SilenceHandler(sm)

	body, _ := json.Marshal(map[string]string{
		"job_name": "cleanup",
		"duration": "2h",
		"reason":   "deploy",
	})
	req := httptest.NewRequest(http.MethodPost, "/silence", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/silence", nil)
	w2 := httptest.NewRecorder()
	handler(w2, req2)
	var out []silenceResponse
	json.NewDecoder(w2.Body).Decode(&out)
	if len(out) != 1 || out[0].JobName != "cleanup" {
		t.Fatalf("expected 1 silence rule for 'cleanup', got %v", out)
	}
}

func TestSilenceHandler_InvalidDuration(t *testing.T) {
	sm := NewSilenceManager()
	handler := SilenceHandler(sm)
	body, _ := json.Marshal(map[string]string{"job_name": "x", "duration": "bad"})
	req := httptest.NewRequest(http.MethodPost, "/silence", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
