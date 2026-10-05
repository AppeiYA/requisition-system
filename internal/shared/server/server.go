package server

import (
	"context"
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/config"
)

type Server struct {
	httpServer *http.Server
}

func New(handler http.Handler, cfg *config.ServerConfig) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Address,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
