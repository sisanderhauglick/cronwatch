// Package runner implements the main check loop for cronwatch.
//
// A Runner is constructed with a parsed Config, a state Updater, and an alert
// Dispatcher. It ticks on a configurable interval, calling Updater.CheckMissed
// for every configured job and forwarding any resulting alerts to the
// Dispatcher.
//
// Typical usage:
//
//	r := runner.New(cfg, updater, dispatcher, time.Minute)
//	go r.Start()   // blocks until Stop is called
//	...
//	r.Stop()
package runner
