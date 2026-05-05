package alert

import "time"

// Level represents the severity of an alert.
type Level string

const (
	LevelMissed Level = "missed"
	LevelFailed Level = "failed"
	LevelOK     Level = "ok"
)

// Alert carries all information about a single notification event.
type Alert struct {
	JobName string
	Level   Level
	Message string
	FiredAt time.Time
}

// Notifier is the interface implemented by all alert back-ends.
type Notifier interface {
	Send(a Alert) error
}

// NewAlert is a convenience constructor that stamps the current time.
func NewAlert(jobName string, level Level, message string) Alert {
	return Alert{
		JobName: jobName,
		Level:   level,
		Message: message,
		FiredAt: time.Now(),
	}
}
