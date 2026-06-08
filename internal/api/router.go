package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Alexmaster12345/infrabios/internal/api/handlers"
	"github.com/Alexmaster12345/infrabios/internal/api/middleware"
	"github.com/Alexmaster12345/infrabios/internal/db"
	"github.com/Alexmaster12345/infrabios/internal/service"
)

func NewRouter(
	database *db.DB,
	compliance *service.ComplianceService,
	drift *service.DriftService,
	apiTokens []string,
	agentToken string,
	log *zap.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(middleware.Recoverer(log))
	r.Use(middleware.RequestLogger(log))

	// Health — no auth
	r.Get("/api/v1/health", handlers.Health)

	// Operator API
	r.Group(func(r chi.Router) {
		r.Use(middleware.BearerAuth(apiTokens))

		srv := handlers.NewServersHandler(database, drift, log)
		r.Post("/api/v1/servers", srv.Create)
		r.Get("/api/v1/servers", srv.List)
		r.Get("/api/v1/servers/{id}", srv.Get)
		r.Delete("/api/v1/servers/{id}", srv.Delete)
		r.Put("/api/v1/servers/{id}/profile", srv.AssignProfile)
		r.Get("/api/v1/servers/{id}/settings", srv.GetSettings)

		prof := handlers.NewProfilesHandler(database, log)
		r.Post("/api/v1/profiles", prof.Create)
		r.Get("/api/v1/profiles", prof.List)
		r.Get("/api/v1/profiles/{id}", prof.Get)
		r.Put("/api/v1/profiles/{id}", prof.Update)
		r.Delete("/api/v1/profiles/{id}", prof.Delete)

		comp := handlers.NewComplianceHandler(database, compliance, log)
		r.Post("/api/v1/servers/{id}/compliance/scan", comp.Scan)
		r.Get("/api/v1/servers/{id}/compliance", comp.Latest)
		r.Get("/api/v1/servers/{id}/compliance/history", comp.History)

		dr := handlers.NewDriftHandler(database, log)
		r.Get("/api/v1/drift", dr.List)
		r.Post("/api/v1/drift/{id}/resolve", dr.Resolve)
		r.Post("/api/v1/drift/{id}/ignore", dr.Ignore)

		fw := handlers.NewFirmwareHandler(database, log)
		r.Get("/api/v1/servers/{id}/firmware", fw.List)
		r.Post("/api/v1/servers/{id}/firmware", fw.Upsert)
		r.Post("/api/v1/servers/{id}/firmware/{component}/approve", fw.Approve)

		ch := handlers.NewChangesHandler(database, log)
		r.Post("/api/v1/changes", ch.Create)
		r.Get("/api/v1/changes", ch.List)
		r.Get("/api/v1/changes/{id}", ch.Get)
		r.Post("/api/v1/changes/{id}/review", ch.Review)

		snap := handlers.NewSnapshotsHandler(database, log)
		r.Post("/api/v1/servers/{id}/snapshots", snap.Create)
		r.Get("/api/v1/servers/{id}/snapshots", snap.List)
		r.Get("/api/v1/snapshots/{id}", snap.Get)
		r.Delete("/api/v1/snapshots/{id}", snap.Delete)

		jobs := handlers.NewJobsHandler(database, log)
		r.Post("/api/v1/jobs", jobs.Create)
		r.Get("/api/v1/jobs", jobs.List)
		r.Get("/api/v1/jobs/{id}", jobs.Get)
		r.Post("/api/v1/jobs/{id}/cancel", jobs.Cancel)
	})

	// Agent API — separate token
	r.Group(func(r chi.Router) {
		r.Use(middleware.AgentAuth(agentToken))

		srv := handlers.NewServersHandler(database, drift, log)
		r.Post("/api/v1/agent/servers/{id}/inventory", srv.SubmitInventory)

		fw := handlers.NewFirmwareHandler(database, log)
		r.Post("/api/v1/agent/servers/{id}/firmware", fw.Upsert)
	})

	return r
}
