package scan

import (
	"time"

	"pipewatch/internal/gauge"
)

// Cycle is one full scan pass over the planned gauges.
type Cycle struct {
	Number    int       `json:"number"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Expected  int       `json:"expected"`
	Completed int       `json:"completed"`
	Failed    int       `json:"failed"`
	Skipped   int       `json:"skipped"`
}

// Item is the outcome of reading one gauge inside a batch.
type Item struct {
	GaugeID string
	Reading gauge.Reading
	Err     error
}

// Batch is the result of one scan cycle's batch pass.
type Batch struct {
	CycleNumber int
	StartedAt   time.Time
	Items       []Item
}

// Completed returns the number of successfully read gauges.
func (b *Batch) Completed() int {
	count := 0
	for _, item := range b.Items {
		if item.Err == nil {
			count++
		}
	}
	return count
}

// Failed returns the number of gauges that could not be read.
func (b *Batch) Failed() int {
	count := 0
	for _, item := range b.Items {
		if item.Err != nil {
			count++
		}
	}
	return count
}
