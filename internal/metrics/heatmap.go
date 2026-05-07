package metrics

import (
	"math"
	"time"
)

// HourlyBucket holds aggregated counts for a single hour-of-week slot.
type HourlyBucket struct {
	DayOfWeek time.Weekday `json:"day_of_week"`
	Hour      int         `json:"hour"`
	Total     int         `json:"total"`
	Failed    int         `json:"failed"`
	Missed    int         `json:"missed"`
	FailRate  float64     `json:"fail_rate"`
}

// HeatmapAnalyzer builds a failure heatmap bucketed by day-of-week and hour.
type HeatmapAnalyzer struct {
	collector *Collector
	window    time.Duration
}

// NewHeatmapAnalyzer returns a HeatmapAnalyzer using the given collector and
// look-back window.
func NewHeatmapAnalyzer(c *Collector, window time.Duration) *HeatmapAnalyzer {
	return &HeatmapAnalyzer{collector: c, window: window}
}

// Analyze returns HourlyBuckets for all (day, hour) slots that have at least
// one recorded run within the configured window.
func (h *HeatmapAnalyzer) Analyze(job string, now time.Time) []HourlyBucket {
	cutoff := now.Add(-h.window)
	snapshots := h.collector.All(job)

	// key: day*24 + hour
	type bucket struct {
		total, failed, missed int
	}
	buckets := make(map[int]*bucket)

	for _, snap := range snapshots {
		if snap.Timestamp.Before(cutoff) {
			continue
		}
		key := int(snap.Timestamp.Weekday())*24 + snap.Timestamp.Hour()
		if _, ok := buckets[key]; !ok {
			buckets[key] = &bucket{}
		}
		b := buckets[key]
		b.total++
		if snap.Failed > 0 {
			b.failed++
		}
		if snap.Missed > 0 {
			b.missed++
		}
	}

	result := make([]HourlyBucket, 0, len(buckets))
	for key, b := range buckets {
		day := time.Weekday(key / 24)
		hour := key % 24
		failRate := 0.0
		if b.total > 0 {
			failRate = math.Round(float64(b.failed+b.missed)/float64(b.total)*1000) / 1000
		}
		result = append(result, HourlyBucket{
			DayOfWeek: day,
			Hour:      hour,
			Total:     b.total,
			Failed:    b.failed,
			Missed:    b.missed,
			FailRate:  failRate,
		})
	}
	return result
}
