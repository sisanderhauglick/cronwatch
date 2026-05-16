package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestSuppressionManager() *SuppressionManager {
	return NewSuppressionManager()
}

func TestSuppression_NotSuppressedByDefault(t *testing.T) {
	sm := newTestSuppressionManager()
	if sm.IsSuppressed("job1", time.Now()) {
		t.Fatal("expected not suppressed")
	}
}

func TestSuppression_SuppressedWithinWindow(t *testing.T) {
	sm := newTestSuppressionManager()
	now := time.Now()
	sm.Add(SuppressionRule{
		JobName: "job1",
		Start:   now.Add(-time.Minute),
		End:     now.Add(time.Minute),
		Reason:  "maintenance",
	})
	if !sm.IsSuppressed("job1", now) {
		t.Fatal("expected suppressed")
	}
}

func TestSuppression_NotSuppressedAfterExpiry(t *testing.T) {
	sm := newTestSuppressionManager()
	now := time.Now()
	sm.Add(SuppressionRule{
		JobName: "job1",
		Start:   now.Add(-2 * time.Minute),
		End:     now.Add(-time.Minute),
		Reason:  "old",
	})
	if sm.IsSuppressed("job1", now) {
		t.Fatal("expected not suppressed after expiry")
	}
}

func TestSuppression_PruneRemovesExpired(t *testing.T) {
	sm := newTestSuppressionManager()
	now := time.Now()
	sm.Add(SuppressionRule{JobName: "a", Start: now.Add(-2 * time.Minute), End: now.Add(-time.Minute)})
	sm.Add(SuppressionRule{JobName: "b", Start: now.Add(-time.Minute), End: now.Add(time.Minute)})
	sm.Prune(now)
	rules := sm.Active()
	if len(rules) != 1 || rules[0].JobName != "b" {
		t.Fatalf("expected 1 active rule for 'b', got %+v", rules)
	}
}

func TestSuppressionHandler_GetContentType(t *testing.T) {
	sm := newTestSuppressionManager()
	h := NewSuppressionHandler(sm)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/suppressions", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestSuppressionHandler_PostAddsRule(t *testing.T) {
	sm := newTestSuppressionManager()
	h := NewSuppressionHandler(sm)

	now := time.Now().UTC().Truncate(time.Second)
	body, _ := json.Marshal(suppressionRuleRequest{
		JobName: "job1",
		Start:   now.Format(time.RFC3339),
		End:     now.Add(time.Hour).Format(time.RFC3339),
		Reason:  "deploy",
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/suppressions", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if len(sm.Active()) != 1 {
		t.Fatal("expected 1 rule after POST")
	}
}

func TestSuppressionHandler_PostMissingJobName(t *testing.T) {
	sm := newTestSuppressionManager()
	h := NewSuppressionHandler(sm)
	now := time.Now().UTC()
	body, _ := json.Marshal(suppressionRuleRequest{
		Start: now.Format(time.RFC3339),
		End:   now.Add(time.Hour).Format(time.RFC3339),
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/suppressions", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
