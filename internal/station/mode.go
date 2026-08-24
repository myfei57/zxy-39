package station

import (
	"fmt"
)

// SwitchMode changes the operator control mode and persists the new mode. The
// in-memory station is updated in place so the control guard reads the new
// mode back through LiveMode and the very next command is judged against the
// switched mode.
func (r *Registry) SwitchMode(id string, mode Mode) (*Station, error) {
	s, ok := r.stations[id]
	if !ok {
		return nil, fmt.Errorf("station: %s not found", id)
	}
	s.Mode = mode
	s.touch()
	if err := r.Save(s); err != nil {
		return nil, err
	}
	return s, nil
}

// LiveMode returns the current control mode of a station.
func (r *Registry) LiveMode(id string) (string, error) {
	s, ok := r.stations[id]
	if !ok {
		return "", fmt.Errorf("station: %s not found", id)
	}
	return string(s.Mode), nil
}
