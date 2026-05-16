package metrics

import (
	"sync"
	"time"
)

// CooldownEntry tracks per-job cooldown state.
type CooldownEntry struct {
	Job       string
	LastReset time.Time
	Cooldown  time.Duration
	Active    bool
}

// CooldownManager enforces per-job cooldown periods after recovery.
type CooldownManager struct {
	mu      sync.Mutex
	entries map[string]*CooldownEntry
	now     func() time.Time
}

// NewCooldownManager creates a new CooldownManager.
func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		entries: make(map[string]*CooldownEntry),
		now:     time.Now,
	}
}

// SetCooldown registers or updates a cooldown duration for a job.
func (c *CooldownManager) SetCooldown(job string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[job]
	if !ok {
		e = &CooldownEntry{Job: job}
		c.entries[job] = e
	}
	e.Cooldown = d
}

// Activate marks a job as entering cooldown (e.g. after a failure recovery).
func (c *CooldownManager) Activate(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[job]
	if !ok {
		e = &CooldownEntry{Job: job}
		c.entries[job] = e
	}
	e.Active = true
	e.LastReset = c.now()
}

// InCooldown reports whether the job is currently in its cooldown window.
func (c *CooldownManager) InCooldown(job string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[job]
	if !ok || !e.Active {
		return false
	}
	if c.now().Before(e.LastReset.Add(e.Cooldown)) {
		return true
	}
	e.Active = false
	return false
}

// Reset clears the cooldown state for a job.
func (c *CooldownManager) Reset(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[job]; ok {
		e.Active = false
	}
}

// All returns a snapshot of all cooldown entries.
func (c *CooldownManager) All() []CooldownEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]CooldownEntry, 0, len(c.entries))
	for _, e := range c.entries {
		out = append(out, *e)
	}
	return out
}
