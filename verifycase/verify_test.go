package verifycase

import (
	"testing"

	"pipewatch/internal/control"
)

func TestPsControlAppliesInIssueOrder(t *testing.T) {
	queue := control.NewQueue()
	open := &control.Command{
		ID: "cmd-a", StationID: "S-1", Kind: control.CmdOpen,
		IssueSeq: 1, Status: control.StatusIssued,
	}
	half := &control.Command{
		ID: "cmd-b", StationID: "S-1", Kind: control.CmdSetPosition,
		Position: 50, IssueSeq: 2, Status: control.StatusIssued,
	}
	queue.Enqueue(half)
	queue.Enqueue(open)
	var applied []string
	for cmd := queue.Next(); cmd != nil; cmd = queue.Next() {
		applied = append(applied, cmd.ID)
	}
	if len(applied) != 2 || applied[0] != "cmd-a" || applied[1] != "cmd-b" {
		t.Fatalf("commands must apply in issue order, got %v", applied)
	}
}
