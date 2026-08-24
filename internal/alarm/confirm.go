package alarm

import "time"

// Confirm acknowledges exactly the selected alarm. Newer alarms raised for
// the same transmitter stay raised until they are reviewed individually.
func (e *Engine) Confirm(id string) (*Alarm, error) {
	a, ok := e.state.Get(id)
	if !ok {
		return nil, errAlarmNotFound
	}
	return e.state.MarkConfirmed(a.GaugeID, time.Now())
}
