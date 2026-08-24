package console

// ModeRequest switches the operator control mode of a station.
type ModeRequest struct {
	Mode string `json:"mode"`
}

// CommandRequest is a control command submitted by the console.
type CommandRequest struct {
	StationID string  `json:"station_id"`
	GaugeID   string  `json:"gauge_id,omitempty"`
	Kind      string  `json:"kind"`
	Position  float64 `json:"position"`
	Automatic bool    `json:"automatic"`
}

// QuotaRequest configures the reading budget of a section.
type QuotaRequest struct {
	SectionID   string `json:"section_id"`
	MaxReadings int    `json:"max_readings"`
}

// StationRequest registers a new station.
type StationRequest struct {
	Name       string `json:"name"`
	PipelineID string `json:"pipeline_id"`
	SegmentID  string `json:"segment_id"`
}

// GaugeRequest registers a new transmitter.
type GaugeRequest struct {
	StationID string  `json:"station_id"`
	Tag       string  `json:"tag"`
	Kind      string  `json:"kind"`
	Deadband  float64 `json:"deadband"`
}

// PipelineRequest registers a new pipeline.
type PipelineRequest struct {
	Name           string  `json:"name"`
	Owner          string  `json:"owner"`
	DesignPressure float64 `json:"design_pressure"`
	LengthKM       float64 `json:"length_km"`
}

// SegmentRequest registers a new pipeline segment.
type SegmentRequest struct {
	PipelineID string  `json:"pipeline_id"`
	Name       string  `json:"name"`
	StartKM    float64 `json:"start_km"`
	EndKM      float64 `json:"end_km"`
	Direction  string  `json:"direction"`
}
