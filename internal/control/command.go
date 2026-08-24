package control

import (
	"fmt"
	"time"

	"pipewatch/internal/store"
)

// Store persists control commands under control/commands/ and keeps the
// execution trail per station.
type Store struct {
	st        *store.Store
	commands  map[string]*Command
	byStation map[string][]string
	seq       int64
}

// NewStore creates an empty command store.
func NewStore(st *store.Store) *Store {
	return &Store{
		st:        st,
		commands:  make(map[string]*Command),
		byStation: make(map[string][]string),
	}
}

// nextSeq advances the issue-sequence counter.
func (c *Store) nextSeq() int64 {
	c.seq++
	return c.seq
}

// Submit accepts a command. A caller-supplied issue sequence is preserved;
// otherwise the arrival sequence is used.
func (c *Store) Submit(cmd Command) (Command, error) {
	if cmd.ID == "" {
		cmd.ID = newCommandID()
	}
	if _, exists := c.commands[cmd.ID]; exists {
		return cmd, fmt.Errorf("control: duplicate command %s", cmd.ID)
	}
	if cmd.IssueSeq == 0 {
		cmd.IssueSeq = c.nextSeq()
	}
	cmd.Status = StatusIssued
	cmd.ArrivedAt = time.Now()
	if cmd.IssuedAt.IsZero() {
		cmd.IssuedAt = cmd.ArrivedAt
	}
	c.commands[cmd.ID] = &cmd
	c.byStation[cmd.StationID] = append(c.byStation[cmd.StationID], cmd.ID)
	if err := c.persist(&cmd); err != nil {
		return cmd, err
	}
	return cmd, nil
}

// Execute marks a command executed and increments its execution count. A
// command that already executed is rejected so a retry never re-fires a
// completed valve action.
func (c *Store) Execute(cmdID string) (*Command, error) {
	cmd, ok := c.commands[cmdID]
	if !ok {
		return nil, fmt.Errorf("control: command %s not found", cmdID)
	}
	if cmd.Executions > 0 {
		return nil, ErrAlreadyExecuted
	}
	cmd.Executions++
	cmd.Status = StatusExecuted
	cmd.ExecutedAt = time.Now()
	if err := c.persist(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

// IsExecuted reports whether a command already produced an execution.
func (c *Store) IsExecuted(cmdID string) bool {
	cmd, ok := c.commands[cmdID]
	return ok && cmd.Executions > 0
}

// Get returns the command with the given id.
func (c *Store) Get(cmdID string) (*Command, bool) {
	cmd, ok := c.commands[cmdID]
	return cmd, ok
}

// Pending returns the issued, not-yet-acked commands of a station.
func (c *Store) Pending(stationID string) []*Command {
	var out []*Command
	for _, id := range c.byStation[stationID] {
		cmd := c.commands[id]
		if cmd.Status != StatusAcked {
			out = append(out, cmd)
		}
	}
	return out
}

// List returns every command of a station in issue order.
func (c *Store) List(stationID string) []*Command {
	out := make([]*Command, 0, len(c.byStation[stationID]))
	for _, id := range c.byStation[stationID] {
		if cmd, ok := c.commands[id]; ok {
			out = append(out, cmd)
		}
	}
	return out
}

// ExecutedTotal returns how many times commands of a station executed.
func (c *Store) ExecutedTotal(stationID string) int {
	total := 0
	for _, cmd := range c.List(stationID) {
		total += cmd.Executions
	}
	return total
}

func (c *Store) persist(cmd *Command) error {
	return store.WriteJSON(c.st, "control/commands/"+cmd.ID+".json", cmd)
}

// Load restores commands from the store.
func (c *Store) Load() error {
	files, err := c.st.List("control/commands")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var cmd Command
		if err := store.ReadJSON(c.st, rel, &cmd); err != nil {
			return err
		}
		if _, exists := c.commands[cmd.ID]; !exists {
			c.byStation[cmd.StationID] = append(c.byStation[cmd.StationID], cmd.ID)
		}
		c.commands[cmd.ID] = &cmd
		if cmd.IssueSeq > c.seq {
			c.seq = cmd.IssueSeq
		}
	}
	return nil
}
