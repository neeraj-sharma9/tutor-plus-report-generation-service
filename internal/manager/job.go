package manager

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/config"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	logger "github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/model"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/utility"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type IJobManager interface {
}

type JobManager struct {
	db *gorm.DB
}

func JobManagerInitializer(config *config.Config, zapLogger *zap.Logger) *JobManager {
	dsn := utility.GetPostgresDSN(config.JobDbHost, config.JobDbPort, config.JobDbName, config.JobDbUser, config.JobDbPassword)
	pg, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		zapLogger.Sugar().Errorf("error connecting job db: %v", err)
	}
	return &JobManager{db: pg}
}

func (jM *JobManager) CreateJob(userId int64, jobId uuid.UUID, retry int16, params interface{}, reportType string) error {
	paramsJson, err := json.Marshal(params)
	if err != nil {
		logger.Log.Sugar().Errorf("error marshaling params to json: %v, for jobId: %v, for params: %v",
			err, jobId, params)
		return err
	}
	job := model.Job{Id: jobId, UserId: userId, Retry: retry, ReportType: reportType,
		State: constant.JOB_RECEIVED, Params: paramsJson}
	if err := jM.db.Model(&model.Job{}).Create(&job).Error; err != nil {
		logger.Log.Sugar().Errorf("error inserting job in the db: %v, for job: %+v", err, job)
		return err
	}
	return nil
}

func (jM *JobManager) UpdateJob(jobId uuid.UUID, args map[string]interface{}) error {
	if err := jM.db.Model(&model.Job{}).Where(&model.Job{Id: jobId}).Updates(args).Error; err != nil {
		logger.Log.Sugar().Errorf("Error updating the job: %+v, with id: %v and args %+v", err, jobId, args)
		return err
	}
	return nil
}
