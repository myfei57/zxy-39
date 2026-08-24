package interlock

import (
	"pipewatch/internal/control"
	"pipewatch/internal/store"
)

// Service owns the interlock latches and the retry flow of a station.
type Service struct {
	st      *store.Store
	control *control.Service
	latches map[string]*Latch
}

// NewService wires the interlock latches to the control service.
func NewService(st *store.Store, ctl *control.Service) *Service {
	return &Service{
		st:      st,
		control: ctl,
		latches: make(map[string]*Latch),
	}
}

// Load restores interlock latches from the store.
func (s *Service) Load() error {
	files, err := s.st.List("interlocks")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var latch Latch
		if err := store.ReadJSON(s.st, rel, &latch); err != nil {
			return err
		}
		s.latches[latch.StationID] = &latch
	}
	return nil
}

// Latches returns copies of every latch for the console.
func (s *Service) Latches() []Latch {
	out := make([]Latch, 0, len(s.latches))
	for _, latch := range s.latches {
		out = append(out, *latch)
	}
	return out
}
