package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutes_RollupEndpoint(t *testing.T) {
	reg := New()
	coll := NewCollector(DefaultRetentionPolicy())
	agg := NewAggregator(coll)
	tracker := NewAlertTracker(50)
	rollup := NewRollupAnalyzer(coll)

	mux := http.NewServeMux()
	RegisterRoutes(mux, reg, coll, agg, tracker)
	mux.Handle("/metrics/rollup", RollupHandler(rollup))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/rollup", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
