package gauge

import (
	"time"

	"github.com/google/uuid"
)

// Kind is the measured physical quantity of a transmitter.
type Kind string

const (
	KindPressure    Kind = "pressure"
	KindFlow        Kind = "flow"
	KindTemperature Kind = "temperature"
)

// Quality describes how trustworthy the latest reading is.
type Quality string

const (
	QualityGood   Quality = "good"
	QualityStale  Quality = "stale"
	QualityFailed Quality = "failed"
)

// CommState is the active communication path of a transmitter.
type CommState string

const (
	CommPrimary CommState = "primary"
	CommBackup  CommState = "backup"
)

// Channel is one communication path of a transmitter.
type Channel struct {
	Name        string `json:"name"`
	Healthy     bool   `json:"healthy"`
	FailureCount int   `json:"failure_count"`
	LastError   string `json:"last_error,omitempty"`
}

// NewChannel creates a healthy communication channel.
func NewChannel(name string) Channel {
	return Channel{Name: name, Healthy: true}
}

// Gauge is a pressure, flow or temperature transmitter attached to a station.
type Gauge struct {
	ID            string    `json:"id"`
	StationID     string    `json:"station_id"`
	Tag           string    `json:"tag"`
	Kind          Kind      `json:"kind"`
	Channels      [2]Channel `json:"channels"`
	ActiveChannel int       `json:"active_channel"`
	CommState     CommState `json:"comm_state"`
	Deadband      float64   `json:"deadband"`
	RawValue      float64   `json:"raw_value"`
	FilteredValue float64   `json:"filtered_value"`
	LastReading   time.Time `json:"last_reading"`
	LastQuality   Quality   `json:"last_quality"`
}

// NewGauge creates a transmitter with a primary and a backup channel.
func NewGauge(stationID, tag string, kind Kind, deadband float64) *Gauge {
	return &Gauge{
		ID:            uuid.NewString(),
		StationID:     stationID,
		Tag:           tag,
		Kind:          kind,
		Channels:      [2]Channel{NewChannel("A"), NewChannel("B")},
		ActiveChannel: 0,
		CommState:     CommPrimary,
		Deadband:      deadband,
		LastQuality:   QualityGood,
	}
}

// Reading is one sampled value from a transmitter.
type Reading struct {
	GaugeID   string    `json:"gauge_id"`
	StationID string    `json:"station_id"`
	SectionID string    `json:"section_id,omitempty"`
	Kind      Kind      `json:"kind"`
	Value     float64   `json:"value"`
	Raw       float64   `json:"raw"`
	Filtered  float64   `json:"filtered"`
	Quality   Quality   `json:"quality"`
	TakenAt   time.Time `json:"taken_at"`
}
