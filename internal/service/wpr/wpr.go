package wpr

import (
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service"
)

type WPRService struct {
	JobManager       *manager.JobManager
	TllmsManager     *manager.TllmsManager
	TutorPlusService *service.TutorPlusService
}

func (w *WPRService) Execute(userId int64, jobId uuid.UUID, fromDate, toDate, subBatchID int64) {
	//TODO implement me
	panic("implement me")
}
