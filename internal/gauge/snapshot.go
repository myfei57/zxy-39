package gauge

import "time"

// Snapshot is the console-facing view of a transmitter.
type Snapshot struct {
	ID            string    `json:"id"`
	StationID     string    `json:"station_id"`
	Tag           string    `json:"tag"`
	Kind          Kind      `json:"kind"`
	CommState     CommState `json:"comm_state"`
	RawValue      float64   `json:"raw_value"`
	FilteredValue float64   `json:"filtered_value"`
	LastReading   time.Time `json:"last_reading"`
	LastQuality   Quality   `json:"last_quality"`
}

// Snapshot returns console-facing copies of every transmitter.
func (r *Registry) Snapshot() []Snapshot {
	out := make([]Snapshot, 0, len(r.order))
	for _, id := range r.order {
		g, ok := r.gauges[id]
		if !ok {
			continue
		}
		out = append(out, Snapshot{
			ID:            g.ID,
			StationID:     g.StationID,
			Tag:           g.Tag,
			Kind:          g.Kind,
			CommState:     g.CommState,
			RawValue:      g.RawValue,
			FilteredValue: g.FilteredValue,
			LastReading:   g.LastReading,
			LastQuality:   g.LastQuality,
		})
	}
	return out
}
