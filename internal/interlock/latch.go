package interlock

import (
	"context"
	"fmt"
	"time"

	"pipewatch/internal/control"
	"pipewatch/internal/store"
)

// Set holds the interlock latch of a station and closes the valve. Re-setting
// an already held latch is idempotent.
func (s *Service) Set(ctx context.Context, stationID, reason string) error {
	if existing, ok := s.latches[stationID]; ok && existing.Held {
		return nil
	}
	cmd := control.NewCommand(stationID, "", control.CmdClose, 0)
	submitted, err := s.control.Dispatch(ctx, *cmd, false)
	if err != nil {
		return fmt.Errorf("interlock: dispatch close for %s: %w", stationID, err)
	}
	if _, err := s.control.Execute(ctx, submitted.ID); err != nil {
		return err
	}
	latch := &Latch{
		StationID: stationID,
		Reason:    reason,
		Held:      true,
		HeldAt:    time.Now(),
	}
	s.latches[stationID] = latch
	return s.persist(latch)
}

// Release clears the interlock latch once the pressure condition is gone and
// reopens the valve. The latch is cleared first; the valve follows. The caller
// (scan cycle or console) is responsible for confirming the pressure has
// returned to normal; the valve is held closed by Set for as long as the latch
// is held, so it must not be a precondition for release.
func (s *Service) Release(ctx context.Context, stationID string) error {
	latch, ok := s.latches[stationID]
	if !ok || !latch.Held {
		return nil
	}
	latch.Held = false
	latch.ReleasedAt = time.Now()
	if err := s.persist(latch); err != nil {
		return err
	}
	return s.control.OpenValve(ctx, stationID)
}

// IsHeld reports whether the interlock latch of a station is currently set.
func (s *Service) IsHeld(stationID string) bool {
	latch, ok := s.latches[stationID]
	return ok && latch.Held
}

func (s *Service) persist(latch *Latch) error {
	return store.WriteJSON(s.st, "interlocks/"+latch.StationID+".json", latch)
}
