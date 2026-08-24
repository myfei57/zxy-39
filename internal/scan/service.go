package scan

import (
	"context"
	"fmt"
	"time"

	"pipewatch/internal/alarm"
	"pipewatch/internal/audit"
	"pipewatch/internal/gauge"
	"pipewatch/internal/historian"
	"pipewatch/internal/interlock"
	"pipewatch/internal/quota"
	"pipewatch/internal/station"
	"pipewatch/internal/store"
)

// Service runs the scan cycles and wires collection, alarm evaluation,
// interlock evaluation, history aggregation, quota and audit together.
type Service struct {
	store      *store.Store
	stations   *station.Registry
	gauges     *gauge.Registry
	alarms     *alarm.Engine
	interlocks *interlock.Service
	history    *historian.Service
	quota      *quota.Service
	audit      *audit.Service
	tracker    *CycleTracker
	poller     *Poller
	rate       *interlock.RateWindow
	plan       []*gauge.Gauge
	sections   []string
}

// NewService wires the scan service to every downstream component.
func NewService(
	st *store.Store,
	stations *station.Registry,
	gauges *gauge.Registry,
	alarms *alarm.Engine,
	interlocks *interlock.Service,
	history *historian.Service,
	quotaSvc *quota.Service,
	auditSvc *audit.Service,
	tracker *CycleTracker,
	poller *Poller,
) *Service {
	return &Service{
		store:      st,
		stations:   stations,
		gauges:     gauges,
		alarms:     alarms,
		interlocks: interlocks,
		history:    history,
		quota:      quotaSvc,
		audit:      auditSvc,
		tracker:    tracker,
		poller:     poller,
		rate:       interlock.NewRateWindow(1.5),
	}
}

// RunCycle performs one full scan cycle and returns its summary.
func (s *Service) RunCycle(ctx context.Context) (*Cycle, error) {
	cycle := s.tracker.Begin()
	cycle.Expected = len(s.plan)
	batch, batchErr := s.ScanBatch(ctx, s.plan)
	if batchErr != nil {
		return cycle, batchErr
	}
	cycle.Completed = batch.Completed()
	cycle.Failed = batch.Failed()
	failedSections := make(map[string]bool)
	readSections := make(map[string]bool)
	for _, item := range batch.Items {
		sectionID := s.sectionOfGauge(item.GaugeID)
		if item.Err != nil {
			failedSections[sectionID] = true
			s.audit.Record(*audit.NewEvent(audit.EventCollection, "", item.GaugeID, sectionID, "read failed: "+item.Err.Error()))
			continue
		}
		r := item.Reading
		r.SectionID = sectionID
		if err := s.quota.Check(sectionID); err != nil {
			cycle.Skipped++
			s.audit.Record(*audit.NewEvent(audit.EventQuota, r.StationID, r.GaugeID, sectionID, "rejected by quota"))
			continue
		}
		_ = s.quota.Consume(sectionID)
		g, _ := s.gauges.Get(r.GaugeID)
		g.ApplyDeadband(&r)
		raised, err := s.alarms.Evaluate(r, sectionID)
		if err != nil {
			s.audit.Record(*audit.NewEvent(audit.EventAlarm, r.StationID, r.GaugeID, sectionID, "evaluate failed"))
			continue
		}
		_ = store.WriteJSON(s.store, fmt.Sprintf("readings/%s-%s.json", r.GaugeID, r.TakenAt.UTC().Format("20060102T150405.000000000")), r)
		s.history.AddReading(r)
		s.audit.Record(*audit.NewEvent(audit.EventCollection, r.StationID, r.GaugeID, sectionID, "collected"))
		for _, a := range raised {
			s.audit.Record(*audit.NewEvent(audit.EventAlarm, r.StationID, r.GaugeID, sectionID, "raised "+string(a.Severity)))
			if a.Severity == alarm.SeverityHigh {
				_ = s.alarms.Trigger(ctx, s.interlocks, r.StationID, "high alarm")
				if st, ok := s.stations.Get(r.StationID); ok {
					_ = s.stations.SetState(st, station.StateAlarm)
					_ = s.stations.SetInterlock(st, true)
					_ = s.stations.SetValve(st, false)
				}
			}
		}
		if r.Kind == gauge.KindPressure {
			s.rate.Append(cycle.Number, r.Value)
			if _, hit := s.rate.Evaluate(cycle.Number); hit {
				s.audit.Record(*audit.NewEvent(audit.EventInterlock, r.StationID, r.GaugeID, sectionID, "rise rate triggered"))
				_ = s.alarms.Trigger(ctx, s.interlocks, r.StationID, "pressure rise rate")
			}
			if r.Value <= s.alarms.Thresholds().PressureLow && s.interlocks.IsHeld(r.StationID) {
				if err := s.interlocks.Release(ctx, r.StationID); err == nil {
					if st, ok := s.stations.Get(r.StationID); ok {
						_ = s.stations.SetState(st, station.StateReleased)
						_ = s.stations.SetInterlock(st, false)
						_ = s.stations.SetValve(st, true)
					}
				}
			}
		}
		readSections[sectionID] = true
	}
	for sectionID := range readSections {
		if failedSections[sectionID] {
			continue
		}
		if err := s.HandleRecovery(ctx, sectionID); err != nil {
			s.audit.Record(*audit.NewEvent(audit.EventRecovery, "", "", sectionID, "recovery write failed"))
		} else {
			s.audit.Record(*audit.NewEvent(audit.EventRecovery, "", "", sectionID, "recovery durable"))
		}
		_, _ = s.history.Aggregate(sectionID, historian.WindowHour, time.Now())
	}
	s.tracker.End()
	return cycle, nil
}

func (s *Service) sectionOfGauge(gaugeID string) string {
	g, ok := s.gauges.Get(gaugeID)
	if !ok {
		return "unknown"
	}
	if st, ok := s.stations.Get(g.StationID); ok {
		return st.SegmentID
	}
	return "unknown"
}
