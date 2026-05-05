package alert

import (
	"errors"
	"strings"
	"testing"
)

type captureNotifier struct {
	received []Alert
	err      error
}

func (c *captureNotifier) Send(a Alert) error {
	c.received = append(c.received, a)
	return c.err
}

func TestMultiNotifier_SendsToAll(t *testing.T) {
	n1 := &captureNotifier{}
	n2 := &captureNotifier{}
	mn := NewMultiNotifier(n1, n2)

	a := makeAlert()
	if err := mn.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(n1.received) != 1 || len(n2.received) != 1 {
		t.Errorf("expected each notifier to receive 1 alert")
	}
}

func TestMultiNotifier_CombinesErrors(t *testing.T) {
	n1 := &captureNotifier{err: errors.New("n1 fail")}
	n2 := &captureNotifier{err: errors.New("n2 fail")}
	mn := NewMultiNotifier(n1, n2)

	err := mn.Send(makeAlert())
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !strings.Contains(err.Error(), "2 error(s)") {
		t.Errorf("expected error count in message, got: %v", err)
	}
}

func TestMultiNotifier_PartialFailure(t *testing.T) {
	n1 := &captureNotifier{}
	n2 := &captureNotifier{err: errors.New("fail")}
	mn := NewMultiNotifier(n1, n2)

	err := mn.Send(makeAlert())
	if err == nil {
		t.Fatal("expected error from partial failure")
	}
	// n1 should still have received the alert
	if len(n1.received) != 1 {
		t.Errorf("expected n1 to receive alert despite n2 failure")
	}
}

func TestMultiNotifier_Add(t *testing.T) {
	mn := NewMultiNotifier()
	n := &captureNotifier{}
	mn.Add(n)

	if err := mn.Send(makeAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(n.received) != 1 {
		t.Error("expected notifier added via Add to receive alert")
	}
}
