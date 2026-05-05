// Package main is the entry point for the cronwatch daemon.
//
// cronwatch monitors cron job execution by periodically checking whether
// scheduled jobs have run within their expected windows. When a job is
// missed or fails, it dispatches alerts through the configured notifiers
// (webhook, email, and/or log).
//
// Usage:
//
//	cronwatch [-config <path>]
//
// Flags:
//
//	-config   Path to the YAML configuration file (default: cronwatch.yaml)
//
// cronwatch listens for SIGINT and SIGTERM to shut down gracefully.
package main
