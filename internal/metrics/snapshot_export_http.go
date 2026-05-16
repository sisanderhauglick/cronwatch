package metrics

import (
	"net/http"
)

// ExportHandler serves collected snapshots as a downloadable file.
// Query parameter "format" selects "json" (default) or "csv".
type ExportHandler struct {
	exporter *SnapshotExporter
}

// NewExportHandler creates an HTTP handler backed by the given Collector.
func NewExportHandler(c *Collector) *ExportHandler {
	return &ExportHandler{exporter: NewSnapshotExporter(c)}
}

func (h *ExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt := ExportFormat(r.URL.Query().Get("format"))
	if fmt == "" {
		fmt = ExportJSON
	}

	switch fmt {
	case ExportCSV:
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="snapshots.csv"`)
	case ExportJSON:
		w.Header().Set("Content-Type", "application/json")
	default:
		http.Error(w, "unsupported format", http.StatusBadRequest)
		return
	}

	if err := h.exporter.Export(w, fmt); err != nil {
		// Headers already (partially) written; log the error via plain text.
		http.Error(w, "export failed: "+err.Error(), http.StatusInternalServerError)
	}
}
