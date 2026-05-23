package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/isaquebock/flags-api/internal/config"
	"github.com/isaquebock/flags-api/internal/snapshot"
)

type Deps struct {
	Logger *slog.Logger
	Store  snapshot.Store
	Config *config.Config
}

type Server struct {
	deps   *Deps
	router chi.Router
}

func New(deps Deps) *Server {
	router := chi.NewRouter()

	srv := &Server{
		deps:   &deps,
		router: router,
	}

	srv.setupRoutes()

	return srv
}

func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:         ":" + s.deps.Config.Port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		s.deps.Logger.Info("shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			s.deps.Logger.Error("shutdown error", "error", err)
		}
	}()

	s.deps.Logger.Info("server started", "port", s.deps.Config.Port)
	return server.ListenAndServe()
}
