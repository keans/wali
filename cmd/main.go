package main

import (
	"fmt"
	"log/slog"
	"wali/internal/models"
	"wali/internal/utils"

	"github.com/alecthomas/kong"
	"github.com/ilyakaznacheev/cleanenv"
)

type Context struct {
	Config utils.AppConfig
}

type CLI struct {
	Run RunCmd `cmd:"" help:"Run tool"`
}

type RunCmd struct {
	YamlFile string `arg:"" name:"yamlfile" help:"YAML file that is read." type:"path" default:"wali.yaml"`
}

func (a *RunCmd) Run(ctx *Context) error {
	var webjobs models.WebJobs

	// read jobs from YAML file
	if err := utils.ReadYaml(a.YamlFile, &webjobs); err != nil {
		return err
	}

	// read config from environment variables
	if err := cleanenv.ReadEnv(&ctx.Config); err != nil {
		return err
	}

	// prepare worker pool
	wp, err := models.NewWorkerPool(ctx.Config.WorkersCount,
		ctx.Config.DbFilename)
	if err != nil {
		return err
	}

	// start worker pool
	wp.Start()

	// add all jobs obtained from YAML
	for _, job := range webjobs.WebJobs {
		fmt.Println(job)
		if job.IsValid() {
			wp.Enqueue(&job)
		} else {
			slog.Error("skipping", "webJob", job)
		}
	}

	// wait until jobs are completed
	wp.Wait()

	// finally shutdown
	wp.Shutdown()

	return nil
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli)
	err := ctx.Run(&Context{})
	ctx.FatalIfErrorf(err)
}
