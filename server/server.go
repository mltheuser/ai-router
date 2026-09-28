// Package server exposes the use cases over HTTP. It knows no use case in
// particular: every route is derived from the usecase.UseCase interface.
package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/debug"
	"github.com/mltheuser/ai-router/usecase"
)

// Config is what the server serves.
type Config struct {
	// Addr is the host:port to listen on.
	Addr string
	// UseCases are served at /v1/<name>.
	UseCases []usecase.UseCase
	// Debug, if set, receives a full request/response log of every use-case
	// request.
	Debug io.Writer
}

// Server is the ai-router HTTP server.
type Server struct {
	httpServer *http.Server
	cfg        Config
}

// New creates a server. Call Start to serve.
func New(cfg Config) *Server {
	s := &Server{cfg: cfg}

	withDebug := func(h http.Handler) http.Handler { return h }
	if cfg.Debug != nil {
		withDebug = debug.Middleware(cfg.Debug)
	}

	mux := http.NewServeMux()
	for _, uc := range cfg.UseCases {
		prefix := "/v1/" + uc.Name()
		mux.Handle("POST "+prefix, withDebug(handle(uc.Handle)))
		mux.Handle("GET "+prefix+"/models", handle(uc.ListModels))
	}
	mux.Handle("POST /v1/test", handle(s.handleTest))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		api.WriteJSON(w, map[string]string{"status": "ok"})
	})

	s.httpServer = &http.Server{
		Addr:         cfg.Addr,
		Handler:      logRequests(mux),
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}
	return s
}

// Start serves until the server is shut down.
func (s *Server) Start() error {
	slog.Info("Server listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server")
	return s.httpServer.Shutdown(ctx)
}

// handle adapts a handler that returns its error to an http.Handler that
// writes the error as the JSON error response.
func handle(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			slog.Warn("Request failed", "path", r.URL.Path, "error", err)
			api.WriteError(w, err)
		}
	})
}

// logRequests logs one line per request.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Unwrap exposes the wrapped writer to http.ResponseController.
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
