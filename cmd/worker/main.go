package main

import (
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/app"
	"log"
	"os"
)

func main() {
	workerApp := app.WorkerApp()
	if err := workerApp.Run(os.Args); err != nil {
		log.Println(err)
	}
}
