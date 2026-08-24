package control

import (
	"time"
)

// Ack records the operator acknowledgement of an executed command. An ack is
// only accepted after the execution record exists.
func (c *Store) Ack(cmdID string, at time.Time) (*Command, error) {
	cmd, ok := c.commands[cmdID]
	if !ok {
		return nil, ErrStationNotFound
	}
	if cmd.Status != StatusExecuted {
		return nil, ErrAckBeforeExecution
	}
	cmd.Status = StatusAcked
	cmd.AckedAt = at
	if err := c.persist(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}
