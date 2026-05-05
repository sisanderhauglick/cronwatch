package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/runner"
	"github.com/example/cronwatch/internal/state"
)

func main() {
	cfgPath := flag.String("config", "cronwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("cronwatch: failed to load config: %v", err)
	}

	store, err := state.New(cfg.StateFile)
	if err != nil {
		log.Fatalf("cronwatch: failed to init state: %v", err)
	}

	notifier := buildNotifier(cfg)

	r := runner.New(cfg, store, notifier)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go r.Start()
	log.Printf("cronwatch: started, monitoring %d job(s)", len(cfg.Jobs))

	<-sigCh
	log.Println("cronwatch: shutting down")
	r.Stop()
}

func buildNotifier(cfg *config.Config) alert.Notifier {
	var notifiers []alert.Notifier

	if cfg.Alerts.WebhookURL != "" {
		notifiers = append(notifiers, alert.NewWebhookNotifier(cfg.Alerts.WebhookURL))
	}

	if cfg.Alerts.Email.To != "" {
		n, err := alert.NewEmailNotifier(cfg.Alerts.Email)
		if err != nil {
			log.Printf("cronwatch: skipping email notifier: %v", err)
		} else {
			notifiers = append(notifiers, n)
		}
	}

	// Always include log notifier as fallback
	notifiers = append(notifiers, alert.NewLogNotifier(nil, "[alert]"))

	if len(notifiers) == 1 {
		return notifiers[0]
	}
	return alert.NewMultiNotifier(notifiers...)
}
