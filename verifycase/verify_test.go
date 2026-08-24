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

func TestPsScanUsesCurrentChannel(t *testing.T) {
	ts := newTestSystem(t)
	ts.addStation(t, "S-1", "SEC-1")
	ts.addGauge(t, "PT-101", "S-1", gauge.KindPressure, 0.1)
	ts.gauges.SetSource(gauge.SourceFunc(func(g *gauge.Gauge) (float64, error) {
		if g.ResolveChannel().Name == "B" {
			return 8.0, nil
		}
		return 5.0, nil
	}))
	ctx := context.Background()
	before, err := ts.poller.ReadOne(ctx, "PT-101")
	if err != nil {
		t.Fatal(err)
	}
	if before.Value != 5.0 {
		t.Fatalf("expected primary channel value 5.0, got %v", before.Value)
	}
	if err := ts.gauges.Failover("PT-101", "primary channel lost"); err != nil {
		t.Fatal(err)
	}
	after, err := ts.poller.ReadOne(ctx, "PT-101")
	if err != nil {
		t.Fatal(err)
	}
	if after.Value != 8.0 {
		t.Fatalf("scan must read from the backup channel after failover, got %v", after.Value)
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
