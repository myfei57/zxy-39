package alarm

import (
	"time"

	"github.com/google/uuid"
)

// Severity of an alarm condition.
type Severity string

const (
	SeverityLow  Severity = "low"
	SeverityHigh Severity = "high"
)

// Status of an alarm lifecycle.
type Status string

const (
	StatusRaised   Status = "raised"
	StatusConfirmed Status = "confirmed"
)

// Alarm is one threshold crossing raised for a transmitter.
type Alarm struct {
	ID          string     `json:"id"`
	GaugeID     string     `json:"gauge_id"`
	StationID   string     `json:"station_id"`
	SectionID   string     `json:"section_id"`
	Severity    Severity   `json:"severity"`
	Value       float64    `json:"value"`
	RaisedAt    time.Time  `json:"raised_at"`
	Status      Status     `json:"status"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

// NewAlarm creates a raised alarm with a fresh identifier.
func NewAlarm(gaugeID, stationID, sectionID string, severity Severity, value float64) *Alarm {
	return &Alarm{
		ID:        uuid.NewString(),
		GaugeID:   gaugeID,
		StationID: stationID,
		SectionID: sectionID,
		Severity:  severity,
		Value:     value,
		RaisedAt:  time.Now(),
		Status:    StatusRaised,
	}
}

// Thresholds are the alarm limits per measurement kind.
type Thresholds struct {
	PressureHigh float64 `json:"pressure_high"`
	PressureLow  float64 `json:"pressure_low"`
	FlowHigh     float64 `json:"flow_high"`
	FlowLow      float64 `json:"flow_low"`
	TempHigh     float64 `json:"temp_high"`
	TempLow      float64 `json:"temp_low"`
}

// DefaultThresholds returns the production alarm limits.
func DefaultThresholds() Thresholds {
	return Thresholds{
		PressureHigh: 9.0,
		PressureLow:  3.0,
		FlowHigh:     80.0,
		FlowLow:      5.0,
		TempHigh:     65.0,
		TempLow:      -10.0,
	}
}
