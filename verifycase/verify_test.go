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

func TestPsInterlockRetrySkipsExecuted(t *testing.T) {
	ts := newTestSystem(t)
	ts.addStation(t, "S-1", "SEC-1")
	ctx := context.Background()
	if err := ts.interlock.Set(ctx, "S-1", "overpressure"); err != nil {
		t.Fatalf("set interlock: %v", err)
	}
	if total := ts.interlock.ExecutedTotal("S-1"); total != 1 {
		t.Fatalf("expected exactly one close execution after the interlock, got %d", total)
	}
	if _, err := ts.interlock.RetryPending(ctx, "S-1"); err != nil {
		t.Fatalf("retry pending: %v", err)
	}
	if total := ts.interlock.ExecutedTotal("S-1"); total != 1 {
		t.Fatalf("retry must skip an already executed close command; got %d executions", total)
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
