package main

import (
	"testing"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/config"
)

func TestBuildNotifier_LogOnly(t *testing.T) {
	cfg := &config.Config{
		Alerts: config.AlertsConfig{},
	}
	n := buildNotifier(cfg)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	// With no webhook or email, should return a single log notifier
	_, ok := n.(*alert.LogNotifier)
	if !ok {
		t.Errorf("expected *alert.LogNotifier, got %T", n)
	}
}

func TestBuildNotifier_MultiWhenWebhook(t *testing.T) {
	cfg := &config.Config{
		Alerts: config.AlertsConfig{
			WebhookURL: "http://example.com/hook",
		},
	}
	n := buildNotifier(cfg)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	// With webhook + log fallback, should be a MultiNotifier
	_, ok := n.(*alert.MultiNotifier)
	if !ok {
		t.Errorf("expected *alert.MultiNotifier, got %T", n)
	}
}

func TestBuildNotifier_MultiWhenEmail(t *testing.T) {
	cfg := &config.Config{
		Alerts: config.AlertsConfig{
			Email: config.EmailConfig{
				To:   "ops@example.com",
				From: "cronwatch@example.com",
				Host: "localhost",
				Port: 25,
			},
		},
	}
	n := buildNotifier(cfg)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	_, ok := n.(*alert.MultiNotifier)
	if !ok {
		t.Errorf("expected *alert.MultiNotifier, got %T", n)
	}
}
