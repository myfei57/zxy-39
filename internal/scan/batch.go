package scan

import (
	"context"
	"time"

	"pipewatch/internal/gauge"
)

// ScanBatch reads every gauge in the plan. A failing gauge is recorded as a
// failed item and does not stop the remaining gauges of the same cycle, so a
// single timeout never blanks a whole line section.
func (s *Service) ScanBatch(ctx context.Context, plan []*gauge.Gauge) (Batch, error) {
	batch := Batch{
		CycleNumber: s.tracker.Current().Number,
		StartedAt:   time.Now(),
	}
	for _, g := range plan {
		reading, err := s.poller.ReadOne(ctx, g.ID)
		if err != nil {
			batch.Items = append(batch.Items, Item{GaugeID: g.ID, Err: err})
			continue
		}
		batch.Items = append(batch.Items, Item{GaugeID: g.ID, Reading: reading})
	}
	return batch, nil
}
