package scan

import (
	"sync"
	"time"
)

// CycleTracker numbers scan cycles so downstream windows align to exactly one
// cycle.
type CycleTracker struct {
	interval time.Duration
	number   int
	current  *Cycle
	mu       sync.Mutex
}

// NewCycleTracker creates a tracker with the configured cycle interval.
func NewCycleTracker(interval time.Duration) *CycleTracker {
	return &CycleTracker{interval: interval}
}

// Begin opens a new scan cycle with the next cycle number.
func (t *CycleTracker) Begin() *Cycle {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.number++
	cycle := &Cycle{Number: t.number, StartedAt: time.Now()}
	t.current = cycle
	return cycle
}

// End closes the current scan cycle.
func (t *CycleTracker) End() *Cycle {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current == nil {
		return nil
	}
	t.current.EndedAt = time.Now()
	return t.current
}

// Current returns the open scan cycle, if any.
func (t *CycleTracker) Current() *Cycle {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.current
}
