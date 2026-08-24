package station

import (
	"time"

	"github.com/google/uuid"
)

// Mode is the operator control mode of a station.
type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeManual Mode = "manual"
)

// State is the operational state of a station along the scan/alarm/interlock
// lifecycle.
type State string

const (
	StateNormal      State = "normal"
	StateAlarm       State = "alarm"
	StateInterlocked State = "interlocked"
	StateReleased    State = "released"
)

// Station is a pump station or valve chamber registered on a pipeline segment.
type Station struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	PipelineID    string    `json:"pipeline_id"`
	SegmentID     string    `json:"segment_id"`
	Mode          Mode      `json:"mode"`
	State         State     `json:"state"`
	ValveOpen     bool      `json:"valve_open"`
	InterlockHeld bool      `json:"interlock_held"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewStation creates a station in auto mode with the valve open.
func NewStation(name, pipelineID, segmentID string) *Station {
	now := time.Now()
	return &Station{
		ID:         uuid.NewString(),
		Name:       name,
		PipelineID: pipelineID,
		SegmentID:  segmentID,
		Mode:       ModeAuto,
		State:      StateNormal,
		ValveOpen:  true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (s *Station) touch() {
	s.UpdatedAt = time.Now()
}
