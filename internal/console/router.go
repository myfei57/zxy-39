package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"pipewatch/internal/alarm"
	"pipewatch/internal/audit"
	"pipewatch/internal/control"
	"pipewatch/internal/gauge"
	"pipewatch/internal/historian"
	"pipewatch/internal/interlock"
	"pipewatch/internal/ns"
	"pipewatch/internal/quota"
	"pipewatch/internal/scan"
	"pipewatch/internal/station"
)

// Deps bundles every component the console needs.
type Deps struct {
	Namespaces *ns.Registry
	Stations   *station.Registry
	Gauges     *gauge.Registry
	Scans      *scan.Service
	Alarms     *alarm.Engine
	Interlocks *interlock.Service
	Control    *control.Service
	History    *historian.Service
	Quota      *quota.Service
	Audit      *audit.Service
}

// NewRouter builds the chi HTTP router for the PipeWatch console.
func NewRouter(deps Deps) http.Handler {
	api := &API{deps: deps}
	r := chi.NewRouter()
	r.Get("/healthz", api.health)
	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Get("/pipelines", api.listPipelines)
		apiRouter.Post("/pipelines", api.createPipeline)
		apiRouter.Post("/segments", api.createSegment)
		apiRouter.Get("/stations", api.listStations)
		apiRouter.Post("/stations", api.createStation)
		apiRouter.Get("/stations/{id}", api.getStation)
		apiRouter.Post("/stations/{id}/mode", api.switchMode)
		apiRouter.Post("/stations/{id}/commands", api.dispatchCommand)
		apiRouter.Get("/gauges", api.listGauges)
		apiRouter.Post("/gauges", api.createGauge)
		apiRouter.Post("/gauges/{id}/failover", api.failoverGauge)
		apiRouter.Post("/scan/run", api.runScan)
		apiRouter.Get("/alarms", api.listAlarms)
		apiRouter.Post("/alarms/{id}/confirm", api.confirmAlarm)
		apiRouter.Get("/interlocks", api.listInterlocks)
		apiRouter.Post("/interlocks/{stationID}/retry", api.retryInterlock)
		apiRouter.Post("/interlocks/{stationID}/release", api.releaseInterlock)
		apiRouter.Get("/commands", api.listCommands)
		apiRouter.Post("/commands/{id}/ack", api.ackCommand)
		apiRouter.Get("/history", api.listHistory)
		apiRouter.Get("/audit", api.listAudit)
		apiRouter.Get("/quota", api.listQuota)
		apiRouter.Post("/quota", api.setQuota)
	})
	return r
}
