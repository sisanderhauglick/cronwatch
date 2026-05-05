package schedule_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/schedule"
)

func mustTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestNextRun_Valid(t *testing.T) {
	// Every hour at minute 0
	next, err := schedule.NextRun("0 * * * *", mustTime("2024-01-01 12:30"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := mustTime("2024-01-01 13:00")
	if !next.Equal(want) {
		t.Errorf("NextRun = %v, want %v", next, want)
	}
}

func TestNextRun_InvalidExpr(t *testing.T) {
	_, err := schedule.NextRun("not-a-cron", time.Now())
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}

func TestPrevRun_Valid(t *testing.T) {
	// Every hour at minute 0; prev before 13:45 should be 13:00
	prev, err := schedule.PrevRun("0 * * * *", mustTime("2024-01-01 13:45"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := mustTime("2024-01-01 13:00")
	if !prev.Equal(want) {
		t.Errorf("PrevRun = %v, want %v", prev, want)
	}
}

func TestIsDue_True(t *testing.T) {
	start := mustTime("2024-01-01 12:59")
	end := mustTime("2024-01-01 13:01")
	due, err := schedule.IsDue("0 * * * *", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !due {
		t.Error("expected job to be due, got false")
	}
}

func TestIsDue_False(t *testing.T) {
	start := mustTime("2024-01-01 13:01")
	end := mustTime("2024-01-01 13:30")
	due, err := schedule.IsDue("0 * * * *", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if due {
		t.Error("expected job not to be due, got true")
	}
}
