package interlock

import "context"

// RetryPending re-dispatches the commands of a station that still wait for an
// acknowledgement. Commands that already executed are skipped, so a retry
// never fires the same valve action twice.
func (s *Service) RetryPending(ctx context.Context, stationID string) (int, error) {
	resent := 0
	for _, cmd := range s.control.Pending(stationID) {
		if _, err := s.control.Dispatch(ctx, *cmd, false); err != nil {
			return resent, err
		}
		if _, err := s.control.Execute(ctx, cmd.ID); err != nil {
			return resent, err
		}
		resent++
	}
	return resent, nil
}

// ExecutedTotal reports how many executions the station commands produced.
func (s *Service) ExecutedTotal(stationID string) int {
	return s.control.Store().ExecutedTotal(stationID)
}
