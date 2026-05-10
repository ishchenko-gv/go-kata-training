package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Job struct {
	id int
}

type WorkerPool interface {
	Start(ctx context.Context)
	Stop()
	Schedule(ctx context.Context, job Job)
}

type workerPool struct {
	logger     *slog.Logger
	poolSize   int
	msgCh      chan string
	jobs       chan Job
	jobIdCount int
}

func NewWorkerPool(msgCh chan string, poolSize int) *workerPool {
	return &workerPool{
		logger:   slog.Default(),
		msgCh:    msgCh,
		poolSize: poolSize,
	}
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
	wp.logger.Info("Scheduling job", "id", job.id)
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
			wp.msgCh <- result
		}
	}
}
