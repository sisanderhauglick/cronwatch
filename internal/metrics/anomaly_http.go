package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// anomalyResponse is the JSON shape returned by AnomalyHandler.
type anomalyResponse struct {
	Job        string    `json:"job"`
	Kind       string    `json:"kind"`
	Score      float64   `json:"score"`
	DetectedAt time.Time `json:"detected_at"`
}

// AnomalyHandler returns an HTTP handler that exposes detected anomalies as JSON.
func AnomalyHandler(det *AnomalyDetector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		reports := det.Detect(now)

		out := make([]anomalyResponse, 0, len(reports))
		for _, rep := range reports {
			out = append(out, anomalyResponse{
				Job:        rep.Job,
				Kind:       rep.Kind,
				Score:      rep.Score,
				DetectedAt: rep.DetectedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
