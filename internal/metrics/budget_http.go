package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// BudgetHandler returns an HTTP handler that serves error budget analysis.
func BudgetHandler(a *BudgetAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := a.Analyze(time.Now())
		if results == nil {
			results = []ErrorBudget{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results) //nolint:errcheck
	}
}
