package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelMissed Level = "missed"
	LevelFailed Level = "failed"
)

// Alert holds the data for a single alert event.
type Alert struct {
	JobName   string    `json:"job_name"`
	Level     Level     `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Notifier sends alerts to a configured destination.
type Notifier interface {
	Send(a Alert) error
}

// WebhookNotifier posts alerts as JSON to an HTTP endpoint.
type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a default HTTP client.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send marshals the alert and POSTs it to the webhook URL.
func (w *WebhookNotifier) Send(a Alert) error {
	payload, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("alert: marshal: %w", err)
	}

	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("alert: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: unexpected status %d from webhook", resp.StatusCode)
	}
	return nil
}
