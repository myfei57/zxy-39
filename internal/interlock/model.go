package interlock

import "time"

// Latch is the interlock state of one station valve.
type Latch struct {
	StationID  string    `json:"station_id"`
	Reason     string    `json:"reason"`
	Held       bool      `json:"held"`
	HeldAt     time.Time `json:"held_at"`
	ReleasedAt time.Time `json:"released_at,omitempty"`
}

// RateWindow evaluates the pressure rise rate inside one scan cycle.
type RateWindow struct {
	limit   float64
	samples []RateSample
}

// RateSample is one pressure reading tagged with its scan cycle.
type RateSample struct {
	Cycle int     `json:"cycle"`
	Value float64 `json:"value"`
}
