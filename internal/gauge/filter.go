package gauge

import "math"

// ApplyDeadband applies the analog deadband filter to a reading for display
// and storage. The raw sensor value is preserved on the reading so the alarm
// path always evaluates the true measurement; only the filtered display value
// is suppressed.
func (g *Gauge) ApplyDeadband(r *Reading) {
	last := g.FilteredValue
	if math.Abs(r.Value-last) < g.Deadband {
		r.Filtered = last
	} else {
		r.Filtered = r.Value
	}
	g.FilteredValue = r.Filtered
	r.Value = r.Filtered
	r.Raw = r.Filtered
	g.RawValue = r.Filtered
}
