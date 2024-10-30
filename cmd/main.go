package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keans/wali/internal/database"
	"github.com/keans/wali/internal/models"
	"github.com/keans/wali/internal/utils"
	"github.com/keans/wali/internal/yaml"

	"github.com/alecthomas/kong"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	tickCount int
)

type Context struct {
	Config     utils.AppConfig
	db         *database.Database
	smtp       *utils.Smtp
	workerPool *models.WorkerPool
	log        *slog.Logger
}

type CLI struct {
	Run RunCmd `cmd:"" help:"Run tool"`
}

type RunCmd struct {
	YamlFile string `arg:"" name:"yamlfile" help:"YAML file that is read." type:"path"`
}

func (c *Context) OnTick(t time.Time) {
	if tickCount == c.Config.ShowTickEvery {
		// show tick that application is still alive and reset counter
		c.log.Info("tick", "at", t)
		tickCount = 0
	} else {
		tickCount++
	}

	// get all jobs from database
	jobs, err := c.db.GetAllJobs()
	if err != nil {
		panic(err)
	}

	for _, j := range jobs {
		if j.IsExceeded() {
			// job exceeded => enqueue
			c.log.Info("job exceeded", "job", j.Key)

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
	var waliYaml yaml.WaliYaml
	if err := yaml.ReadYaml(c.YamlFile, &waliYaml); err != nil {
		return err
	}

	// prepare smtp server
	ctx.smtp = utils.NewSmtp(waliYaml.Smtp.Host, waliYaml.Smtp.Port,
		waliYaml.Smtp.Username, waliYaml.Smtp.Password,
		waliYaml.Smtp.From, waliYaml.Smtp.To)

	// prepare logger
	ctx.log = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// initialize the database and reset jobs statuses
	ctx.db = database.NewDb(ctx.Config.DbFilename)
	if err := ctx.db.Open(); err != nil {
		return err
	}
	defer ctx.db.Close()

	// ensure that database files are available
	if err := ctx.db.CreateTables(); err != nil {
		return err
	}

	// add all webjobs from YAML to the database (or update if existing)
	ctx.db.AddFromYaml(&waliYaml)

	// ensure that all jobs are marked as 'stopped' on startup to avoid
	// problems of ungracefully shutdown artefacts
	ctx.db.ResetJobsStatuses()

	// prepare worker pool
	var err error
	ctx.workerPool, err = models.NewWorkerPool(ctx.Config.WorkersCount,
		ctx.db, ctx.log, ctx.smtp, true)
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

	return nil
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli)
	err := ctx.Run(&Context{})
	ctx.FatalIfErrorf(err)
}
