package station

import (
	"context"

	"pipewatch/internal/control"
)

// Dispatch stamps the issue sequence at the moment the operator or the scan
// loop issues the command, then submits it to the control service. Issue
// order is the order in which commands were issued, independent of when they
// arrive at the station.
func (r *Registry) Dispatch(ctx context.Context, ctl *control.Service, cmd control.Command, automatic bool) (control.Command, error) {
	if _, ok := r.Get(cmd.StationID); !ok {
		return cmd, control.ErrStationNotFound
	}
	cmd.IssueSeq = ctl.NextIssueSeq()
	return ctl.Dispatch(ctx, cmd, automatic)
}
