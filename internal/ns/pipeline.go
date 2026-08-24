package ns

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Pipeline describes one trunk pipeline in the transmission network.
type Pipeline struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Owner          string    `json:"owner"`
	DesignPressure float64   `json:"design_pressure"`
	LengthKM       float64   `json:"length_km"`
	CreatedAt      time.Time `json:"created_at"`
}

// NewPipeline creates a pipeline with a fresh identifier.
func NewPipeline(name, owner string, designPressure, lengthKM float64) *Pipeline {
	return &Pipeline{
		ID:             uuid.NewString(),
		Name:           name,
		Owner:          owner,
		DesignPressure: designPressure,
		LengthKM:       lengthKM,
		CreatedAt:      time.Now(),
	}
}

// Validate checks the pipeline fields that downstream logic depends on.
func (p *Pipeline) Validate() error {
	if p.ID == "" || p.Name == "" {
		return fmt.Errorf("ns: pipeline id and name are required")
	}
	if p.DesignPressure <= 0 {
		return fmt.Errorf("ns: design pressure must be positive")
	}
	if p.LengthKM <= 0 {
		return fmt.Errorf("ns: pipeline length must be positive")
	}
	return nil
}
