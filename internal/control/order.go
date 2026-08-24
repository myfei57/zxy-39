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
	best := 0
	for i := 1; i < len(q.pending); i++ {
		if q.pending[i].IssueSeq < q.pending[best].IssueSeq {
			best = i
		}
	}
	cmd := q.pending[best]
	q.pending = append(q.pending[:best], q.pending[best+1:]...)
	return cmd
}
