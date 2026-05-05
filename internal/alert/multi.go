package alert

import "fmt"

// Notifier is the interface implemented by all alert backends.
type Notifier interface {
	Send(Alert) error
}

// MultiNotifier fans out an alert to multiple Notifier backends.
// All notifiers are attempted; errors are combined.
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier creates a MultiNotifier that sends to all provided notifiers.
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Send delivers the alert to every registered notifier.
// Returns a combined error if one or more notifiers fail.
func (m *MultiNotifier) Send(a Alert) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Send(a); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("multinotifier: %d error(s): %v", len(errs), errs)
}

// Add appends a notifier to the fan-out list.
func (m *MultiNotifier) Add(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}
