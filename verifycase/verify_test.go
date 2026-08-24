package verifycase

import (
	"testing"
	"time"

	"pipewatch/internal/interlock"
	"pipewatch/internal/scan"
)

func TestPsRateWindowAlignedToCycle(t *testing.T) {
	tracker := scan.NewCycleTracker(time.Hour)
	window := interlock.NewRateWindow(1.0)
	tracker.Begin()
	first := tracker.Current()
	window.Append(first.Number, 5.0)
	window.Append(first.Number, 5.2)
	tracker.End()
	tracker.Begin()
	current := tracker.Current()
	window.Append(current.Number, 7.0)
	window.Append(current.Number, 9.0)
	rate, triggered := window.Evaluate(current.Number)
	if !triggered {
		t.Fatalf("rise rate window must cover exactly the current scan cycle; rate=%.2f did not trigger", rate)
	}
}
