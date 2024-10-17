package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wali/internal/database"
	"wali/internal/models"
	"wali/internal/utils"
	"wali/internal/yaml"

	"github.com/alecthomas/kong"
	"github.com/ilyakaznacheev/cleanenv"
)

type Context struct {
	Config     utils.AppConfig
	db         *database.Database
	workerPool *models.WorkerPool
}

type CLI struct {
	Run RunCmd `cmd:"" help:"Run tool"`
}

type RunCmd struct {
	YamlFile string `arg:"" name:"yamlfile" help:"YAML file that is read." type:"path" default:"wali.yaml"`
}

func (c *Context) OnTick(t time.Time) {
	fmt.Println("Tick at", t)

	jobs, err := c.db.GetAllJobs()
	if err != nil {
		panic(err)
	}

	for _, j := range jobs {
		if j.IsExceeded() {
			// job exceeded => enqueue
			fmt.Println("exceed", j)

			// add to queue
			c.workerPool.Enqueue(j)

			// mark for enqueued in database
			j.Status = database.Enqueued
			if err := c.db.UpdateJob(j); err != nil {
				panic(err)
			}
		}

	}
}

func (c *RunCmd) Run(ctx *Context) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// read config from environment variables
	if err := cleanenv.ReadEnv(&ctx.Config); err != nil {
		return err
	}

	// read jobs from YAML file
	var webjobs yaml.Jobs
	if err := yaml.ReadYaml(c.YamlFile, &webjobs); err != nil {
		return err
	}

	// initialize the database and reset jobs statuses
	ctx.db = database.NewDb(ctx.Config.DbFilename)
	ctx.db.Open()
	defer ctx.db.Close()

	// ensure that database files are available
	if err := ctx.db.CreateTables(); err != nil {
		return err
	}

	// add all webjobs from YAML to the database (or update if existing)
	ctx.db.AddFromYaml(&webjobs)

	// ensure that all jobs are marked as 'stopped' on startup to avoid
	// problems of ungracefully shutdown artefacts
	ctx.db.ResetJobsStatuses()

	// prepare worker pool
	var err error
	ctx.workerPool, err = models.NewWorkerPool(ctx.Config.WorkersCount,
		ctx.db, true)
	if err != nil {
		return err
	}
	defer ctx.workerPool.Shutdown()

	// start the schedule in a go subroutine
	var sched *models.Scheduler
	go func() {
		// start the scheduler
		sched = models.NewScheduler(1000, ctx.OnTick, true)

	}()
	sig := <-signalChan
	log.Printf("%s signal caught", sig)

	// stop the scheduler
	sched.Stop()

	// // add all jobs obtained from YAML
	// for _, job := range webjobs.WebJobs {
	// 	fmt.Println(job)
	// 	if job.IsValid() {
	// 		wp.Enqueue(&job)
	// 	} else {
	// 		slog.Error("skipping", "webJob", job)
	// 	}
	// }

	return nil
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli)
	err := ctx.Run(&Context{})
	ctx.FatalIfErrorf(err)
}
