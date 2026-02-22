package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/router"
)

// Server is the AI Router HTTP server.
type Server struct {
	httpServer *http.Server
	router     *router.Router
	catalog    *router.ModelCatalog
	logger     *slog.Logger
}

// New creates a new server.
func New(host string, port int, r *router.Router, catalog *router.ModelCatalog, logger *slog.Logger) *Server {
	s := &Server{
		router:  r,
		catalog: catalog,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/models", s.handleListModels)
	mux.HandleFunc("POST /v1/models/refresh", s.handleRefreshModels)
	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("POST /v1/embeddings", s.handleEmbed)
	mux.HandleFunc("POST /v1/chat/completions", s.handleChatCompletions)
	mux.HandleFunc("POST /v1/test", s.handleTest)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      s.withMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start begins listening. It blocks until the server is shut down.
func (s *Server) Start() error {
	s.logger.Info("Server listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
