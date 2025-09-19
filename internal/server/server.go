package server

import (
	"context"
	"net/http"
	"time"
)

type Deps struct {
	Port              string
	MaxHeaderBytes    int
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
}

type Server struct {
	httpServer *http.Server
}

func New(handler http.Handler, deps Deps) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + deps.Port,
			Handler:           handler,
			MaxHeaderBytes:    deps.MaxHeaderBytes,
			ReadHeaderTimeout: deps.ReadHeaderTimeout,
			WriteTimeout:      deps.WriteTimeout,
		},
	}
}

func (s *Server) Strart() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
