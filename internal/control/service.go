package control

import (
	"context"
	"time"

	"pipewatch/internal/store"
)

// Service orchestrates command intake, ordering, execution and acknowledgement.
type Service struct {
	store  *Store
	queue  *Queue
	guard  *Guard
	valves *Valves
}

// NewService wires the command store, issue-order queue, mode guard and valve
// registry together.
func NewService(st *store.Store, guard *Guard, valves *Valves) *Service {
	return &Service{
		store:  NewStore(st),
		queue:  NewQueue(),
		guard:  guard,
		valves: valves,
	}
}

// Load restores commands and valve states from the store.
func (s *Service) Load() error {
	if err := s.store.Load(); err != nil {
		return err
	}
	return s.valves.Load()
}

// Store exposes the command store for retry and audit flows.
func (s *Service) Store() *Store {
	return s.store
}

// NextIssueSeq reserves the next issue-order sequence number.
func (s *Service) NextIssueSeq() int64 {
	return s.store.nextSeq()
}

// Dispatch accepts a command. Automatic commands pass the mode guard first.
func (s *Service) Dispatch(ctx context.Context, cmd Command, automatic bool) (Command, error) {
	if automatic && !s.guard.Allow(cmd.StationID, cmd.Kind) {
		return cmd, ErrRejectedByMode
	}
	submitted, err := s.store.Submit(cmd)
	if err != nil {
		return cmd, err
	}
	queued := submitted
	s.queue.Enqueue(&queued)
	return submitted, nil
}

// Allows reports whether an automatic command would pass the mode guard.
func (s *Service) Allows(stationID string, kind Kind) bool {
	return s.guard.Allow(stationID, kind)
}

// ValveOpen reports whether the station valve is currently open.
func (s *Service) ValveOpen(ctx context.Context, stationID string) bool {
	valve, err := s.valves.State(stationID)
	return err == nil && valve.Open
}

// ValveState returns the current valve state of a station.
func (s *Service) ValveState(stationID string) (Valve, error) {
	return s.valves.State(stationID)
}

// OpenValve opens the station valve.
func (s *Service) OpenValve(ctx context.Context, stationID string) error {
	return s.valves.OpenValve(stationID)
}

// Ack records the acknowledgement of an executed command.
func (s *Service) Ack(ctx context.Context, cmdID string) (*Command, error) {
	return s.store.Ack(cmdID, time.Now())
}

// Pending lists the waiting commands of a station.
func (s *Service) Pending(stationID string) []*Command {
	return s.store.Pending(stationID)
}
