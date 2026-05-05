package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// JobState holds the last known execution state for a single cron job.
type JobState struct {
	LastSeen  time.Time `json:"last_seen"`
	LastStatus string   `json:"last_status"` // "ok" | "failed" | "missed"
}

// Store persists job states to a JSON file.
type Store struct {
	mu     sync.RWMutex
	path   string
	states map[string]JobState
}

// New loads (or creates) a state store at the given file path.
func New(path string) (*Store, error) {
	s := &Store{
		path:   path,
		states: make(map[string]JobState),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Get returns the state for a job by name.
func (s *Store) Get(name string) (JobState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.states[name]
	return st, ok
}

// Set updates the state for a job and flushes to disk.
func (s *Store) Set(name string, st JobState) error {
	s.mu.Lock()
	s.states[name] = st
	s.mu.Unlock()
	return s.flush()
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.NewDecoder(f).Decode(&s.states)
}

func (s *Store) flush() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.states); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(tmp, s.path)
}
