package mpr

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/tutor_plus"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/utility"
	"strconv"
	"sync"
)

type MPRService struct {
	JobManager       *manager.JobManager
	TllmsManager     *manager.TllmsManager
	TutorPlusService *tutor_plus.TutorPlusService
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
	params := make(map[string]string)
	params["toDate"] = strconv.FormatInt(fromDate, 10)
	params["fromDate"] = strconv.FormatInt(toDate, 10)
	updateArgs := map[string]interface{}{
		"state":   constant.JOB_RECEIVED,
		"params":  params,
		"comment": "",
	}
	m.JobManager.UpdateJob(jobId, updateArgs)
	data := m.GetMPRData(args)
	fmt.Println(data)
}

func (m *MPRService) GetMPRData(args map[string]interface{}) interface{} {
	mprReq := m.newMprReq(args)
	pd := &ProgressData{TllmsManager: m.TllmsManager}
	mprReq.ReqStatus, mprReq.State, mprReq.ErrorMsg,
		mprReq.UserDetailsResponse = m.TutorPlusService.GetUserDetails(mprReq.UserId, mprReq.EpchFrmDate, mprReq.EpchToDate)
	if mprReq.EpchFrmDate <= 0 {
		mprReq.EpchFrmDate = mprReq.UserDetailsResponse.ReportMonth.FromDateEpoch
	}
	if mprReq.EpchToDate <= 0 {
		mprReq.EpchToDate = mprReq.UserDetailsResponse.ReportMonth.ToDateEpoch
	}

	params := make(map[string]string)
	params["toDate"] = strconv.FormatInt(mprReq.UserDetailsResponse.ReportMonth.ToDateEpoch, 10)
	params["fromDate"] = strconv.FormatInt(mprReq.UserDetailsResponse.ReportMonth.FromDateEpoch, 10)

	jobID := mprReq.JobId
	updateArgs := map[string]interface{}{
		"state":   constant.ADD_START_AND_END_DATE,
		"params":  params,
		"comment": "",
		"success": true,
	}
	m.JobManager.UpdateJob(jobID, updateArgs)

	if mprReq.ReqStatus {
		pd.Cover = m.getCoverDetails(mprReq)
		pd.UserInfo = mprReq.UserDetailsResponse.UserInfo
		pd.SubjectWisePerformance = m.getSubjectWisePage(mprReq)
		pd.SummaryOfLearning = m.getSummaryPage(mprReq, *pd)
		pd.LookingAhead = m.getLookingAheadDetails(mprReq)
		pd.MonthlyTestPerformance = m.getMonthlyTestPerformance(mprReq)
	}
	return nil
}

func (m *MPRService) newMprReq(args map[string]interface{}) *MPRReq {
	mprReq := &MPRReq{UserId: args["userId"].(int64), ReqStatus: true, JobId: args["jobId"].(uuid.UUID)}

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

func (m *MPRService) getCoverDetails(mprReq *MPRReq) DateRange {
	cover := DateRange{}
	cover.ToDate = mprReq.UserDetailsResponse.ReportMonth.ToDate
	cover.FromDate = mprReq.UserDetailsResponse.ReportMonth.FromDate
	return cover
}

func (m *MPRService) getSummaryPage(mprReq *MPRReq, pd ProgressData) SummaryOfLearning {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("getSummaryPage paniced - %s", r)
		}
	}()
	summary := SummaryOfLearning{}
	summaryService := SummaryService{TllmsManager: m.TllmsManager}
	summary.ChaptersCovered = summaryService.GetChaptersCovered(mprReq)
	summary.ClassAttendance = summaryService.GetClassAttendance(mprReq)
	summary.Assignments = summaryService.GetSummaryPageAssignments(mprReq, pd.SubjectWisePerformance)
	summary.LearnerTags = summaryService.GetSummaryLearnerTags(mprReq, pd)
	return summary
}

func (m *MPRService) getSubjectWisePage(mprReq *MPRReq) []SubjectWisePerformance {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("getSubjectWisePage paniced - %s", r)
		}
	}()

	var subjectWisePerformance []SubjectWisePerformance
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		subjectWiseData := &SubjectWisePerformance{}
		subjectWisePerformanceService := SubjectWisePerformanceService{TllmsManager: m.TllmsManager, SWP: subjectWiseData}
		subjectWiseData.Subject = neoSubject.Subject
		var waitGroup sync.WaitGroup
		waitGroup.Add(5)
		subjectWisePerformanceService.SetClassAttendance(mprReq, neoSubject.Subject, &waitGroup)
		subjectWisePerformanceService.SetSessionWiseBreakdown(mprReq, neoSubject.Subject, &waitGroup)
		subjectWisePerformanceService.SetPerformanceTillDate(mprReq, neoSubject.Subject, &waitGroup)
		subjectWisePerformanceService.SetClassQuiz(mprReq, neoSubject.Subject, &waitGroup)
		subjectWisePerformanceService.SetSubjectWiseAssignmentsAndCallouts(mprReq, neoSubject.Subject, &waitGroup)
		waitGroup.Wait()
		subjectWisePerformanceService.SetInClassCallout(mprReq, neoSubject.Subject,
			subjectWiseData.ClassQuiz)
		subjectWisePerformance = append(subjectWisePerformance, *subjectWiseData)
	}
	return subjectWisePerformance
}

func (m *MPRService) getLookingAheadDetails(mprReq *MPRReq) helper.LookingAhead {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("getLookingAheadDetails paniced - %s", r)
		}
	}()
	return mprReq.UserDetailsResponse.LookingAhead
}

func (m *MPRService) getMonthlyTestPerformance(mprReq *MPRReq) []MonthlyTestPerformance {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetMonthlyTestPerformance paniced - %s", r)
		}
	}()
	var monthlyTests []MonthlyTestPerformance
	monthlyTestService := MonthlyTestPerformanceService{TllmsManager: m.TllmsManager}
	totalAssessmentIds := getMonthlyTestAssessments(mprReq.UserDetailsResponse.MonthlyExamClassmates)
	totalAssessmentIds = monthlyTestService.assessmentIdsWithoutSubjectiveAssessment(totalAssessmentIds)
	var attendedAssessments = make(utility.Set)
	attendedAssessments.AddMulti(monthlyTestService.GetAttendedAssessments(totalAssessmentIds, mprReq)...)
	for _, assessmentId := range totalAssessmentIds {
		var monthlyTestPerformance = &MonthlyTestPerformance{Attended: false, AssessmentId: assessmentId}
		if attendedAssessments.Has(assessmentId) {
			monthlyTestPerformance.Attended = true
		}
		monthlyTestService = MonthlyTestPerformanceService{TllmsManager: m.TllmsManager, MTP: monthlyTestPerformance}
		var waitGroup sync.WaitGroup
		waitGroup.Add(10)
		go monthlyTestService.setAssessmentQuestions(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setAssessmentTime(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setAssessmentTimeTaken(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setAssessmentQuestionAttempts(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setChapterWiseAnalysis(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setDifficultyAnalysis(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setSkillAnalysis(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setSubjectWiseScore(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.setPeersPercentageScores(mprReq, assessmentId, &waitGroup)
		go monthlyTestService.SetRanks(mprReq, assessmentId, &waitGroup)
		waitGroup.Wait()
		monthlyTestService.SetMonthlyTestPerformanceCallout(mprReq, assessmentId)
		monthlyTests = append(monthlyTests, *monthlyTestPerformance)
	}
	return monthlyTests
}
