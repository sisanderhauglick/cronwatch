package schedule_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/schedule"
)

func TestFindMissed_NoneWhenSeen(t *testing.T) {
	start := mustTime("2024-01-01 12:00")
	end := mustTime("2024-01-01 14:05")
	grace := 5 * time.Minute

	// Hourly job seen at both 13:00 and 14:00
	lastSeen := []time.Time{
		mustTime("2024-01-01 13:00"),
		mustTime("2024-01-01 14:00"),
	}

	missed, err := schedule.FindMissed("0 * * * *", start, end, lastSeen, grace)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missed) != 0 {
		t.Errorf("expected 0 missed, got %d: %v", len(missed), missed)
	}
}

func TestFindMissed_DetectsMissed(t *testing.T) {
	start := mustTime("2024-01-01 12:00")
	end := mustTime("2024-01-01 14:10")
	grace := 5 * time.Minute

	// Only 13:00 seen; 14:00 missed
	lastSeen := []time.Time{
		mustTime("2024-01-01 13:00"),
	}

	missed, err := schedule.FindMissed("0 * * * *", start, end, lastSeen, grace)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missed) != 1 {
		t.Fatalf("expected 1 missed, got %d", len(missed))
	}
	want := mustTime("2024-01-01 14:00")
	if !missed[0].ScheduledAt.Equal(want) {
		t.Errorf("missed ScheduledAt = %v, want %v", missed[0].ScheduledAt, want)
	}
}

func TestFindMissed_GraceNotElapsed(t *testing.T) {
	start := mustTime("2024-01-01 12:00")
	// end is only 2 minutes after the 14:00 tick, grace is 5 min
	end := mustTime("2024-01-01 14:02")
	grace := 5 * time.Minute

	missed, err := schedule.FindMissed("0 * * * *", start, end, nil, grace)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 14:00 grace hasn't elapsed; only 13:00 should be missed
	if len(missed) != 1 {
		t.Fatalf("expected 1 missed (13:00), got %d: %v", len(missed), missed)
	}
	want := mustTime("2024-01-01 13:00")
	if !missed[0].ScheduledAt.Equal(want) {
		t.Errorf("missed ScheduledAt = %v, want %v", missed[0].ScheduledAt, want)
	}
}
