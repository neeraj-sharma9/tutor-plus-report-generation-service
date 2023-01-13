package mpr

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service"
	"strconv"
	"sync"
)

type MPRService struct {
	JobManager       *manager.JobManager
	TllmsManager     *manager.TllmsManager
	TutorPlusService *service.TutorPlusService
}

func (m *MPRService) Execute(userId int64, jobId uuid.UUID, fromDate, toDate, subBatchID int64) {
	args := map[string]interface{}{
		"userId": userId,
		"jobId":  jobId,
		"params": map[string]interface{}{
			"fromDate": fromDate,
			"toDate":   toDate,
		},
	}
	m.JobManager.CreateJob(userId, jobId, constant.RETRY, args["params"], constant.MPR)
	data := m.GetMPRData(args)
	fmt.Println(data)
}

func (m *MPRService) GetMPRData(args map[string]interface{}) interface{} {
	mprReq := m.newMprReq(args)
	pd := &ProgressData{TllmsManager: m.TllmsManager}
	userDetails := m.TutorPlusService.GetUserDetails(mprReq)
	mprReq.UserDetailsResponse = userDetails
	if mprReq.EpchFrmDate <= 0 {
		mprReq.EpchFrmDate = userDetails.ReportMonth.FromDateEpoch
	}
	if mprReq.EpchToDate <= 0 {
		mprReq.EpchToDate = userDetails.ReportMonth.ToDateEpoch
	}

	params := make(map[string]string)
	params["toDate"] = strconv.FormatInt(userDetails.ReportMonth.ToDateEpoch, 10)
	params["fromDate"] = strconv.FormatInt(userDetails.ReportMonth.FromDateEpoch, 10)

	jobID := mprReq.JobId
	updateArgs := map[string]interface{}{
		"state":   constant.ADD_START_AND_END_DATE,
		"params":  params,
		"comment": "",
		"success": true,
	}
	m.JobManager.UpdateJob(jobID, updateArgs)

	if mprReq.ReqStatus {
		var wg sync.WaitGroup
		wg.Add(6)
		go pd.setCoverDetails(mprReq, &wg)
		go pd.setUserDetails(mprReq, &wg)
		go pd.setSummaryPage(mprReq, &wg)
		go pd.setSubjectWisePage(mprReq, &wg)
		go pd.setLookingAheadDetails(mprReq, &wg)
		go pd.SetMonthlyTestPerformance(mprReq, &wg)
		wg.Wait()
	}
	return nil
}

func (m *MPRService) newMprReq(args map[string]interface{}) *helper.MPRReq {
	mprReq := &helper.MPRReq{UserId: args["userId"].(int64), ReqStatus: true, JobId: args["jobId"].(uuid.UUID)}

	if val, ok := args["fromDate"]; ok {
		mprReq.EpchFrmDate = val.(int64)
	} else {
		logger.Log.Sugar().Errorf("unable to fetch from_date: %v", args["fromDate"])
	}

	if val, ok := args["toDate"]; ok {
		mprReq.EpchToDate = val.(int64)
	} else {
		logger.Log.Sugar().Errorf("unable to fetch from_date: %v", args["fromDate"])
	}
	return mprReq
}

func (report *ProgressData) setCoverDetails(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setCoverDetails paniced - %s", r)
		}
		wg.Done()
	}()
	cover := &DateRange{}
	cover.ToDate = mprReq.UserDetailsResponse.ReportMonth.ToDate
	cover.FromDate = mprReq.UserDetailsResponse.ReportMonth.FromDate
	report.Cover = *cover
}

func (report *ProgressData) setUserDetails(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setUserdetails paniced - %s", r)
		}
		wg.Done()
	}()
	report.UserInfo = mprReq.UserDetailsResponse.UserInfo
}

func (report *ProgressData) setSubjectWisePage(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSubjectWisePage paniced - %s", r)
		}
		wg.Done()
	}()

	var subjectWisePerformance []SubjectWisePerformance
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		subjectWiseData := &SubjectWisePerformance{}
		subjectWiseData.Subject = neoSubject.Subject
		subjectWisePerformance = append(subjectWisePerformance, *subjectWiseData)
	}
	report.SubjectWisePerformance = subjectWisePerformance
	var waitGroup sync.WaitGroup
	waitGroup.Add(5)
	go report.SetClassAttendance(mprReq, &waitGroup)
	go report.SetSessionWiseBreakdown(mprReq, &waitGroup)
	go report.SetPerformanceTillDate(mprReq, &waitGroup)
	go report.SetClassQuiz(mprReq, &waitGroup)
	go report.SetSubjectWiseAssignments(mprReq, &waitGroup)
	waitGroup.Wait()
	report.SetInClassCallout(mprReq)
	report.SetSummaryPageAssignments(mprReq)
	report.SetSummaryLearnerTags(mprReq)
}

func (report *ProgressData) setLookingAheadDetails(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setLookingAheadDetails paniced - %s", r)
		}
		wg.Done()
	}()
	report.LookingAhead = mprReq.UserDetailsResponse.LookingAhead
}
