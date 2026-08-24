package alarm

import "context"

// InterlockTrigger is the interlock entry point driven by alarm conditions.
type InterlockTrigger interface {
	Set(ctx context.Context, stationID, reason string) error
}

// Trigger raises the interlock for a station through the interlock service.
func (e *Engine) Trigger(ctx context.Context, il InterlockTrigger, stationID, reason string) error {
	return il.Set(ctx, stationID, reason)
}
