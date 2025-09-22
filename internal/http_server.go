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
	Port   int
	config Config
	server *netHttp.Server
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
	if err := registerHealthHandler(s.config.HTTP.Health, s.config.MongoDb.URI); err != nil {
		log.Fatal(err.Error())
	}
	registerLivelinessHandler(s.config.HTTP.Liveliness)
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

// Shutdown stops the running server
func (s *HttpServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func registerHealthHandler(path string, mongoUri string) error {
	handler, err := RegisterHealthChecks(mongoUri)
	if err != nil {
		return err
	}
	netHttp.Handle(path, handler)
	return nil
}

func registerLivelinessHandler(path string) {
	netHttp.HandleFunc(path, func(w netHttp.ResponseWriter, request *netHttp.Request) {
		w.WriteHeader(204)
	})
}

func registerPrometheusHandler(path string) {
	netHttp.Handle(path, promhttp.Handler())
}
