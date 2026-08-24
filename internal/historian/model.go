package historian

import (
	"fmt"
	"time"
)

// WindowKind is the aggregation granularity of a history window.
type WindowKind string

const (
	WindowHour WindowKind = "hour"
	WindowDay  WindowKind = "day"
)

// Summary is the aggregated statistics of one window.
type Summary struct {
	ID        string     `json:"id"`
	SectionID string     `json:"section_id"`
	Kind      WindowKind `json:"kind"`
	Start     time.Time  `json:"start"`
	End       time.Time  `json:"end"`
	Count     int        `json:"count"`
	Sum       float64    `json:"sum"`
	Min       float64    `json:"min"`
	Max       float64    `json:"max"`
}

// Avg returns the mean value of the window.
func (s *Summary) Avg() float64 {
	if s.Count == 0 {
		return 0
	}
	return s.Sum / float64(s.Count)
}

// Windows computes hour/day window boundaries for the historian.
type Windows struct{}

// NewWindows creates the boundary calculator.
func NewWindows() *Windows {
	return &Windows{}
}

// Start returns the start of the window containing at.
func (w *Windows) Start(kind WindowKind, at time.Time) time.Time {
	switch kind {
	case WindowDay:
		y, m, d := at.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, at.Location())
	default:
		y, m, d := at.Date()
		return time.Date(y, m, d, at.Hour(), 0, 0, 0, at.Location())
	}
}

// End returns the exclusive end of the window starting at start.
func (w *Windows) End(start time.Time, kind WindowKind) time.Time {
	switch kind {
	case WindowDay:
		return start.AddDate(0, 0, 1)
	default:
		return start.Add(time.Hour)
	}
}

// Key returns the document key of a window.
func (w *Windows) Key(kind WindowKind, start time.Time) string {
	format := "20060102T15"
	if kind == WindowDay {
		format = "20060102"
	}
	return fmt.Sprintf("historian/summaries/%s/%s.json", kind, start.Format(format))
}
