package audit

import (
	"time"

	"github.com/google/uuid"
)

// Kind is the audit event category.
type Kind string

const (
	EventCollection Kind = "collection"
	EventControl    Kind = "control"
	EventAlarm      Kind = "alarm"
	EventInterlock  Kind = "interlock"
	EventQuota      Kind = "quota"
	EventRecovery   Kind = "recovery"
)

// Event is one immutable audit record.
type Event struct {
	ID        string    `json:"id"`
	Kind      Kind      `json:"kind"`
	StationID string    `json:"station_id,omitempty"`
	GaugeID   string    `json:"gauge_id,omitempty"`
	SectionID string    `json:"section_id,omitempty"`
	Detail    string    `json:"detail"`
	At        time.Time `json:"at"`
}

// NewEvent creates an audit event with a fresh identifier.
func NewEvent(kind Kind, stationID, gaugeID, sectionID, detail string) *Event {
	return &Event{
		ID:        uuid.NewString(),
		Kind:      kind,
		StationID: stationID,
		GaugeID:   gaugeID,
		SectionID: sectionID,
		Detail:    detail,
		At:        time.Now(),
	}
}
