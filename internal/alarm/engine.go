package alarm

import (
	"errors"
	"time"

	"pipewatch/internal/store"
)

var errAlarmNotFound = errors.New("alarm: not found")

// Engine evaluates readings, raises and confirms alarms, and manages flood
// suppression per pipeline section.
type Engine struct {
	st         *store.Store
	state      *State
	suppress   *Suppression
	thresholds Thresholds
	floodLimit int
	floodWin   time.Duration
	recent     map[string][]time.Time
}

// NewEngine wires the alarm state, suppression set and thresholds.
func NewEngine(st *store.Store, thresholds Thresholds) *Engine {
	return &Engine{
		st:         st,
		state:      NewState(st),
		suppress:   NewSuppression(),
		thresholds: thresholds,
		floodLimit: 5,
		floodWin:   time.Minute,
		recent:     make(map[string][]time.Time),
	}
}

// Load restores alarms and suppression entries from the store.
func (e *Engine) Load() error {
	if err := e.state.Load(); err != nil {
		return err
	}
	files, err := e.st.List("alarms/suppression")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var entry SectionSuppression
		if err := store.ReadJSON(e.st, rel, &entry); err != nil {
			return err
		}
		e.suppress.entries[entry.SectionID] = &entry
	}
	return nil
}

// Raise persists a new alarm and applies flood detection for its section.
func (e *Engine) Raise(a Alarm) (*Alarm, error) {
	created := &a
	if err := e.state.Add(created); err != nil {
		return nil, err
	}
	e.trackFlood(a.SectionID)
	return created, nil
}

func (e *Engine) trackFlood(sectionID string) {
	now := time.Now()
	window := e.recent[sectionID]
	window = append(window, now)
	cutoff := now.Add(-e.floodWin)
	kept := window[:0]
	for _, at := range window {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	e.recent[sectionID] = kept
	if len(kept) >= e.floodLimit {
		_ = e.StartSuppression(sectionID, "alarm flood", now.Add(5*time.Minute))
	}
}

// Get returns the alarm with the given id.
func (e *Engine) Get(id string) (*Alarm, bool) {
	return e.state.Get(id)
}

// List returns all alarms.
func (e *Engine) List() []*Alarm {
	return e.state.List()
}

// Thresholds returns the configured alarm limits.
func (e *Engine) Thresholds() Thresholds {
	return e.thresholds
}
