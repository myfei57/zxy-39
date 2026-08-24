package control

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Kind is the action a control command requests.
type Kind string

const (
	CmdOpen        Kind = "open"
	CmdClose       Kind = "close"
	CmdSetPosition Kind = "set-position"
)

// Status is the lifecycle of a control command.
type Status string

const (
	StatusIssued   Status = "issued"
	StatusExecuted Status = "executed"
	StatusAcked    Status = "acked"
)

// ErrStationNotFound is returned when a command targets an unknown station.
var ErrStationNotFound = errors.New("control: station not found")

// ErrAlreadyExecuted is returned when a command is executed a second time.
var ErrAlreadyExecuted = errors.New("control: command already executed")

// ErrAckBeforeExecution is returned when an ack arrives before execution.
var ErrAckBeforeExecution = errors.New("control: ack before execution")

// ErrRejectedByMode is returned when the mode guard rejects an automatic
// command.
var ErrRejectedByMode = errors.New("control: rejected by mode guard")

// Command is one valve instruction with its full execution trail.
type Command struct {
	ID         string    `json:"id"`
	StationID  string    `json:"station_id"`
	GaugeID    string    `json:"gauge_id,omitempty"`
	Kind       Kind      `json:"kind"`
	Position   float64   `json:"position"`
	IssueSeq   int64     `json:"issue_seq"`
	Status     Status    `json:"status"`
	Executions int       `json:"executions"`
	IssuedAt   time.Time `json:"issued_at"`
	ArrivedAt  time.Time `json:"arrived_at"`
	ExecutedAt time.Time `json:"executed_at,omitempty"`
	AckedAt    time.Time `json:"acked_at,omitempty"`
}

// NewCommand creates an issued command with a fresh identifier.
func NewCommand(stationID, gaugeID string, kind Kind, position float64) *Command {
	return &Command{
		ID:        uuid.NewString(),
		StationID: stationID,
		GaugeID:   gaugeID,
		Kind:      kind,
		Position:  position,
		Status:    StatusIssued,
		IssuedAt:  time.Now(),
	}
}

func newCommandID() string {
	return uuid.NewString()
}
