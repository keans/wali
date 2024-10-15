package models

import (
	"os"
	"sync"
	"time"
	"wali/internal/database"

	"log/slog"
)

type Job interface {
	Execute(db *database.Database) bool
}

type WorkerPool struct {
	workerCount int
	wg          sync.WaitGroup
	jobQueue    chan Job
	isRunning   bool
	log         *slog.Logger
	db          *database.Database
}

func NewWorkerPool(workerCount int, dbFilename string) (*WorkerPool, error) {
	wp := &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, 1),
		isRunning:   true,
		log:         slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		db:          database.NewDb(dbFilename),
	}

	if err := wp.db.Open(); err != nil {
		return nil, err
	}

	if err := wp.db.CreateTables(); err != nil {
		return nil, err
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
}

func (wp *WorkerPool) Enqueue(job Job) {
	wp.log.Info("enqueuing job")
	wp.jobQueue <- job
}

func (wp *WorkerPool) worker(workerId int, ch chan bool) {
	defer wp.wg.Done()

	wp.log.Info("starting worker", "workerId", workerId)

	maxWait := 3
out:
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
			job.Execute(wp.db)
			wp.log.Info("job completed", "workerId", workerId)

			// reset wait mechanism
			maxWait = 3

		default:
			// wait for
			time.Sleep(time.Duration(1 * time.Second))
			if maxWait == 0 {
				wp.log.Info("no job in queue")
				break out
			} else {
				maxWait--
				wp.log.Debug("waiting", "maxWait", maxWait)
			}
		}
	}

	ch <- true

	wp.log.Info("stopping worker", "workerId", workerId)
}
