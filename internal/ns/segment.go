package ns

import (
	"fmt"

	"github.com/google/uuid"
)

// Segment is a numbered section of a pipeline between two stations.
type Segment struct {
	ID         string  `json:"id"`
	PipelineID string  `json:"pipeline_id"`
	Name       string  `json:"name"`
	StartKM    float64 `json:"start_km"`
	EndKM      float64 `json:"end_km"`
	Direction  string  `json:"direction"`
}

// NewSegment creates a segment belonging to pipelineID.
func NewSegment(pipelineID, name string, startKM, endKM float64, direction string) *Segment {
	return &Segment{
		ID:         uuid.NewString(),
		PipelineID: pipelineID,
		Name:       name,
		StartKM:    startKM,
		EndKM:      endKM,
		Direction:  direction,
	}
}

// Validate checks the segment geometry used by the scanning plan.
func (s *Segment) Validate() error {
	if s.ID == "" || s.PipelineID == "" || s.Name == "" {
		return fmt.Errorf("ns: segment identity fields are required")
	}
	if s.EndKM <= s.StartKM {
		return fmt.Errorf("ns: segment end must be after start")
	}
	if s.Direction == "" {
		return fmt.Errorf("ns: segment direction is required")
	}
	return nil
}
