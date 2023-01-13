package cmd

import (
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/worker"
	"github.com/urfave/cli/v2"
)

var WorkerStartCommand = &cli.Command{
	Name:    "start",
	Aliases: []string{"up"},
	Usage:   "Starts report generation service",
	Action:  startAction,
}

func startAction(ctx *cli.Context) (err error) {
	worker.Run(ctx.Context)
	return nil
}
