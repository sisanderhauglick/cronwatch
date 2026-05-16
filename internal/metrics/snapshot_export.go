package metrics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportFormat controls the serialisation format used by SnapshotExporter.
type ExportFormat string

const (
	ExportJSON ExportFormat = "json"
	ExportCSV  ExportFormat = "csv"
)

// ExportRow is a flattened representation of a single job snapshot.
type ExportRow struct {
	Job       string    `json:"job"`
	Timestamp time.Time `json:"timestamp"`
	Seen      int64     `json:"seen"`
	Missed    int64     `json:"missed"`
	Failed    int64     `json:"failed"`
}

// SnapshotExporter writes collected snapshots to an io.Writer in the
// requested format.
type SnapshotExporter struct {
	collector *Collector
}

// NewSnapshotExporter creates an exporter backed by the given Collector.
func NewSnapshotExporter(c *Collector) *SnapshotExporter {
	return &SnapshotExporter{collector: c}
}

// Export writes all snapshots held by the collector to w.
func (e *SnapshotExporter) Export(w io.Writer, format ExportFormat) error {
	rows := e.buildRows()
	switch format {
	case ExportCSV:
		return e.writeCSV(w, rows)
	case ExportJSON:
		return json.NewEncoder(w).Encode(rows)
	default:
		return fmt.Errorf("unsupported export format: %q", format)
	}
}

func (e *SnapshotExporter) buildRows() []ExportRow {
	all := e.collector.All()
	rows := make([]ExportRow, 0, len(all))
	for _, snap := range all {
		rows = append(rows, ExportRow{
			Job:       snap.Job,
			Timestamp: snap.Timestamp,
			Seen:      snap.Seen,
			Missed:    snap.Missed,
			Failed:    snap.Failed,
		})
	}
	return rows
}

func (e *SnapshotExporter) writeCSV(w io.Writer, rows []ExportRow) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"job", "timestamp", "seen", "missed", "failed"}); err != nil {
		return err
	}
	for _, r := range rows {
		record := []string{
			r.Job,
			r.Timestamp.UTC().Format(time.RFC3339),
			fmt.Sprintf("%d", r.Seen),
			fmt.Sprintf("%d", r.Missed),
			fmt.Sprintf("%d", r.Failed),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
