package verifycase

import (
	"testing"
	"time"

	"pipewatch/internal/alarm"
	"pipewatch/internal/audit"
	"pipewatch/internal/control"
	"pipewatch/internal/gauge"
	"pipewatch/internal/historian"
	"pipewatch/internal/interlock"
	"pipewatch/internal/quota"
	"pipewatch/internal/scan"
	"pipewatch/internal/station"
	"pipewatch/internal/store"
)

type testSystem struct {
	st        *store.Store
	stations  *station.Registry
	gauges    *gauge.Registry
	alarms    *alarm.Engine
	control   *control.Service
	interlock *interlock.Service
	quota     *quota.Service
	tracker   *scan.CycleTracker
	poller    *scan.Poller
	scan      *scan.Service
}

func newTestSystem(t *testing.T) *testSystem {
	t.Helper()
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stations := station.NewRegistry(st)
	gauges := gauge.NewRegistry(st)
	alarms := alarm.NewEngine(st, alarm.DefaultThresholds())
	guard := control.NewGuard(func(id string) (string, error) { return stations.LiveMode(id) })
	valves := control.NewValves(st)
	ctl := control.NewService(st, guard, valves)
	interlocks := interlock.NewService(st, ctl)
	history := historian.NewService(st)
	quotas := quota.NewService(st)
	auditSvc := audit.NewService(st)
	tracker := scan.NewCycleTracker(time.Minute)
	poller := scan.NewPoller(gauges)
	scanSvc := scan.NewService(st, stations, gauges, alarms, interlocks, history, quotas, auditSvc, tracker, poller)
	return &testSystem{
		st:        st,
		stations:  stations,
		gauges:    gauges,
		alarms:    alarms,
		control:   ctl,
		interlock: interlocks,
		quota:     quotas,
		tracker:   tracker,
		poller:    poller,
		scan:      scanSvc,
	}
}

func TestPsConfirmOnlyTargetAlarm(t *testing.T) {
	ts := newTestSystem(t)
	ts.addStation(t, "S-1", "SEC-1")
	ts.addGauge(t, "PT-101", "S-1", gauge.KindPressure, 0.1)
	first := time.Now().Add(-time.Minute)
	second := time.Now()
	if _, err := ts.alarms.Raise(alarm.Alarm{
		ID: "alarm-1", GaugeID: "PT-101", StationID: "S-1", SectionID: "SEC-1",
		Severity: alarm.SeverityLow, Value: 8.0, RaisedAt: first, Status: alarm.StatusRaised,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := ts.alarms.Raise(alarm.Alarm{
		ID: "alarm-2", GaugeID: "PT-101", StationID: "S-1", SectionID: "SEC-1",
		Severity: alarm.SeverityHigh, Value: 10.0, RaisedAt: second, Status: alarm.StatusRaised,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := ts.alarms.Confirm("alarm-1"); err != nil {
		t.Fatalf("confirm old alarm: %v", err)
	}
	newer, ok := ts.alarms.Get("alarm-2")
	if !ok {
		t.Fatal("newer alarm missing")
	}
	if newer.Status != alarm.StatusRaised {
		t.Fatalf("confirming the old alarm must not confirm the newer alarm, got %s", newer.Status)
	}
}

func (ts *testSystem) addStation(t *testing.T, id, segment string) *station.Station {
	t.Helper()
	s := station.NewStation(id, "P-1", segment)
	s.ID = id
	if err := ts.stations.Register(s); err != nil {
		t.Fatal(err)
	}
	return s
}

func (ts *testSystem) addGauge(t *testing.T, id, stationID string, kind gauge.Kind, deadband float64) *gauge.Gauge {
	t.Helper()
	g := gauge.NewGauge(stationID, id, kind, deadband)
	g.ID = id
	if err := ts.gauges.Register(g); err != nil {
		t.Fatal(err)
	}
	return g
}
