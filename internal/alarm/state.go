package alarm

import (
	"fmt"
	"time"

	"pipewatch/internal/store"
)

// State persists alarm documents and keeps per-gauge alarm history.
type State struct {
	st      *store.Store
	alarms  map[string]*Alarm
	order   []string
}

// NewState creates an empty alarm state.
func NewState(st *store.Store) *State {
	return &State{
		st:      st,
		alarms:  make(map[string]*Alarm),
	}
}

// Add persists a new alarm.
func (s *State) Add(a *Alarm) error {
	if _, exists := s.alarms[a.ID]; exists {
		return fmt.Errorf("alarm: duplicate %s", a.ID)
	}
	s.alarms[a.ID] = a
	s.order = append(s.order, a.ID)
	return store.WriteJSON(s.st, "alarms/"+a.ID+".json", a)
}

// Get returns the alarm with the given id.
func (s *State) Get(id string) (*Alarm, bool) {
	a, ok := s.alarms[id]
	return a, ok
}

// List returns all alarms in raise order.
func (s *State) List() []*Alarm {
	out := make([]*Alarm, 0, len(s.order))
	for _, id := range s.order {
		if a, ok := s.alarms[id]; ok {
			out = append(out, a)
		}
	}
	return out
}

// MarkConfirmed confirms exactly one alarm: the alarm whose id was passed.
func (s *State) MarkConfirmed(id string, at time.Time) (*Alarm, error) {
	a, ok := s.alarms[id]
	if !ok {
		return nil, fmt.Errorf("alarm: %s not found", id)
	}
	a.Status = StatusConfirmed
	a.ConfirmedAt = &at
	if err := store.WriteJSON(s.st, "alarms/"+a.ID+".json", a); err != nil {
		return nil, err
	}
	return a, nil
}

// Load restores alarms from the store.
func (s *State) Load() error {
	files, err := s.st.List("alarms")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var a Alarm
		if err := store.ReadJSON(s.st, rel, &a); err != nil {
			return err
		}
		s.alarms[a.ID] = &a
		s.order = append(s.order, a.ID)
	}
	return nil
}
