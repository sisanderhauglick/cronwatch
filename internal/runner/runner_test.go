package runner

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
	"github.com/yourorg/cronwatch/internal/config"
	"github.com/yourorg/cronwatch/internal/state"
)

func newTestRunner(t *testing.T, webhookURL string) (*Runner, *state.Updater) {
	t.Helper()
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "heartbeat", Schedule: "* * * * *", Grace: "2m"},
		},
		WebhookURL: webhookURL,
	}
	st, err := state.New("")
	if err != nil {
		t.Fatalf("state.New: %v", err)
	}
	updater := state.NewUpdater(st)
	dispatcher := alert.NewDispatcher(alert.NewWebhookNotifier(webhookURL))
	return New(cfg, updater, dispatcher, 50*time.Millisecond), updater
}

func TestRunner_StopExits(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r, _ := newTestRunner(t, ts.URL)
	done := make(chan struct{})
	go func() {
		r.Start()
		close(done)
	}()
	time.Sleep(60 * time.Millisecond)
	r.Stop()
	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runner did not stop in time")
	}
}

func TestRunner_Tick_SendsAlertForMissedJob(t *testing.T) {
	var callCount int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		atomic.AddInt64(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r, _ := newTestRunner(t, ts.URL)
	// Tick far in the future so the job looks missed.
	r.tick(time.Now().Add(24 * time.Hour))

	if atomic.LoadInt64(&callCount) == 0 {
		t.Error("expected at least one alert to be dispatched for missed job")
	}
}

func TestRunner_Tick_NoAlertWhenRecentlySeen(t *testing.T) {
	var callCount int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r, updater := newTestRunner(t, ts.URL)
	job := r.cfg.Jobs[0]
	now := time.Now()
	if err := updater.RecordSeen(job.Name, now); err != nil {
		t.Fatalf("RecordSeen: %v", err)
	}
	// Tick just slightly ahead — job was seen recently, no alert expected.
	r.tick(now.Add(30 * time.Second))

	if c := atomic.LoadInt64(&callCount); c != 0 {
		t.Errorf("expected 0 alerts, got %d", c)
	}
	_ = os.Remove("") // no-op, state is in-memory
}
