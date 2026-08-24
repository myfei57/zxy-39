package control

import (
	"context"
)

// Execute applies a command to the valve and records the execution. Commands
// that already executed are rejected by the store.
func (s *Service) Execute(ctx context.Context, cmdID string) (*Command, error) {
	cmd, err := s.store.Execute(cmdID)
	if err != nil {
		return nil, err
	}
	switch cmd.Kind {
	case CmdOpen:
		return cmd, s.valves.OpenValve(cmd.StationID)
	case CmdClose:
		return cmd, s.valves.CloseValve(cmd.StationID)
	case CmdSetPosition:
		return cmd, s.valves.SetPosition(cmd.StationID, cmd.Position)
	default:
		return cmd, nil
	}
}

// Drain applies every waiting command in issue order.
func (s *Service) Drain(ctx context.Context) (int, error) {
	count := 0
	for {
		cmd := s.queue.Next()
		if cmd == nil {
			return count, nil
		}
		if _, err := s.Execute(ctx, cmd.ID); err != nil {
			return count, err
		}
		count++
	}
}
