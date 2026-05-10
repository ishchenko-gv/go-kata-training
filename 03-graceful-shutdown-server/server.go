package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type server struct {
	logger     *slog.Logger
	httpServer *http.Server
	workerPool WorkerPool
	broker     Broker
}

var _ Server = (*server)(nil)

func NewServer(
	workerPool WorkerPool,
	broker Broker,
	options ...ServerOption,
) *server {
	s := &server{
		logger: slog.Default(),
		httpServer: &http.Server{
			Addr: ":3000",
		},
		workerPool: workerPool,
		broker:     broker,
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

type ServerOption func(s *server)

func WithServerPort(port int) ServerOption {
	return func(s *server) {
		s.httpServer.Addr = fmt.Sprintf(":%d", port)
	}
}

func (s *server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())

	go startDebugEndpoint()
	go s.listenForInterruption(cancel)
	go s.workerPool.Start(ctx)
	go s.broker.Start(ctx)

	s.httpServer.Handler = http.HandlerFunc(s.handleHttp)
	go s.httpServer.ListenAndServe()

	<-ctx.Done()
	s.Stop(ctx)
	return nil
}

func (s *server) handleHttp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/schedule-job" {
		j := Job{}
		s.workerPool.Schedule(r.Context(), j)
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == "/jobs-status" {
		f, ok := w.(http.Flusher)
		if !ok {
			return
		}

		client := make(chan string)
		s.broker.Subscribe(client)
		s.logger.Info("Client connected", "addr", r.RemoteAddr)
		defer func() {
			s.broker.Unsubscribe(client)
			s.logger.Info("Client disconnected", "addr", r.RemoteAddr)
		}()

		w.Header().Set("Content-Type", "text/event-stream")

		for {
			select {
			case <-r.Context().Done():
				return
			case msg := <-client:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				f.Flush()
			}
		}
	}

	http.Error(w, "Route not found", http.StatusNotFound)
}

func startDebugEndpoint() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}

func (s *server) listenForInterruption(cancel context.CancelFunc) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
	s.logger.Info("Got interruption signal")
	cancel()
}

func (s *server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting server down")
	s.httpServer.Shutdown(ctx)
	s.workerPool.Stop()
	return nil
}
