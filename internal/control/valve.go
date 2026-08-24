package control

import (
	"fmt"
	"time"

	"pipewatch/internal/store"
)

// Valve is the current mechanical state of one station valve.
type Valve struct {
	StationID string    `json:"station_id"`
	Open      bool      `json:"open"`
	Position  float64   `json:"position"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Valves persists valve states under control/valves/.
type Valves struct {
	st     *store.Store
	valves map[string]*Valve
}

// NewValves creates an empty valve registry.
func NewValves(st *store.Store) *Valves {
	return &Valves{
		st:     st,
		valves: make(map[string]*Valve),
	}
}

// OpenValve opens the station valve fully.
func (v *Valves) OpenValve(stationID string) error {
	valve, ok := v.valves[stationID]
	if !ok {
		valve = &Valve{StationID: stationID}
		v.valves[stationID] = valve
	}
	valve.Open = true
	valve.Position = 100
	valve.UpdatedAt = time.Now()
	return v.persist(valve)
}

// CloseValve closes the station valve.
func (v *Valves) CloseValve(stationID string) error {
	valve, ok := v.valves[stationID]
	if !ok {
		valve = &Valve{StationID: stationID}
		v.valves[stationID] = valve
	}
	valve.Open = false
	valve.Position = 0
	valve.UpdatedAt = time.Now()
	return v.persist(valve)
}

// SetPosition moves the valve to the requested position.
func (v *Valves) SetPosition(stationID string, position float64) error {
	if position < 0 || position > 100 {
		return fmt.Errorf("control: position out of range")
	}
	valve, ok := v.valves[stationID]
	if !ok {
		valve = &Valve{StationID: stationID}
		v.valves[stationID] = valve
	}
	valve.Open = position > 0
	valve.Position = position
	valve.UpdatedAt = time.Now()
	return v.persist(valve)
}

// State returns the current valve state of a station.
func (v *Valves) State(stationID string) (Valve, error) {
	valve, ok := v.valves[stationID]
	if !ok {
		return Valve{}, fmt.Errorf("control: no valve for %s", stationID)
	}
	return *valve, nil
}

func (v *Valves) persist(valve *Valve) error {
	return store.WriteJSON(v.st, "control/valves/"+valve.StationID+".json", valve)
}

// Load restores valve states from the store.
func (v *Valves) Load() error {
	files, err := v.st.List("control/valves")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var valve Valve
		if err := store.ReadJSON(v.st, rel, &valve); err != nil {
			return err
		}
		v.valves[valve.StationID] = &valve
	}
	return nil
}
