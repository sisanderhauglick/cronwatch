package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/alert"
)

func makeAlert(level alert.Level) alert.Alert {
	return alert.Alert{
		JobName:   "backup",
		Level:     level,
		Message:   "job did not run",
		Timestamp: time.Now().UTC(),
	}
}

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received alert.Alert

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := alert.NewWebhookNotifier(srv.URL)
	a := makeAlert(alert.LevelMissed)

	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != a.JobName {
		t.Errorf("job name: got %q, want %q", received.JobName, a.JobName)
	}
	if received.Level != alert.LevelMissed {
		t.Errorf("level: got %q, want %q", received.Level, alert.LevelMissed)
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := alert.NewWebhookNotifier(srv.URL)
	if err := n.Send(makeAlert(alert.LevelFailed)); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookNotifier_Send_BadURL(t *testing.T) {
	n := alert.NewWebhookNotifier("http://127.0.0.1:0/nowhere")
	if err := n.Send(makeAlert(alert.LevelMissed)); err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}
