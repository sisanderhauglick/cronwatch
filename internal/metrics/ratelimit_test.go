package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAlertRateLimiter_FirstAlwaysAllowed(t *testing.T) {
	rl := NewAlertRateLimiter(5 * time.Minute)
	now := time.Now()
	if !rl.Allow("job1", now) {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestAlertRateLimiter_SuppressedWithinCooldown(t *testing.T) {
	rl := NewAlertRateLimiter(5 * time.Minute)
	now := time.Now()
	rl.Allow("job1", now)
	if rl.Allow("job1", now.Add(1*time.Minute)) {
		t.Fatal("expected alert to be suppressed within cooldown")
	}
}

func TestAlertRateLimiter_AllowedAfterCooldown(t *testing.T) {
	rl := NewAlertRateLimiter(5 * time.Minute)
	now := time.Now()
	rl.Allow("job1", now)
	if !rl.Allow("job1", now.Add(6*time.Minute)) {
		t.Fatal("expected alert to be allowed after cooldown expires")
	}
}

func TestAlertRateLimiter_Reset(t *testing.T) {
	rl := NewAlertRateLimiter(5 * time.Minute)
	now := time.Now()
	rl.Allow("job1", now)
	rl.Reset("job1")
	if !rl.Allow("job1", now.Add(10*time.Second)) {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestAlertRateLimiter_Stats(t *testing.T) {
	rl := NewAlertRateLimiter(5 * time.Minute)
	now := time.Now()
	rl.Allow("job1", now)

	entry, ok := rl.Stats("job1")
	if !ok {
		t.Fatal("expected stats to exist")
	}
	if entry.Count != 1 {
		t.Fatalf("expected count 1, got %d", entry.Count)
	}

	_, ok = rl.Stats("missing")
	if ok {
		t.Fatal("expected no stats for unknown job")
	}
}

func TestRateLimitHandler_ContentTypeAndBody(t *testing.T) {
	reg := New()
	reg.RecordSeen("jobA")

	rl := NewAlertRateLimiter(10 * time.Minute)
	now := time.Now()
	rl.Allow("jobA", now)

	h := RateLimitHandler(reg, rl)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ratelimit", nil))

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var results []rateLimitStatus
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Job != "jobA" {
		t.Fatalf("expected job jobA, got %s", results[0].Job)
	}
	if results[0].AlertCount != 1 {
		t.Fatalf("expected alert count 1, got %d", results[0].AlertCount)
	}
}
