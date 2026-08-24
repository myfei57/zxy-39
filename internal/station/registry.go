package station

import (
	"fmt"

	"pipewatch/internal/store"
)

// Registry owns the station directory, persisted under stations/.
type Registry struct {
	st       *store.Store
	stations map[string]*Station
	order    []string
}

// NewRegistry creates an empty station registry.
func NewRegistry(st *store.Store) *Registry {
	return &Registry{
		st:       st,
		stations: make(map[string]*Station),
	}
}

// Register stores a new station document.
func (r *Registry) Register(s *Station) error {
	if s.ID == "" || s.Name == "" || s.SegmentID == "" {
		return fmt.Errorf("station: id, name and segment are required")
	}
	if _, exists := r.stations[s.ID]; !exists {
		r.order = append(r.order, s.ID)
	}
	r.stations[s.ID] = s
	return r.Save(s)
}

// Get returns the station with the given id.
func (r *Registry) Get(id string) (*Station, bool) {
	s, ok := r.stations[id]
	return s, ok
}

// List returns all stations in registration order.
func (r *Registry) List() []*Station {
	var out []*Station
	for _, id := range r.order {
		if s, ok := r.stations[id]; ok {
			out = append(out, s)
		}
	}
	return out
}

// Save persists one station document.
func (r *Registry) Save(s *Station) error {
	return store.WriteJSON(r.st, "stations/"+s.ID+".json", s)
}

// SetValve updates the reported valve position of a station.
func (r *Registry) SetValve(s *Station, open bool) error {
	s.ValveOpen = open
	s.touch()
	return r.Save(s)
}

// SetInterlock updates the interlock marker of a station.
func (r *Registry) SetInterlock(s *Station, held bool) error {
	s.InterlockHeld = held
	s.touch()
	return r.Save(s)
}

// SetState moves a station along the operational state machine.
func (r *Registry) SetState(s *Station, state State) error {
	s.State = state
	s.touch()
	return r.Save(s)
}

// Load restores stations from the store.
func (r *Registry) Load() error {
	files, err := r.st.List("stations")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var s Station
		if err := store.ReadJSON(r.st, rel, &s); err != nil {
			return err
		}
		r.Register(&s)
	}
	return nil
}
