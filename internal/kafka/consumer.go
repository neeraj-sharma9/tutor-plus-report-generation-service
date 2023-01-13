package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/utility"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/kafkapubsub"
)

func RunKafkaConsumers(conf *KafkaConf, reportService *service.ReportService) {
	sem := utility.CreateSemaphore(constant.ASYNC_KAFKA_CONSUMERS)
	subscriptions, err := getKafkaSubscriptions(conf, constant.REPORT_TOPICS)
	if err != nil {
		logger.Log.Sugar().Errorf("error getting kafka subscriptions: %v, for topics: %v", err, constant.REPORT_TOPICS)
		return
	}
	defer closeKafkaSubscriptions(subscriptions)
	i := 0
	topicsCount := len(constant.REPORT_TOPICS)
	for {
		<-sem
		go func() {
			consume(constant.REPORT_TOPICS[i], subscriptions[i], reportService)
			sem <- 1
		}()
		i++
		i = i % topicsCount
	}
}

func consume(topic string, subscription *pubsub.Subscription, reportService *service.ReportService) {
	msg, err := subscription.Receive(context.Background())
	if err != nil {
		logger.Log.Sugar().Errorf("error receiving msg from topic: %v, error: %v", topic, err)
		return
	}
	msg.Ack()
	var reportType string
	switch topic {
	case constant.MPR_JOB_TOPIC, constant.MPR_PRIORITY_JOB_TOPIC, constant.MPR_RETRY_JOB_TOPIC:
		reportType = constant.MPR
	case constant.WEEKLY_REPORT_JOB_TOPIC, constant.WEEKLY_REPORT_PRIORITY_JOB_TOPIC, constant.WEEKLY_REPORT_RETRY_JOB_TOPIC:
		reportType = constant.WEEKLY_REPORT
	default:
		reportType = constant.MPR
	}

	var job Job
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		logger.Log.Sugar().Errorf("json unmarshal error: %v, for json: %v", err, string(msg.Body))
		return
	}
	reportService.NewReportGenerator(reportType, job.UserID, job.JobID, job.FromDate, job.ToDate, job.SubBatchID)

}

func closeKafkaSubscriptions(subscriptions []*pubsub.Subscription) {
	for _, subsciption := range subscriptions {
		subsciption.Shutdown(context.Background())
	}
}

func getKafkaSubscriptions(conf *KafkaConf, topics []string) ([]*pubsub.Subscription, error) {
	var subscriptions []*pubsub.Subscription
	return subscriptions, fmt.Errorf("error custom")
	for _, topic := range topics {
		subscription, err := kafkapubsub.OpenSubscription(conf.Brokers, conf.Config,
			constant.KAFKA_CONSUMER_GROUP, []string{topic}, &kafkapubsub.SubscriptionOptions{})
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}
