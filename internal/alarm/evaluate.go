package alarm

import (
	"pipewatch/internal/gauge"
)

// Evaluate checks a reading against the alarm thresholds. The raw sensor
// value is evaluated; the deadband filter is applied afterwards for display
// and storage, so a small genuine fluctuation still reaches the alarm path.
func (e *Engine) Evaluate(r gauge.Reading, sectionID string) ([]*Alarm, error) {
	var raised []*Alarm
	value := r.Raw
	switch r.Kind {
	case gauge.KindPressure:
		if value >= e.thresholds.PressureHigh {
			raised = append(raised, e.mustRaise(r, sectionID, SeverityHigh))
		}
		if value <= e.thresholds.PressureLow {
			raised = append(raised, e.mustRaise(r, sectionID, SeverityLow))
		}
	case gauge.KindFlow:
		if value >= e.thresholds.FlowHigh {
			raised = append(raised, e.mustRaise(r, sectionID, SeverityHigh))
		}
		if value <= e.thresholds.FlowLow {
			raised = append(raised, e.mustRaise(r, sectionID, SeverityLow))
		}
	case gauge.KindTemperature:
		if value >= e.thresholds.TempHigh {
			raised = append(raised, e.mustRaise(r, sectionID, SeverityHigh))
		}
		if value <= e.thresholds.TempLow {
			raised = append(raised, e.mustRaise(r, sectionID, SeverityLow))
		}
	}
	return raised, nil
}

func (e *Engine) mustRaise(r gauge.Reading, sectionID string, severity Severity) *Alarm {
	a, err := e.Raise(*NewAlarm(r.GaugeID, r.StationID, sectionID, severity, r.Value))
	if err != nil {
		return nil
	}
	return a
}
