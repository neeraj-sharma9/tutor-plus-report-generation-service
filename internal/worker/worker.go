package worker

import (
	"context"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/config"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/kafka"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/tutor_plus"
	"go.uber.org/fx"
)

func workerFxApp(ctx context.Context) *fx.App {
	return fx.New(
		fx.Provide(func() context.Context { return ctx }),
		fx.Provide(config.ConfigInitializer),
		fx.Provide(logger.LogInitializer),
		fx.Provide(manager.JobManagerInitializer),
		fx.Provide(manager.TllmsManagerInitializer),
		fx.Provide(tutor_plus.TutorPlusServiceInitializer),
		fx.Provide(service.ReportServiceInitializer),
		fx.Provide(kafka.KafkaConfInitializer),
		fx.Invoke(kafka.RunKafkaConsumers),
	)
}

func Run(ctx context.Context) {
	workerFxApp(ctx).Run()
}
