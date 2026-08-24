package station

// Snapshot is the console-facing view of a station.
type Snapshot struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	PipelineID    string `json:"pipeline_id"`
	SegmentID     string `json:"segment_id"`
	Mode          Mode   `json:"mode"`
	State         State  `json:"state"`
	ValveOpen     bool   `json:"valve_open"`
	InterlockHeld bool   `json:"interlock_held"`
}

// Snapshot returns console-facing copies of every station.
func (r *Registry) Snapshot() []Snapshot {
	out := make([]Snapshot, 0, len(r.order))
	for _, id := range r.order {
		s, ok := r.stations[id]
		if !ok {
			continue
		}
		out = append(out, Snapshot{
			ID:            s.ID,
			Name:          s.Name,
			PipelineID:    s.PipelineID,
			SegmentID:     s.SegmentID,
			Mode:          s.Mode,
			State:         s.State,
			ValveOpen:     s.ValveOpen,
			InterlockHeld: s.InterlockHeld,
		})
	}
	return out
}
