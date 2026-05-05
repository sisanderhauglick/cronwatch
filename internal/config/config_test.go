package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `
log_level: info
alerts:
  email: ops@example.com
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout: 30m
    command: /usr/local/bin/backup.sh
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log_level=info, got %q", cfg.LogLevel)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name=backup, got %q", cfg.Jobs[0].Name)
	}
	if cfg.Jobs[0].Timeout != 30*time.Minute {
		t.Errorf("expected timeout=30m, got %v", cfg.Jobs[0].Timeout)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `log_level: debug\njobs: []\n`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs list")
	}
}

func TestLoad_DuplicateJobName(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: cleanup
    schedule: "*/5 * * * *"
  - name: cleanup
    schedule: "*/10 * * * *"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for duplicate job name")
	}
}

func TestLoad_MissingSchedule(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: noschedule
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing schedule")
	}
}
