package console

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"pipewatch/internal/audit"
	"pipewatch/internal/control"
	"pipewatch/internal/gauge"
	"pipewatch/internal/historian"
	"pipewatch/internal/ns"
	"pipewatch/internal/station"
)

// API implements the console HTTP handlers.
type API struct {
	deps Deps
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) listPipelines(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.Namespaces.ListPipelines())
}

func (a *API) createPipeline(w http.ResponseWriter, r *http.Request) {
	var req PipelineRequest
	if !readBody(w, r, &req) {
		return
	}
	p := ns.NewPipeline(req.Name, req.Owner, req.DesignPressure, req.LengthKM)
	if err := a.deps.Namespaces.RegisterPipeline(p); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (a *API) createSegment(w http.ResponseWriter, r *http.Request) {
	var req SegmentRequest
	if !readBody(w, r, &req) {
		return
	}
	s := ns.NewSegment(req.PipelineID, req.Name, req.StartKM, req.EndKM, req.Direction)
	if err := a.deps.Namespaces.RegisterSegment(s); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (a *API) listStations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.Stations.Snapshot())
}

func (a *API) createStation(w http.ResponseWriter, r *http.Request) {
	var req StationRequest
	if !readBody(w, r, &req) {
		return
	}
	s := station.NewStation(req.Name, req.PipelineID, req.SegmentID)
	if err := a.deps.Stations.Register(s); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (a *API) getStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, ok := a.deps.Stations.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, fmt.Errorf("station %s not found", id))
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (a *API) switchMode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req ModeRequest
	if !readBody(w, r, &req) {
		return
	}
	s, err := a.deps.Stations.SwitchMode(id, station.Mode(req.Mode))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	a.deps.Audit.Record(*audit.NewEvent(audit.EventControl, id, "", "", "mode switched to "+req.Mode))
	writeJSON(w, http.StatusOK, s)
}

func (a *API) dispatchCommand(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req CommandRequest
	if !readBody(w, r, &req) {
		return
	}
	if req.Automatic && !a.deps.Control.Allows(id, control.Kind(req.Kind)) {
		writeErr(w, http.StatusConflict, fmt.Errorf("automatic command rejected in current station mode"))
		return
	}
	cmd := control.NewCommand(id, req.GaugeID, control.Kind(req.Kind), req.Position)
	submitted, err := a.deps.Stations.Dispatch(r.Context(), a.deps.Control, *cmd, req.Automatic)
	if err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	a.deps.Audit.Record(*audit.NewEvent(audit.EventControl, id, req.GaugeID, "", "dispatched "+req.Kind))
	writeJSON(w, http.StatusCreated, submitted)
}

func (a *API) listGauges(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.Gauges.Snapshot())
}

func (a *API) createGauge(w http.ResponseWriter, r *http.Request) {
	var req GaugeRequest
	if !readBody(w, r, &req) {
		return
	}
	g := gauge.NewGauge(req.StationID, req.Tag, gauge.Kind(req.Kind), req.Deadband)
	if err := a.deps.Gauges.Register(g); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (a *API) failoverGauge(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.deps.Gauges.Failover(id, "console failover request"); err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"gauge_id": id, "comm_state": string(gauge.CommBackup)})
}

func (a *API) runScan(w http.ResponseWriter, r *http.Request) {
	cycle, err := a.deps.Scans.RunCycle(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cycle)
}

func (a *API) listAlarms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.Alarms.List())
}

func (a *API) confirmAlarm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	al, err := a.deps.Alarms.Confirm(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	a.deps.Audit.Record(*audit.NewEvent(audit.EventAlarm, al.StationID, al.GaugeID, al.SectionID, "confirmed"))
	writeJSON(w, http.StatusOK, al)
}

func (a *API) listInterlocks(w http.ResponseWriter, r *http.Request) {
	states := make([]map[string]any, 0)
	for _, latch := range a.deps.Interlocks.Latches() {
		state := map[string]any{
			"station_id": latch.StationID,
			"held":       latch.Held,
			"valve_open": a.deps.Control.ValveOpen(r.Context(), latch.StationID),
		}
		if valve, err := a.deps.Control.ValveState(latch.StationID); err == nil {
			state["position"] = valve.Position
		}
		states = append(states, state)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"latches":      states,
		"suppressions": a.deps.Alarms.ListSuppressions(),
	})
}

func (a *API) retryInterlock(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")
	resent, err := a.deps.Interlocks.RetryPending(r.Context(), stationID)
	if err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	a.deps.Audit.Record(*audit.NewEvent(audit.EventInterlock, stationID, "", "", fmt.Sprintf("retried %d commands", resent)))
	writeJSON(w, http.StatusOK, map[string]int{"resent": resent})
}

func (a *API) releaseInterlock(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")
	if err := a.deps.Interlocks.Release(r.Context(), stationID); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	a.deps.Audit.Record(*audit.NewEvent(audit.EventInterlock, stationID, "", "", "released"))
	writeJSON(w, http.StatusOK, map[string]bool{"released": true})
}

func (a *API) listCommands(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station_id")
	if stationID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("station_id is required"))
		return
	}
	writeJSON(w, http.StatusOK, a.deps.Control.Pending(stationID))
}

func (a *API) ackCommand(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cmd, err := a.deps.Control.Ack(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, cmd)
}

func (a *API) listHistory(w http.ResponseWriter, r *http.Request) {
	sectionID := r.URL.Query().Get("section")
	kind := historian.WindowKind(r.URL.Query().Get("kind"))
	if sectionID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("section is required"))
		return
	}
	if kind == "" {
		kind = historian.WindowHour
	}
	summaries, err := a.deps.History.List(sectionID, kind)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, map[string]any{
			"id":        summary.ID,
			"section":   summary.SectionID,
			"kind":      summary.Kind,
			"start":     summary.Start,
			"end":       summary.End,
			"count":     summary.Count,
			"min":       summary.Min,
			"max":       summary.Max,
			"avg":       summary.Avg(),
		})
	}
	writeJSON(w, http.StatusOK, items)
}

func (a *API) listAudit(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	events, err := a.deps.Audit.Recent(limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (a *API) listQuota(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.Quota.List())
}

func (a *API) setQuota(w http.ResponseWriter, r *http.Request) {
	var req QuotaRequest
	if !readBody(w, r, &req) {
		return
	}
	if err := a.deps.Quota.Set(req.SectionID, req.MaxReadings); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}
