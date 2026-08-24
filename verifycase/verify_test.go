package verifycase

import (
	"context"
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

func TestPsScanContinuesAfterGaugeTimeout(t *testing.T) {
	ts := newTestSystem(t)
	ts.addStation(t, "S-1", "SEC-1")
	ts.addGauge(t, "PT-101", "S-1", gauge.KindPressure, 0.1)
	ts.addGauge(t, "PT-102", "S-1", gauge.KindPressure, 0.1)
	ts.addGauge(t, "PT-103", "S-1", gauge.KindPressure, 0.1)
	ts.gauges.SetSource(gauge.SourceFunc(func(g *gauge.Gauge) (float64, error) {
		if g.ID == "PT-102" {
			return 0, gauge.ErrTimeout
		}
		return 7.5, nil
	}))
	ts.tracker.Begin()
	plan := ts.gauges.List()
	batch, err := ts.scan.ScanBatch(context.Background(), plan)
	if err != nil {
		t.Fatalf("one gauge timeout must not abort the whole batch: %v", err)
	}
	if len(batch.Items) != 3 {
		t.Fatalf("expected all three gauges in the batch, got %d items", len(batch.Items))
	}
	if batch.Completed() != 2 || batch.Failed() != 1 {
		t.Fatalf("expected 2 completed and 1 failed, got completed=%d failed=%d", batch.Completed(), batch.Failed())
	}
}
