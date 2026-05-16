package metrics

import (
	"math"
	"sync"
	"time"
)

// JitterSample holds a single observed schedule deviation.
type JitterSample struct {
	Job       string
	Expected  time.Time
	Actual    time.Time
	Delta     time.Duration
	Recorded  time.Time
}

// JitterStats summarises deviation for one job.
type JitterStats struct {
	Job    string        `json:"job"`
	Count  int           `json:"count"`
	MeanMs float64       `json:"mean_ms"`
	StdMs  float64       `json:"std_ms"`
	MaxMs  float64       `json:"max_ms"`
}

// JitterTracker measures how far cron runs deviate from their scheduled time.
type JitterTracker struct {
	mu      sync.Mutex
	samples []JitterSample
	maxAge  time.Duration
}

// NewJitterTracker creates a tracker that keeps samples within maxAge.
func NewJitterTracker(maxAge time.Duration) *JitterTracker {
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}
	return &JitterTracker{maxAge: maxAge}
}

// Record stores an observed execution time against its expected time.
func (j *JitterTracker) Record(job string, expected, actual time.Time) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.samples = append(j.samples, JitterSample{
		Job:      job,
		Expected: expected,
		Actual:   actual,
		Delta:    actual.Sub(expected),
		Recorded: time.Now(),
	})
}

// Stats returns jitter statistics for each job within the retention window.
func (j *JitterTracker) Stats(now time.Time) []JitterStats {
	j.mu.Lock()
	defer j.mu.Unlock()

	cutoff := now.Add(-j.maxAge)
	byJob := map[string][]float64{}
	for _, s := range j.samples {
		if s.Recorded.Before(cutoff) {
			continue
		}
		ms := float64(s.Delta.Abs()) / float64(time.Millisecond)
		byJob[s.Job] = append(byJob[s.Job], ms)
	}

	var out []JitterStats
	for job, vals := range byJob {
		var sum, max float64
		for _, v := range vals {
			sum += v
			if v > max {
				max = v
			}
		}
		mean := sum / float64(len(vals))
		var variance float64
		for _, v := range vals {
			d := v - mean
			variance += d * d
		}
		if len(vals) > 1 {
			variance /= float64(len(vals))
		}
		out = append(out, JitterStats{
			Job:    job,
			Count:  len(vals),
			MeanMs: math.Round(mean*100) / 100,
			StdMs:  math.Round(math.Sqrt(variance)*100) / 100,
			MaxMs:  max,
		})
	}
	return out
}
