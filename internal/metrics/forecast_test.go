package metrics

import (
	"testing"
	"time"
)

func newTestForecastAnalyzer() (*ForecastAnalyzer, *Collector) {
	c := NewCollector(DefaultRetentionPolicy())
	return NewForecastAnalyzer(c, 24*time.Hour), c
}

func TestForecast_EmptyCollector(t *testing.T) {
	fa, _ := newTestForecastAnalyzer()
	results := fa.Predict(time.Now())
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestForecast_ZeroProbWhenNoFailures(t *testing.T) {
	fa, c := newTestForecastAnalyzer()
	now := time.Now()

	snap := Snapshot{
		CollectedAt: now.Add(-1 * time.Hour),
		Jobs: map[string]JobStats{
			"backup": {Seen: 10, Failed: 0, Missed: 0},
		},
	}
	c.store(snap)

	results := fa.Predict(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].FailureProb != 0.0 {
		t.Errorf("expected 0 failure prob, got %f", results[0].FailureProb)
	}
}

func TestForecast_DetectsHighFailureProbability(t *testing.T) {
	fa, c := newTestForecastAnalyzer()
	now := time.Now()

	for i := 0; i < 5; i++ {
		snap := Snapshot{
			CollectedAt: now.Add(-time.Duration(5-i) * time.Hour),
			Jobs: map[string]JobStats{
				"mailer": {Seen: 1, Failed: 9, Missed: 0},
			},
		}
		c.store(snap)
	}

	results := fa.Predict(now)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].FailureProb <= 0.5 {
		t.Errorf("expected high failure prob, got %f", results[0].FailureProb)
	}
}

func TestForecast_ExcludesSnapshotsOutsideWindow(t *testing.T) {
	fa, c := newTestForecastAnalyzer()
	now := time.Now()

	// Outside window
	c.store(Snapshot{
		CollectedAt: now.Add(-48 * time.Hour),
		Jobs: map[string]JobStats{
			"cleaner": {Seen: 0, Failed: 100},
		},
	})
	// Inside window
	c.store(Snapshot{
		CollectedAt: now.Add(-1 * time.Hour),
		Jobs: map[string]JobStats{
			"cleaner": {Seen: 10, Failed: 0},
		},
	})

	results := fa.Predict(now)
	if len(results) == 0 {
		t.Fatal("expected a result")
	}
	if results[0].FailureProb != 0.0 {
		t.Errorf("expected 0 prob (old failures excluded), got %f", results[0].FailureProb)
	}
}

func TestForecast_SampleWindowCount(t *testing.T) {
	fa, c := newTestForecastAnalyzer()
	now := time.Now()

	for i := 0; i < 3; i++ {
		c.store(Snapshot{
			CollectedAt: now.Add(-time.Duration(i+1) * time.Hour),
			Jobs: map[string]JobStats{
				"sync": {Seen: 5, Failed: 1},
			},
		})
	}

	results := fa.Predict(now)
	if len(results) == 0 {
		t.Fatal("expected a result")
	}
	if results[0].SampleWindow != 3 {
		t.Errorf("expected 3 samples, got %d", results[0].SampleWindow)
	}
}
