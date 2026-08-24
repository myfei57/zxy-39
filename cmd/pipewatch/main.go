package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pipewatch/internal/alarm"
	"pipewatch/internal/audit"
	"pipewatch/internal/console"
	"pipewatch/internal/control"
	"pipewatch/internal/gauge"
	"pipewatch/internal/historian"
	"pipewatch/internal/interlock"
	"pipewatch/internal/ns"
	"pipewatch/internal/quota"
	"pipewatch/internal/scan"
	"pipewatch/internal/station"
	"pipewatch/internal/store"
)

func main() {
	config := LoadConfig()
	if err := run(config); err != nil {
		log.Fatalf("pipewatch: %v", err)
	}
}

func run(config Config) error {
	st, err := store.New(config.DataDir)
	if err != nil {
		return err
	}
	namespaces := ns.NewRegistry(st)
	stations := station.NewRegistry(st)
	gauges := gauge.NewRegistry(st)
	gauges.SetSource(gauge.DemoSource())
	alarms := alarm.NewEngine(st, alarm.DefaultThresholds())
	guard := control.NewGuard(func(stationID string) (string, error) {
		return stations.LiveMode(stationID)
	})
	valves := control.NewValves(st)
	controlSvc := control.NewService(st, guard, valves)
	interlocks := interlock.NewService(st, controlSvc)
	history := historian.NewService(st)
	quotas := quota.NewService(st)
	auditSvc := audit.NewService(st)
	tracker := scan.NewCycleTracker(config.ScanInterval)
	poller := scan.NewPoller(gauges)
	scanSvc := scan.NewService(st, stations, gauges, alarms, interlocks, history, quotas, auditSvc, tracker, poller)

	if err := namespaces.Load(); err != nil {
		return err
	}
	if err := stations.Load(); err != nil {
		return err
	}
	if err := gauges.Load(); err != nil {
		return err
	}
	if err := alarms.Load(); err != nil {
		return err
	}
	if err := controlSvc.Load(); err != nil {
		return err
	}
	if err := interlocks.Load(); err != nil {
		return err
	}
	if err := quotas.Load(); err != nil {
		return err
	}
	if err := auditSvc.Load(); err != nil {
		return err
	}
	if len(stations.List()) == 0 {
		seed(st, namespaces, stations, gauges, quotas)
	}
	scanSvc.SetPlan(gauges.List(), sectionIDs(stations))

	router := console.NewRouter(console.Deps{
		Namespaces: namespaces,
		Stations:   stations,
		Gauges:     gauges,
		Scans:      scanSvc,
		Alarms:     alarms,
		Interlocks: interlocks,
		Control:    controlSvc,
		History:    history,
		Quota:      quotas,
		Audit:      auditSvc,
	})
	server := &http.Server{
		Addr:              config.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		ticker := time.NewTicker(config.ScanInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := scanSvc.RunCycle(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("scan cycle failed: %v", err)
				}
				if _, err := controlSvc.Drain(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("command drain failed: %v", err)
				}
			}
		}
	}()

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("pipewatch console listening on %s", config.Addr)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		<-scanDone
		return nil
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func seed(st *store.Store, namespaces *ns.Registry, stations *station.Registry, gauges *gauge.Registry, quotas *quota.Service) {
	pipeline := ns.NewPipeline("华东一号干线", "调度中心", 10.0, 486)
	_ = namespaces.RegisterPipeline(pipeline)
	sectionA := ns.NewSegment(pipeline.ID, "A 段", 0, 120, "forward")
	sectionB := ns.NewSegment(pipeline.ID, "B 段", 120, 260, "forward")
	_ = namespaces.RegisterSegment(sectionA)
	_ = namespaces.RegisterSegment(sectionB)
	s1 := station.NewStation("首站泵房", pipeline.ID, sectionA.ID)
	s2 := station.NewStation("中间阀室", pipeline.ID, sectionB.ID)
	_ = stations.Register(s1)
	_ = stations.Register(s2)
	gauges.Register(gauge.NewGauge(s1.ID, "PT-101", gauge.KindPressure, 0.2))
	gauges.Register(gauge.NewGauge(s1.ID, "FT-102", gauge.KindFlow, 0.5))
	gauges.Register(gauge.NewGauge(s2.ID, "PT-201", gauge.KindPressure, 0.2))
	gauges.Register(gauge.NewGauge(s2.ID, "TT-202", gauge.KindTemperature, 1.0))
	_ = quotas.Set(sectionA.ID, 10000)
	_ = quotas.Set(sectionB.ID, 10000)
}

func sectionIDs(stations *station.Registry) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range stations.List() {
		if !seen[s.SegmentID] {
			seen[s.SegmentID] = true
			out = append(out, s.SegmentID)
		}
	}
	return out
}
