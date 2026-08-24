package interlock

// NewRateWindow creates a rise-rate window limited to limit bar/cycle.
func NewRateWindow(limit float64) *RateWindow {
	return &RateWindow{limit: limit}
}

// Append records one pressure sample for a scan cycle.
func (w *RateWindow) Append(cycle int, value float64) {
	if len(w.samples) > 0 {
		last := w.samples[len(w.samples)-1]
		if last.Cycle < cycle {
			w.samples = w.samples[:0]
		}
	}
	w.samples = append(w.samples, RateSample{Cycle: cycle, Value: value})
}

// Evaluate returns the pressure rise inside exactly one scan cycle. A cycle
// with fewer than two samples cannot trigger.
func (w *RateWindow) Evaluate(cycle int) (float64, bool) {
	var first, last float64
	count := 0
	for _, sample := range w.samples {
		if sample.Cycle == cycle {
			if count == 0 {
				first = sample.Value
			}
			last = sample.Value
			count++
		}
	}
	if count < 2 {
		return 0, false
	}
	rise := last - first
	return rise, rise >= w.limit
}
