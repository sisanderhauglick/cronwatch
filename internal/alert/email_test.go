package alert

import (
	"io"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

// stubSMTPServer starts a minimal SMTP server on a random port and returns
// the listener and a channel that receives the raw message data.
func stubSMTPServer(t *testing.T) (net.Listener, chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	msgCh := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		tc := textproto.NewConn(conn)
		_ = tc.PrintfLine("220 stub ready")
		var buf strings.Builder
		for {
			line, err := tc.ReadLine()
			if err != nil || err == io.EOF {
				break
			}
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				_ = tc.PrintfLine("250 ok")
			case strings.HasPrefix(upper, "MAIL"), strings.HasPrefix(upper, "RCPT"):
				_ = tc.PrintfLine("250 ok")
			case upper == "DATA":
				_ = tc.PrintfLine("354 go ahead")
			case line == ".":
				_ = tc.PrintfLine("250 queued")
				msgCh <- buf.String()
			case strings.HasPrefix(upper, "QUIT"):
				_ = tc.PrintfLine("221 bye")
				return
			default:
				buf.WriteString(line + "\n")
			}
		}
	}()
	return ln, msgCh
}

func TestEmailNotifier_Send_Success(t *testing.T) {
	_ = smtp.SendMail // ensure import used
	a := Alert{
		JobName: "backup",
		Level:   "missed",
		Message: "job did not run",
		FiredAt: time.Now(),
	}

	ln, msgCh := stubSMTPServer(t)
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	n := NewEmailNotifier("127.0.0.1", addr.Port, "cron@example.com", []string{"ops@example.com"}, "", "")

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	select {
	case body := <-msgCh:
		if !strings.Contains(body, "backup") {
			t.Errorf("expected job name in body, got: %s", body)
		}
	default:
		t.Error("no message received by stub server")
	}
}

func TestEmailNotifier_NewEmailNotifier_Fields(t *testing.T) {
	n := NewEmailNotifier("mail.example.com", 587, "from@example.com", []string{"a@b.com"}, "user", "pass")
	if n.Host != "mail.example.com" {
		t.Errorf("unexpected Host: %s", n.Host)
	}
	if n.Port != 587 {
		t.Errorf("unexpected Port: %d", n.Port)
	}
	if n.Username != "user" {
		t.Errorf("unexpected Username: %s", n.Username)
	}
}
