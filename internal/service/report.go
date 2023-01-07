package service

import (
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/mpr"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/tutor_plus"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/wpr"
)

type IReportService interface {
	NewReportGenerator(reportType string, userId int64, jobId, fromDate, toDate, subBatchId string)
}

type IReport interface {
	Execute(userId int64, jobId uuid.UUID, fromDate, toDate, subBatchID int64)
}

type ReportService struct {
	jobManager       *manager.JobManager
	tllmsManager     *manager.TllmsManager
	tutorPlusService *tutor_plus.TutorPlusService
}

func ReportServiceInitializer(jobManager *manager.JobManager, tllmsManager *manager.TllmsManager,
	tutorPlusService *tutor_plus.TutorPlusService) *ReportService {
	return &ReportService{jobManager: jobManager, tllmsManager: tllmsManager, tutorPlusService: tutorPlusService}
}

func (rS *ReportService) NewReportGenerator(reportType string, userId int64, jobId uuid.UUID, fromDate, toDate, subBatchId int64) {
	var r IReport
	if reportType == constant.MPR {
		r = &mpr.MPRService{JobManager: rS.jobManager, TllmsManager: rS.tllmsManager, TutorPlusService: rS.tutorPlusService}
	} else if reportType == constant.WEEKLY_REPORT {
		r = &wpr.WPRService{JobManager: rS.jobManager, TllmsManager: rS.tllmsManager, TutorPlusService: rS.tutorPlusService}
	}
	r.Execute(userId, jobId, fromDate, toDate, subBatchId)
}
