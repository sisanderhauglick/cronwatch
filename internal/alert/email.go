package alert

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// EmailNotifier sends alert notifications via SMTP.
type EmailNotifier struct {
	Host     string
	Port     int
	From     string
	To       []string
	Username string
	Password string
}

// NewEmailNotifier creates an EmailNotifier from the provided configuration.
func NewEmailNotifier(host string, port int, from string, to []string, username, password string) *EmailNotifier {
	return &EmailNotifier{
		Host:     host,
		Port:     port,
		From:     from,
		To:       to,
		Username: username,
		Password: password,
	}
}

// Send delivers the alert as an email message.
func (e *EmailNotifier) Send(a Alert) error {
	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)

	var auth smtp.Auth
	if e.Username != "" {
		auth = smtp.PlainAuth("", e.Username, e.Password, e.Host)
	}

	subject := fmt.Sprintf("[cronwatch] %s – job %q", a.Level, a.JobName)
	body := fmt.Sprintf(
		"Job:     %s\nLevel:   %s\nMessage: %s\nTime:    %s\n",
		a.JobName,
		a.Level,
		a.Message,
		a.FiredAt.Format(time.RFC1123),
	)

	msg := []byte(strings.Join([]string{
		"From: " + e.From,
		"To: " + strings.Join(e.To, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		body,
	}, "\r\n"))

	return smtp.SendMail(addr, auth, e.From, e.To, msg)
}
