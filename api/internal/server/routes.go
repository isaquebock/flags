package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/isaquebock/flags-api/internal/handlers"
	mw "github.com/isaquebock/flags-api/internal/middleware"
)

func (s *Server) setupRoutes() {
	// Global middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(mw.LoggerMiddleware(s.deps.Logger))
	s.router.Use(middleware.Recoverer)
	s.router.Use(mw.CORSMiddleware(s.deps.Config.AllowedOrigins))

	// Health check (public)
	s.router.Get("/healthz", handlers.HealthHandler)

	// Internal routes (require internal token)
	s.router.Route("/internal", func(r chi.Router) {
		r.Use(mw.InternalTokenMiddleware(s.deps.Config.InternalToken))
		r.Use(mw.ClientIDMiddleware)

		r.Get("/snapshot", handlers.NewFlagsHandler(s.deps.Store).InternalSnapshot)
	})

	// Public API routes (require client ID)
	s.router.Route("/v1/flags", func(r chi.Router) {
		r.Use(mw.ClientIDMiddleware)

		handler := handlers.NewFlagsHandler(s.deps.Store)

		r.Get("/", handler.List)
		r.Post("/", handler.Create)

		r.Route("/{key}", func(r chi.Router) {
			r.Get("/", handler.Get)
			r.Patch("/", handler.Update)
			r.Delete("/", handler.Delete)
		})
	})
}
