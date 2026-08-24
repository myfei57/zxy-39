package gauge

import "math"

// ApplyDeadband applies the analog deadband filter to a reading for display
// and storage. It only computes the filtered display value; the raw sensor
// measurement on r.Value, r.Raw and g.RawValue is left untouched so the alarm
// path always evaluates the true reading. Without this, a small genuine
// fluctuation that never clears the deadband would be suppressed before the
// alarm judgement runs and could never raise an alarm.
func (g *Gauge) ApplyDeadband(r *Reading) {
	last := g.FilteredValue
	if math.Abs(r.Value-last) < g.Deadband {
		r.Filtered = last
	} else {
		r.Filtered = r.Value
	}
	g.FilteredValue = r.Filtered
}
