package wpr

import (
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/tutor_plus"
)

type WPRService struct {
	JobManager       *manager.JobManager
	TllmsManager     *manager.TllmsManager
	TutorPlusService *tutor_plus.TutorPlusService
}

func (w *WPRService) Execute(userId int64, jobId uuid.UUID, fromDate, toDate, subBatchID int64) {
	//TODO implement me
	panic("implement me")
}
