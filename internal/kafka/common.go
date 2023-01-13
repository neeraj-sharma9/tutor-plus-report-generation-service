package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/config"
	"os"
)

type KafkaConf struct {
	Brokers []string
	Config  *sarama.Config
}

type Job struct {
	UserID     int64     `json:"user_id"`
	JobID      uuid.UUID `json:"job_id"`
	FromDate   int64     `json:"from_date"`
	ToDate     int64     `json:"to_date"`
	CohortList string    `json:"cohort_list"`
	SubBatchID int64     `json:"sub_batch_id"`
}

func KafkaConfInitializer() *KafkaConf {
	kafkaConf := config.GetKafkaConfig()
	kafka := KafkaConf{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Config:  kafkaConf,
	}
	return &kafka
}
