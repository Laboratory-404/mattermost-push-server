package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Laboratory-404/mattermost-push-server/pkg/config"
	"github.com/Laboratory-404/mattermost-push-server/pkg/metrics"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
)

// Server represents the push proxy HTTP server
type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

// NewServer initializes the HTTP multiplexer and server options
func NewServer(cfg *config.Config, pushService *push.Service) *Server {
	handlers := NewHandlerContainer(pushService)

	mux := http.NewServeMux()

	// Mattermost API Endpoints
	mux.HandleFunc("/api/v1/send_push", handlers.HandleSendPush)
	mux.HandleFunc("/api/v2/send_push", handlers.HandleSendPush)
	mux.HandleFunc("/api/v1/ack", handlers.HandleAck)
	mux.HandleFunc("/api/v2/ack", handlers.HandleAck)

	// Monitoring & Health
	mux.HandleFunc("/healthz", handlers.HandleHealthz)
	mux.HandleFunc("/metrics", metrics.Handler())

	// Root ping
	mux.HandleFunc("/", handlers.HandleAck)

	wrappedMux := SecurityAndLoggingMiddleware(cfg.Server, mux)

	srv := &http.Server{
		Addr:           cfg.Server.ListenAddress,
		Handler:        wrappedMux,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return &Server{
		cfg:        cfg,
		httpServer: srv,
	}
}

// Start runs the HTTP server listening loop
func (s *Server) Start() error {
	log.Printf("[Server] Starting Mattermost Push Proxy Server on %s", s.cfg.Server.ListenAddress)

	if s.cfg.Server.TLSCertFile != "" && s.cfg.Server.TLSKeyFile != "" {
		log.Printf("[Server] Enabling TLS with cert: %s", s.cfg.Server.TLSCertFile)
		if err := s.httpServer.ListenAndServeTLS(s.cfg.Server.TLSCertFile, s.cfg.Server.TLSKeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop gracefully shuts down the server with timeout
func (s *Server) Stop(ctx context.Context) error {
	log.Printf("[Server] Shutting down server gracefully...")
	return s.httpServer.Shutdown(ctx)
}
