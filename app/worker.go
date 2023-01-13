package app

import (
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/cmd"
	"github.com/urfave/cli/v2"
)

func WorkerApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Student report generation application"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		cmd.WorkerStartCommand,
	}

	return app
}
