package models

import (
	"sync"
	"time"

	"github.com/keans/wali/internal/database"
	"github.com/keans/wali/internal/utils"

	"log/slog"
)

type Job interface {
	Execute(db *database.Database, smtp *utils.Smtp, log *slog.Logger) bool
}

type WorkerPool struct {
	workerCount int
	wg          sync.WaitGroup
	jobQueue    chan Job
	isRunning   bool
	log         *slog.Logger
	db          *database.Database
	smtp        *utils.Smtp
}

func NewWorkerPool(workerCount int, db *database.Database, log *slog.Logger,
	smtp *utils.Smtp, autoStart bool) (*WorkerPool, error) {

	wp := &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, 1),
		isRunning:   true,
		log:         log,
		db:          db,
		smtp:        smtp,
	}

	if autoStart {
		wp.Start()
	}

	return wp, nil
}

func (wp *WorkerPool) Start() {
	ch := make(chan bool, wp.workerCount)

	// start all workers
	wp.log.Info("starting worker pool", "workerCount", wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i, ch)
	}
}

func (wp *WorkerPool) Wait() {
	wp.log.Info("waiting for workers to complete their work",
		"workerCount", wp.workerCount)
	wp.wg.Wait()
	wp.log.Info("worker pool completed", "workerCount", wp.workerCount)
}

func (wp *WorkerPool) Shutdown() {
	wp.log.Info("shutting down worker pool", "workerCount", wp.workerCount)
	wp.isRunning = false
	wp.db.Close()

	// wait until all workers are shutdown
	wp.Wait()
}

func (wp *WorkerPool) Enqueue(job Job) {
	// some sophisticated logging, depending on type of job interface
	switch v := job.(type) {
	case *database.Job:
		wp.log.Info("enqueuing web job", "webJob", job.(*database.Job))
	default:
		wp.log.Warn("enqueuing unknown job type", "jobType", v)
	}

	wp.jobQueue <- job
}

func (wp *WorkerPool) worker(workerId int, ch chan bool) {
	defer wp.wg.Done()

	wp.log.Debug("starting worker", "workerId", workerId)

	for {
		if !wp.isRunning {
			// not running => simply stop worker
			wp.log.Info("worker pool has quitted", "workerId", workerId)
			break
		}

		select {
		case job := <-wp.jobQueue:
			// get job from queue and execute it
			wp.log.Info("executing job", "workerId", workerId)
			job.Execute(wp.db, wp.smtp, wp.log)
			wp.log.Info("job completed", "workerId", workerId)

		default:
			// wait for 1s, if there are no jobs
			time.Sleep(time.Duration(1 * time.Second))
		}
	}

	ch <- true

	wp.log.Debug("stopping worker", "workerId", workerId)
}
