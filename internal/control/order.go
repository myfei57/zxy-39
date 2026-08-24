package control

// Queue holds commands waiting to be applied and hands them out in issue
// order, so a later-issued command never overtakes an earlier one even when
// the messages arrive out of order.
type Queue struct {
	pending []*Command
}

// NewQueue creates an empty command queue.
func NewQueue() *Queue {
	return &Queue{}
}

// Enqueue adds a command to the waiting set.
func (q *Queue) Enqueue(cmd *Command) {
	q.pending = append(q.pending, cmd)
}

// Next pops the waiting command with the smallest issue sequence. It returns
// nil when the queue is empty.
func (q *Queue) Next() *Command {
	if len(q.pending) == 0 {
		return nil
	}
	cmd := q.pending[0]
	q.pending = q.pending[1:]
	return cmd
}
