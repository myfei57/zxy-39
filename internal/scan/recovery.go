package scan

import (
	"context"
	"time"

	"pipewatch/internal/store"
)

// Recovery records that a pipeline section returned to normal operation.
type Recovery struct {
	SectionID   string    `json:"section_id"`
	CycleNumber int       `json:"cycle_number"`
	RecoveredAt time.Time `json:"recovered_at"`
}

// HandleRecovery clears flood suppression only after the recovery record is
// durable. A failed recovery write leaves the section suppressed so later
// real alarms stay visible.
func (s *Service) HandleRecovery(ctx context.Context, sectionID string) error {
	if !s.alarms.IsSuppressed(sectionID) {
		return nil
	}
	recovery := Recovery{
		SectionID:   sectionID,
		CycleNumber: s.tracker.Current().Number,
		RecoveredAt: time.Now(),
	}
	if err := s.alarms.EndSuppression(sectionID); err != nil {
		return err
	}
	_ = store.WriteJSON(s.store, "recovery/"+sectionID+".json", recovery)
	return nil
}
