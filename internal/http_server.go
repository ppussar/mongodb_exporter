package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	netHttp "net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HttpServer serves endpoints from the given Config
type HttpServer struct {
	Port          int
	config        Config
	server        *netHttp.Server
	healthCloser  func(context.Context) error
}

// NewHttpServer creates a new instance of the HttpServer
func NewHttpServer(config Config) *HttpServer {
	return &HttpServer{
		config: config,
	}
}

// Start the HTTP server
// Returns a WaitGroup which will be released as soon as the server stops
func (s *HttpServer) Start(wg *sync.WaitGroup) {
	closer, err := registerHealthHandler(s.config.HTTP.Health, s.config.MongoDb.URI)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.healthCloser = closer
	registerLivenessHandler(s.config.HTTP.Liveness)
	registerPrometheusHandler(s.config.HTTP.Prometheus)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.HTTP.Port))
	if err != nil {
		log.Fatal(err.Error())
	}
	s.Port = listener.Addr().(*net.TCPAddr).Port
	s.server = &netHttp.Server{}

	go func() {
		defer wg.Done()
		defer log.Info("Stopping server")
		log.Info(fmt.Sprintf("Serving endpoint on port: %v", s.Port))
		if err := s.server.Serve(listener); !errors.Is(err, netHttp.ErrServerClosed) {
			log.Fatal(fmt.Sprintf("ListenAndServe(): %v", err))
		}
	}()
}

// Shutdown stops the running server and closes the health-check MongoDB connection.
func (s *HttpServer) Shutdown(ctx context.Context) error {
	if s.healthCloser != nil {
		if err := s.healthCloser(ctx); err != nil {
			log.Error(fmt.Sprintf("health check client disconnect error: %v", err))
		}
	}
	return s.server.Shutdown(ctx)
}

func registerHealthHandler(path string, mongoUri string) (func(context.Context) error, error) {
	handler, closer, err := RegisterHealthChecks(mongoUri)
	if err != nil {
		return nil, err
	}
	netHttp.Handle(path, handler)
	return closer, nil
}

func registerLivenessHandler(path string) {
	netHttp.HandleFunc(path, func(w netHttp.ResponseWriter, request *netHttp.Request) {
		w.WriteHeader(204)
	})
}

func registerPrometheusHandler(path string) {
	netHttp.Handle(path, promhttp.Handler())
}
