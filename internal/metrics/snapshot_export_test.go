package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestExporter(t *testing.T) (*SnapshotExporter, *Collector) {
	t.Helper()
	c := NewCollector(DefaultRetentionPolicy())
	return NewSnapshotExporter(c), c
}

func TestExport_EmptyJSON(t *testing.T) {
	ex, _ := newTestExporter(t)
	var buf bytes.Buffer
	if err := ex.Export(&buf, ExportJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var rows []ExportRow
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestExport_JSONContainsSnapshot(t *testing.T) {
	ex, c := newTestExporter(t)
	snap := Snapshot{Job: "backup", Timestamp: time.Now(), Seen: 3, Missed: 1, Failed: 0}
	c.Collect(snap)

	var buf bytes.Buffer
	if err := ex.Export(&buf, ExportJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var rows []ExportRow
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(rows) != 1 || rows[0].Job != "backup" || rows[0].Seen != 3 {
		t.Errorf("unexpected rows: %+v", rows)
	}
}

func TestExport_CSVHeader(t *testing.T) {
	ex, _ := newTestExporter(t)
	var buf bytes.Buffer
	if err := ex.Export(&buf, ExportCSV); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if lines[0] != "job,timestamp,seen,missed,failed" {
		t.Errorf("unexpected CSV header: %q", lines[0])
	}
}

func TestExport_CSVRow(t *testing.T) {
	ex, c := newTestExporter(t)
	c.Collect(Snapshot{Job: "sync", Timestamp: time.Now(), Seen: 5, Missed: 0, Failed: 2})

	var buf bytes.Buffer
	if err := ex.Export(&buf, ExportCSV); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "sync") {
		t.Errorf("expected job name in CSV output")
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	ex, _ := newTestExporter(t)
	var buf bytes.Buffer
	if err := ex.Export(&buf, ExportFormat("xml")); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestExportHandler_ContentTypeJSON(t *testing.T) {
	_, c := newTestExporter(t)
	h := NewExportHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/export", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestExportHandler_ContentTypeCSV(t *testing.T) {
	_, c := newTestExporter(t)
	h := NewExportHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/export?format=csv", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %q", ct)
	}
}

func TestExportHandler_BadFormat(t *testing.T) {
	_, c := newTestExporter(t)
	h := NewExportHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/export?format=xml", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
