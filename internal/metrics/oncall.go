package metrics

import (
	"sync"
	"time"
)

// OncallShift represents a named on-call rotation window.
type OncallShift struct {
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OncallManager tracks on-call shifts and maps alert events to the
// shift that was active when the alert fired.
type OncallManager struct {
	mu     sync.RWMutex
	shifts []OncallShift
}

// NewOncallManager returns an empty OncallManager.
func NewOncallManager() *OncallManager {
	return &OncallManager{}
}

// AddShift registers a new on-call shift.
func (m *OncallManager) AddShift(shift OncallShift) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shifts = append(m.shifts, shift)
}

// ActiveAt returns the shift that covers the given time, or nil if none.
func (m *OncallManager) ActiveAt(t time.Time) *OncallShift {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range m.shifts {
		s := &m.shifts[i]
		if !t.Before(s.Start) && t.Before(s.End) {
			return s
		}
	}
	return nil
}

// All returns a copy of all registered shifts.
func (m *OncallManager) All() []OncallShift {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]OncallShift, len(m.shifts))
	copy(out, m.shifts)
	return out
}

// Prune removes shifts whose End time is before the given cutoff.
func (m *OncallManager) Prune(cutoff time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.shifts[:0]
	for _, s := range m.shifts {
		if !s.End.Before(cutoff) {
			kept = append(kept, s)
		}
	}
	m.shifts = kept
}
