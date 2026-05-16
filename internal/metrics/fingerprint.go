package metrics

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
	"time"
)

// FingerprintRecord holds a computed fingerprint for a job alert event.
type FingerprintRecord struct {
	Job       string    `json:"job"`
	Reason    string    `json:"reason"`
	Hash      string    `json:"hash"`
	FirstSeen time.Time `json:"first_seen"`
	Count     int       `json:"count"`
}

// FingerprintStore tracks unique alert fingerprints to identify recurring patterns.
type FingerprintStore struct {
	mu      sync.Mutex
	entries map[string]*FingerprintRecord
	maxAge  time.Duration
}

// NewFingerprintStore creates a FingerprintStore with the given max age for entries.
func NewFingerprintStore(maxAge time.Duration) *FingerprintStore {
	return &FingerprintStore{
		entries: make(map[string]*FingerprintRecord),
		maxAge:  maxAge,
	}
}

// Record adds or increments a fingerprint entry for the given job and reason.
// Returns the fingerprint hash.
func (f *FingerprintStore) Record(job, reason string, at time.Time) string {
	hash := computeHash(job, reason)
	f.mu.Lock()
	defer f.mu.Unlock()
	if rec, ok := f.entries[hash]; ok {
		rec.Count++
		return hash
	}
	f.entries[hash] = &FingerprintRecord{
		Job:       job,
		Reason:    reason,
		Hash:      hash,
		FirstSeen: at,
		Count:     1,
	}
	return hash
}

// Get returns the record for the given hash, or nil if not found.
func (f *FingerprintStore) Get(hash string) *FingerprintRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec := f.entries[hash]
	return rec
}

// All returns all fingerprint records sorted by first seen descending.
func (f *FingerprintStore) All(now time.Time) []FingerprintRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prune(now)
	result := make([]FingerprintRecord, 0, len(f.entries))
	for _, rec := range f.entries {
		result = append(result, *rec)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FirstSeen.After(result[j].FirstSeen)
	})
	return result
}

func (f *FingerprintStore) prune(now time.Time) {
	for k, rec := range f.entries {
		if now.Sub(rec.FirstSeen) > f.maxAge {
			delete(f.entries, k)
		}
	}
}

func computeHash(job, reason string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s", job, reason)))
	return fmt.Sprintf("%x", h[:8])
}
