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
	"time"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type server struct {
	logger     *slog.Logger
	httpServer *http.Server
	workerPool *workerPool
}

var _ Server = (*server)(nil)

func NewServer() *server {
	return &server{
		logger: slog.Default(),
		workerPool: &workerPool{
			poolSize: 10,
		},
	}
}

func (s *server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.httpServer = &http.Server{
		Addr:    ":3000",
		Handler: http.HandlerFunc(s.handleHttp),
	}

	go startDebugEndpoint()
	go s.listenForInterruption(cancel)
	go s.workerPool.Start(ctx)
	go s.httpServer.ListenAndServe()

	<-ctx.Done()
	s.Stop(ctx)
	return nil
}

func (s *server) handleHttp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/schedule-job" {
		doneCh := make(chan []byte)
		j := Job{
			done: doneCh,
		}
		s.workerPool.Schedule(r.Context(), j)

		_, err := w.Write(<-doneCh)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
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

type Job struct {
	id   int
	done chan []byte
}

type workerPool struct {
	logger     *slog.Logger
	poolSize   int
	jobs       chan Job
	jobIdCount int
}

func (wp *workerPool) Start(ctx context.Context) {
	wp.logger = slog.Default()
	wp.logger.Info("Starting worker pool", "pool_size", wp.poolSize)
	wp.jobs = make(chan Job, wp.poolSize)
	for range wp.poolSize {
		go wp.spawnWorker(ctx)
	}
	wp.logger.Info("Worker has started")
}

func (wp *workerPool) Stop() {
	wp.logger.Info("Closing jobs chanel")
	close(wp.jobs)
}

func (wp *workerPool) Schedule(ctx context.Context, job Job) {
	wp.jobIdCount++
	job.id = wp.jobIdCount
	wp.logger.Info("Schediling job", "id", job.id)
	wp.jobs <- job
	wp.logger.Info("Job has been scheduled", "id", job.id)
}

func (wp *workerPool) spawnWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			wp.logger.Info("Context cancelled in worker")
			return
		case j := <-wp.jobs:
			start := time.Now()
			wp.logger.Info("Processing job", "id", j.id)
			time.Sleep(3 * time.Second)
			wp.logger.Info("Job has finished", "id", j.id, "duration", time.Since(start).String())

			result := fmt.Sprintf(`{"status":"ok","jobId":%d}`, j.id)
			j.done <- []byte(result)
		}
	}
}
