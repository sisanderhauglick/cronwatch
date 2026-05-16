package metrics

import (
	"testing"
	"time"
)

func newTestJitterTracker() *JitterTracker {
	return NewJitterTracker(1 * time.Hour)
}

func TestJitter_EmptyReturnsNoStats(t *testing.T) {
	tr := newTestJitterTracker()
	stats := tr.Stats(time.Now())
	if len(stats) != 0 {
		t.Fatalf("expected empty stats, got %d", len(stats))
	}
}

func TestJitter_SingleSampleMeanEqualsAbsDelta(t *testing.T) {
	tr := newTestJitterTracker()
	now := time.Now()
	expected := now.Add(-5 * time.Second)
	actual := now
	tr.Record("backup", expected, actual)

	stats := tr.Stats(now)
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	s := stats[0]
	if s.Job != "backup" {
		t.Errorf("unexpected job name %q", s.Job)
	}
	if s.Count != 1 {
		t.Errorf("expected count 1, got %d", s.Count)
	}
	wantMs := 5000.0
	if s.MeanMs != wantMs {
		t.Errorf("expected mean %.2f ms, got %.2f", wantMs, s.MeanMs)
	}
	if s.StdMs != 0 {
		t.Errorf("expected std 0 for single sample, got %.2f", s.StdMs)
	}
}

func TestJitter_MultipleJobsTrackedSeparately(t *testing.T) {
	tr := newTestJitterTracker()
	now := time.Now()
	tr.Record("alpha", now.Add(-1*time.Second), now)
	tr.Record("beta", now.Add(-3*time.Second), now)

	stats := tr.Stats(now)
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	byJob := map[string]JitterStats{}
	for _, s := range stats {
		byJob[s.Job] = s
	}
	if byJob["alpha"].MeanMs != 1000 {
		t.Errorf("alpha mean wrong: %.2f", byJob["alpha"].MeanMs)
	}
	if byJob["beta"].MeanMs != 3000 {
		t.Errorf("beta mean wrong: %.2f", byJob["beta"].MeanMs)
	}
}

func TestJitter_ExcludesOldSamples(t *testing.T) {
	tr := newTestJitterTracker()
	now := time.Now()
	old := now.Add(-2 * time.Hour)

	// Manually insert an old sample.
	tr.mu.Lock()
	tr.samples = append(tr.samples, JitterSample{
		Job:      "cleanup",
		Expected: old,
		Actual:   old.Add(2 * time.Second),
		Delta:    2 * time.Second,
		Recorded: old,
	})
	tr.mu.Unlock()

	stats := tr.Stats(now)
	if len(stats) != 0 {
		t.Fatalf("expected old sample to be excluded, got %d stats", len(stats))
	}
}

func TestJitter_MaxReflectsLargestDelta(t *testing.T) {
	tr := newTestJitterTracker()
	now := time.Now()
	tr.Record("sync", now.Add(-1*time.Second), now)
	tr.Record("sync", now.Add(-10*time.Second), now)
	tr.Record("sync", now.Add(-4*time.Second), now)

	stats := tr.Stats(now)
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].MaxMs != 10000 {
		t.Errorf("expected max 10000 ms, got %.2f", stats[0].MaxMs)
	}
	if stats[0].Count != 3 {
		t.Errorf("expected count 3, got %d", stats[0].Count)
	}
}
