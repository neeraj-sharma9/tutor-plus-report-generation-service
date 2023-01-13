package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	APIKey                string `mapstructure:"API_KEY"`
	APISecret             string `mspstructure:"API_SECRET"`
	ServerAddress         string `mapstructure:"SERVER_ADDRESS"`
	CronAccessToken       string `mapstructure:"CRON_ACCESS_TOKEN"`
	SchedulerAccessToken  string `mapstructure:"SCHEDULER_ACCESS_TOKEN"`
	SchedulerHost         string `mapstructure:"SCHEDULER_HOST"`
	SchedulerCallbackHost string `mapstructure:"SCHEDULER_CALLBACK_HOST"`
	MaxRetry              int16  `mapstructure:"MAX_FAILURE_RETRY"`
	JobDefinition         string `mapstructure:"BATCH_JOB_DEFINATION"`
	JobQueue              string `mapstructure:"BATCH_JOB_QUEUE"`
	BatchRegion           string `mapstructure:"BATCH_REGION"`
	RecordLimit           string `mapstructure:"RECORD_LIMIT"`
	JobDbDriver           string `mapstructure:"JOB_DB_DRIVER"`
	MaxWorker             int32  `mapstructure:"MAX_WORKER"`
	MaxQueue              int32  `mapstructure:"MAX_QUEUE_SIZE"`
	JobDbHost             string `mapstructure:"JOB_DB_HOST"`
	JobDbPort             int32  `mapstructure:"JOB_DB_PORT"`
	JobDbUser             string `mapstructure:"JOB_DB_USER"`
	JobDbPassword         string `mapstructure:"JOB_DB_PASSWORD"`
	JobDbName             string `mapstructure:"JOB_DB_NAME"`
	ReplicaDbHost         string `mapstructure:"TLLMS_REPLICA_DB_HOST"`
	ReplicaDbPort         int32  `mapstructure:"TLLMS_REPLICA_DB_PORT"`
	ReplicaDbUser         string `mapstructure:"TLLMS_REPLICA_DB_USER"`
	ReplicaDbPassword     string `mapstructure:"TLLMS_REPLICA_DB_PASSWORD"`
	ReplicaDbName         string `mapstructure:"TLLMS_REPLICA_DB_NAME"`
	CohortList            string `mapstructure:"ENABLED_COHORT_LIST"`
	ConsumerQuit          bool   `mapstructure:"QUIT"`
	SqsMaxMessage         int64  `mapstructure:"MAX_MESSAGES"`
	MaxConsumer           int64  `mapstructure:"MAX_CONSUMERS"`
	SqlMinMessage         int64  `mapstructure:"MIN_MESSAGES"`
	MaxPollMessage        int64  `mapstructure:"MAX_MESSAGES_FROM_QUEUE"`
	VisibilityTimeout     int64  `mapstructure:"VISIBILITY_TIMEOUT_SECONDS"`
	SqsUrl                string `mapstructure:"QUEUE_URL"`
	SqsRegion             string `mapstructure:"QUEUE_REGION"`
	SFKafkaAddress        string `mapstructure:"SF_KAFKA_ADDRESS"`
	SFMPRTopic            string `mapstructure:"SF_MPR_TOPIC"`
	ReportMonth           int    `mapstructure:"REPORT_MONTH"`
	KafkaBroker           string `mapstructure:"KAFKA_BROKER"`
	KafkaUsername         string `mapstructure:"KAFKA_USERNAME"`
	KafkaPassword         string `mapstructure:"KAFKA_PASSWORD"`
	TutorPlusBaseURL      string `mapstructure:"TUTOR_PLUS_BASE_URL"`
}

func ConfigInitializer() (*Config, error) {
	config := Config{}
	viper.AddConfigPath("./")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return &config, err
	}

	err = viper.Unmarshal(&config)
	return &config, err
}
