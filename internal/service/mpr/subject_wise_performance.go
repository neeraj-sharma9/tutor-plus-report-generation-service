package mpr

import (
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/contract"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/mpr/callout"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/utility"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

type SubjectWisePerformanceService struct {
	TllmsManager *manager.TllmsManager
	SWP          *SubjectWisePerformance
}

func (sWPS *SubjectWisePerformanceService) SetClassAttendance(mprReq *MPRReq,
	subject string, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetClassAttendance SubjectWise paniced - %s", r)
		}
		wg.Done()
	}()

	var classAttendance helper.ClassAttendance
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		if subject == neoSubject.Subject {
			classAttendance.TotalClasses = neoSubject.TotalClasses
			classAttendance.OnTime = neoSubject.OnTime
			classAttendance.LateDays = neoSubject.LateDays
			classAttendance.Missed = neoSubject.TotalClasses - neoSubject.TotalAttended
			break
		}
	}

	sWPS.SWP.ClassAttendance = classAttendance
}

func (sWPS *SubjectWisePerformanceService) SetSessionWiseBreakdown(mprReq *MPRReq, subject string, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetSessionWiseBreakdown paniced - %s", r)
		}
		wg.Done()
	}()

	var sessionWiseBreakdownList []SessionWiseBreakdown
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		if subject == neoSubject.Subject {
			for _, class := range neoSubject.Classes {
				sessionWiseBreakdown := SessionWiseBreakdown{Session: class.TopicName, Attended: class.Attended}
				sessionWiseBreakdown.PostClass = sWPS.getSessionWisePostClassInfo(&class, mprReq)
				sessionWiseBreakdown.PreClass = sWPS.getSessionWisePreClassInfo(&class, mprReq)
				sessionWiseBreakdown.InClass = sWPS.getSessionWiseInClassInfo(&class, mprReq)
				sessionWiseBreakdownList = append(sessionWiseBreakdownList, sessionWiseBreakdown)
			}
			break
		}
	}
	sWPS.SWP.SessionWiseBreakdown = sessionWiseBreakdownList
}

func (sWPS *SubjectWisePerformanceService) getSessionWisePostClassInfo(class *helper.ClassesModel, mprReq *MPRReq) PostClass {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("getSessionWisePostClassInfo paniced - %s", r)
		}
	}()
	postClass := PostClass{}
	postClassAssignments := class.PostRequisites.Assessments
	assessmentAttempted := sWPS.QuerySessionWisePostClass(postClassAssignments, mprReq.UserId)
	postClass.Attempted = assessmentAttempted
	postClass.Total = len(postClassAssignments)
	return postClass
}

func (sWPS *SubjectWisePerformanceService) getSessionWisePreClassInfo(class *helper.ClassesModel, mprReq *MPRReq) PreClass {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("getSessionWisePreClassInfo paniced - %s", r)
		}
	}()
	var preClass PreClass
	preClassJourney := class.PreRequisites.Journeys
	//preClassVideos := class.PreRequisites.K12Videos
	sessionWisePreClassJourneys := sWPS.QuerySessionWisePreClassJourney(preClassJourney, mprReq.UserId)
	//videos := QuerySessionWisePreClassVideo(preClassVideos, mprReq.UserId)
	if len(sessionWisePreClassJourneys) > 0 {
		if sessionWisePreClassJourneys[0].Attempted.Valid {
			preClass.Attempted = int(sessionWisePreClassJourneys[0].Attempted.Int32)
		}
		if sessionWisePreClassJourneys[0].Completed.Valid {
			preClass.Completed = int(sessionWisePreClassJourneys[0].Completed.Int32)
		}
		//session.PreClass.Total = len(preClassJourney) + len(preClassVideos)
		preClass.Total = len(preClassJourney)
	}
	return preClass
}

func (sWPS *SubjectWisePerformanceService) getSessionWiseInClassInfo(class *helper.ClassesModel, mprReq *MPRReq) InClass {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("getSessionWiseInClassInfo paniced - %s", r)
		}
	}()
	var inClass InClass
	inClass.Total = class.PollQuiz.TotalQuestion
	inClass.Attempted = class.PollQuiz.Correct + class.PollQuiz.Incorrect
	return inClass
}

func (sWPS *SubjectWisePerformanceService) SetSubjectWiseAssignmentsAndCallouts(mprReq *MPRReq, subject string,
	wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetSubjectWiseAssignmentsAndCallouts paniced - %s", r)
		}
		wg.Done()
	}()

	var assignments PostAssignments
	var postClassCallout, cumulativeCallout string
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		if neoSubject.Subject == subject {
			subjectWisePostClassAssessments := make(utility.Set)
			for _, chapter := range neoSubject.Classes {
				subjectWisePostClassAssessments.AddMulti(chapter.PostRequisites.Assessments...)
			}
			assignments.TotalAssignments = len(subjectWisePostClassAssessments.List())
			assignments.PercentageScore, assignments.CompletedAssignments = sWPS.GetPostAssessmentScore(
				subjectWisePostClassAssessments.List(), mprReq)
			assignments.Correct, assignments.Incorrect, assignments.NotAttempted = sWPS.GetPostAssessmentQuestions(
				subjectWisePostClassAssessments.List(), mprReq)
			assignments.ClassAvg = sWPS.GetPostAssessmentClassAvg(subjectWisePostClassAssessments.List(), mprReq)
			assignments.NationalAvg = sWPS.GetPostAssessmentNationalAvg(subjectWisePostClassAssessments.List(), mprReq)

			proficiency := assignments.PercentageScore
			coverage := 0.0
			if assignments.TotalAssignments > 0 {
				completedAssignment := assignments.CompletedAssignments
				coverage = float64(completedAssignment*100) / float64(assignments.TotalAssignments)
			}
			postClassCallout = strings.Replace(callout.GetSubjectWisePostClassCallout(proficiency, float64(coverage)),
				"<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
			postClassCallout = strings.Replace(postClassCallout, "<Subject>", neoSubject.Subject, -1)
			cumulativeCallout = strings.Replace(callout.GetSubjectWiseCumulativeCallout(), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
			cumulativeCallout = strings.Replace(cumulativeCallout, "<Subject>", neoSubject.Subject, -1)
			break
		}
	}
	sWPS.SWP.Assignments = assignments
	sWPS.SWP.CumulativeCallout = cumulativeCallout
	sWPS.SWP.PostClassCallout = postClassCallout
}

func (sWPS *SubjectWisePerformanceService) GetPostAssessmentScore(assessmentIds []int, mprReq *MPRReq) (int, int) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("GetPostAssessmentScore panicked - %s", r)
		}
	}()
	percentageScore := 0
	completedAssignments := 0
	if len(assessmentIds) == 0 {
		logger.Log.Sugar().Info("GetPostAssessmentScore:: No assessment ID")
		return percentageScore, completedAssignments
	}
	subjectWisePostClassInfo := sWPS.QuerySubjectWiseAssignmentScore(assessmentIds, mprReq)
	if len(subjectWisePostClassInfo) > 0 {
		if subjectWisePostClassInfo[0].PercentageScore.Valid {
			percentageScore = int(math.Round(subjectWisePostClassInfo[0].PercentageScore.Float64))
		}
		if subjectWisePostClassInfo[0].Completed.Valid {
			completedAssignments = int(subjectWisePostClassInfo[0].Completed.Int32)
		}
	}
	return percentageScore, completedAssignments
}

func (sWPS *SubjectWisePerformanceService) GetPostAssessmentQuestions(assessmentIds []int, mprReq *MPRReq) (int, int, int) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("QueryPostAssessmentQuestions panicked - %s", r)
		}
	}()
	correct := 0
	incorrect := 0
	notAttempted := 0
	if len(assessmentIds) == 0 {
		logger.Log.Sugar().Info("QueryPostAssessmentQuestions:: No assessment ID")
		return correct, incorrect, notAttempted
	}
	attemptedAssessmentIds, unattemptedAssessmentIds := sWPS.QueryPostClassAssessmentStatus(assessmentIds, mprReq)
	logger.Log.Sugar().Info("GetPostAssessmentQuestions:: ", attemptedAssessmentIds, unattemptedAssessmentIds)
	if len(attemptedAssessmentIds) > 0 {
		subjectWiseQuestions := sWPS.QuerySubjectWiseAssignmentQuestions(attemptedAssessmentIds, mprReq.UserId)
		for _, item := range subjectWiseQuestions {
			if item.Correct.Valid {
				if strings.ToLower(item.Correct.String) == "true" {
					correct += item.Count
				} else if strings.ToLower(item.Correct.String) == "false" {
					incorrect += item.Count
				} else {
					notAttempted += item.Count
				}
			} else {
				notAttempted += item.Count
			}
		}
	}
	if len(unattemptedAssessmentIds) > 0 {
		notAttempted += sWPS.QueryUnattemptedAssessmentQuestion(unattemptedAssessmentIds)
	}
	return correct, incorrect, notAttempted
}

func (sWPS *SubjectWisePerformanceService) GetPostAssessmentClassAvg(assessmentIds []int, mprReq *MPRReq) int {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("GetPostAssessmentClassAvg panicked - %s", r)
		}
	}()
	var userIds []int64
	classAvg := 0
	for _, list := range mprReq.UserDetailsResponse.MonthlyExamClassmates {
		userIds = append(userIds, list...)
		break
	}
	if len(assessmentIds) == 0 || len(userIds) == 0 {
		logger.Log.Sugar().Info("GetPostAssessmentClassAvg:: Assessment list or User list empty")
		return classAvg
	}
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	subjectWisePostClassInfo := sWPS.QuerySubjectWiseAssignmentClassAverage(assessmentIds, userIds, fromDate, toDate)
	if subjectWisePostClassInfo[0].PercentageScore.Valid {
		classAvg = int(math.Round(subjectWisePostClassInfo[0].PercentageScore.Float64))
	}
	return classAvg
}

func (sWPS *SubjectWisePerformanceService) SetPerformanceTillDate(mprReq *MPRReq, subject string, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetPerformanceTillDate paniced - %s", r)
		}
		wg.Done()
	}()
	var performanceTillDate PerformanceTillDate

	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		if subject == neoSubject.Subject {
			performanceTillDate.PostRequisite = sWPS.GetPostClassPerformanceTillDate(neoSubject.PerformanceTillDate.PostRequisite, mprReq)
			performanceTillDate.PreRequisite = sWPS.GetPreClassPerformanceTillDate(neoSubject.PerformanceTillDate.PostRequisite, mprReq)
			performanceTillDate.ClassAttendance = neoSubject.PerformanceTillDate.ClassAttendance
			performanceTillDate.PollQuestion = neoSubject.PerformanceTillDate.PollQuestion
			break
		}
	}
	sWPS.SWP.PerformanceTillDate = performanceTillDate
}

func (sWPS *SubjectWisePerformanceService) SetClassQuiz(mprReq *MPRReq, subject string, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetClassQuiz paniced - %s", r)
		}
		wg.Done()
	}()
	var classQuiz ClassQuiz
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		if neoSubject.Subject == subject {
			var correctAnswer, totalQues, totalAttempted, classAvg, incorrectAnswer, classCount, totalClassAverages, classAvgPercentage, percentage = 0, 0, 0, 0, 0, 0, 0, 0, 0.0
			for _, classes := range neoSubject.Classes {
				correctAnswer += classes.PollQuiz.Correct
				incorrectAnswer += classes.PollQuiz.Incorrect
				totalQues += classes.PollQuiz.TotalQuestion
				totalAttempted += classes.PollQuiz.TotalAttempted
				classAvg += classes.PollQuiz.ClassAverage
				if classes.PollQuiz.TotalQuestion > 0 {
					classCount += 1
					if classes.PollQuiz.TotalQuestion > 0 {
						totalClassAverages += classes.PollQuiz.ClassAverage * 100 / classes.PollQuiz.TotalQuestion
					}

				}
			}
			if classCount > 0 {
				classAvgPercentage = totalClassAverages / classCount
			}

			if totalQues > 0 {
				score := (float64(correctAnswer) / float64(totalQues)) * 100
				percentage = math.Round(score)
			}
			var notAttempted = totalQues - totalAttempted
			classQuiz.Attempted = totalAttempted
			classQuiz.ClassAvg = classAvgPercentage
			classQuiz.PercentageScore = percentage
			classQuiz.TotalClassQuiz = classCount
			classQuiz.Correct = correctAnswer
			classQuiz.TotalQuestions = totalQues
			classQuiz.Incorrect = incorrectAnswer
			classQuiz.NotAttempted = notAttempted
			classQuiz.NationalAvg = neoSubject.PollQuestionNationalPer

			break
		}
	}
	sWPS.SWP.ClassQuiz = classQuiz
}

func (sWPS *SubjectWisePerformanceService) SetInClassCallout(mprReq *MPRReq, subject string, classQuiz ClassQuiz) {
	var inClassCallout string
	for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
		if subject == neoSubject.Subject {
			attended := neoSubject.TotalAttended
			var attendancePer = 0.0
			totalClasses := neoSubject.TotalClasses
			if totalClasses > 0 {
				per := (float64(attended) / float64(totalClasses)) * 100
				attendancePer = math.Round(per)
			}
			attempted := classQuiz.Attempted
			totalQuizClass := classQuiz.TotalQuestions
			var inClassAttemptedPercentage = 0.0
			if totalQuizClass > 0 {
				score := (float64(attempted) / float64(totalQuizClass)) * 100
				inClassAttemptedPercentage = math.Round(score)
				// inClassAttemptedPercentage = (attempted / totalQuizClass) * 100
			}
			inClassCallOut := strings.Replace(callout.GetInClassCallout(attendancePer, inClassAttemptedPercentage),
				"<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
			inClassCallOut = strings.Replace(inClassCallOut, "<Subject>", neoSubject.Subject, -1)
			break
		}
	}
	sWPS.SWP.InClassCallout = inClassCallout
}

func (sWPS *SubjectWisePerformanceService) QuerySessionWisePostClass(assessmentIds []int, userId int64) int {
	return int(sWPS.TllmsManager.SessionWiseAssignmentsCount(assessmentIds, userId))
}

func (sWPS *SubjectWisePerformanceService) QuerySessionWisePreClassJourney(journeys []int, userId int64) []contract.AssignmentsCompletionCursor {
	var sessionWisePreClassJourney []contract.AssignmentsCompletionCursor
	if len(journeys) == 0 {
		return sessionWisePreClassJourney
	}
	sessionWisePreClassJourney = sWPS.TllmsManager.SessionWiseLearnJourneyCount(journeys, userId)
	return sessionWisePreClassJourney
}

func (sWPS *SubjectWisePerformanceService) QuerySessionWisePreClassVideo(videos []int, userId int64) int {
	//if len(videos) == 0 {
	//	return 0
	//}
	//query, args, err := sqlx.In(SessionWiseVideoCount,
	//	videos,
	//	userId)
	//db.CheckError(err, "sqlx-QuerySessionWisePreClassVideo")
	//
	//var sessionWisePreClassVideo []AssignmentAttemptedCursor
	//query = db.GetPGDbReader().Rebind(query)
	//err = db.GetPGDbReader().
	//	Select(&sessionWisePreClassVideo, query, args...)
	//
	//db.CheckError(err, "QuerySessionWisePreClassVideo")
	//
	//return sessionWisePreClassVideo[0].Count
	return 0
}

func (sWPS *SubjectWisePerformanceService) QuerySubjectWiseAssignmentScore(assessmentIds []int,
	mprReq *MPRReq) []contract.AssignmentScoreCursor {
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	return sWPS.TllmsManager.SubjectWiseAssignmentsScore(assessmentIds, []int64{mprReq.UserId}, fromDate, toDate)
}

func (sWPS *SubjectWisePerformanceService) QuerySubjectWiseAssignmentQuestions(assessmentIds []int,
	userId int64) []contract.AssignmentQuestionsCursor {
	return sWPS.TllmsManager.SubjectWiseAssignmentsQuestions(assessmentIds, userId)
}

func (sWPS *SubjectWisePerformanceService) QuerySubjectWiseAssignmentClassAverage(assessmentIds []int, userIds []int64,
	fromDate string, toDate string) []contract.AssignmentScoreCursor {
	return sWPS.TllmsManager.SubjectWiseAssignmentsAverage(assessmentIds, userIds, fromDate, toDate)
}

func (sWPS *SubjectWisePerformanceService) GetPostAssessmentNationalAvg(assessmentIds []int, mprReq *MPRReq) int {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("GetPostAssessmentNationalAvg panicked - %s", r)
		}
	}()
	nationalAvg := 0
	var userIds []int64
	userIds = append(userIds, mprReq.UserDetailsResponse.SameBatchUserList...)
	if len(assessmentIds) == 0 || len(userIds) == 0 {
		logger.Log.Info("GetPostAssessmentNationalAvg:: Assessment list or User list empty")
		return nationalAvg
	}
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	subjectWisePostClassInfo := sWPS.QuerySubjectWiseAssignmentNationalAverage(assessmentIds, userIds, fromDate, toDate)
	if subjectWisePostClassInfo[0].PercentageScore.Valid {
		nationalAvg = int(math.Round(subjectWisePostClassInfo[0].PercentageScore.Float64))
	}
	return nationalAvg
}

func (sWPS *SubjectWisePerformanceService) QuerySubjectWiseAssignmentNationalAverage(assessmentIds []int, userIds []int64,
	fromDate string, toDate string) []contract.AssignmentScoreCursor {
	return sWPS.TllmsManager.SubjectWiseAssignmentsAverage(assessmentIds, userIds, fromDate, toDate)
}

func (sWPS *SubjectWisePerformanceService) GetPostClassPerformanceTillDate(assessmentIds []int, mprReq *MPRReq) RequisiteStatus {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Fatal("GetPostClassPerformanceTillDate panicked - %v", r)
		}
	}()
	postClassPerformanceTillDate := RequisiteStatus{Total: len(assessmentIds), Completed: 0, AttemptedPercentage: 0, Missed: len(assessmentIds), InProgress: 0}
	if len(assessmentIds) == 0 {
		logger.Log.Sugar().Info("GetPostClassPerformanceTillDate:: No assessment ID")
	} else {
		postClassTillDateDB := sWPS.QueryPostClassPerformanceTillDate(assessmentIds, mprReq.UserId)
		for _, dbEntry := range postClassTillDateDB {
			if dbEntry.Status.Valid {
				if strings.ToUpper(dbEntry.Status.String) == "GRADED" {
					postClassPerformanceTillDate.Completed += dbEntry.Count
				} else if strings.ToUpper(dbEntry.Status.String) == "STARTED" {
					postClassPerformanceTillDate.InProgress += dbEntry.Count
				}
			}
		}
		postClassPerformanceTillDate.Missed = postClassPerformanceTillDate.Total - (postClassPerformanceTillDate.Completed + postClassPerformanceTillDate.InProgress)
		postClassPerformanceTillDate.AttemptedPercentage = ((postClassPerformanceTillDate.Completed + postClassPerformanceTillDate.InProgress) * 100) / postClassPerformanceTillDate.Total
	}
	return postClassPerformanceTillDate
}

func (sWPS *SubjectWisePerformanceService) QueryPostClassPerformanceTillDate(assessmentIds []int, userId int64) []contract.PostAssignmentTillDate {
	return sWPS.TllmsManager.PostClassPerformanceTillDate(assessmentIds, userId)
}

func (sWPS *SubjectWisePerformanceService) GetPreClassPerformanceTillDate(journeyIds []int, mprReq *MPRReq) RequisiteStatus {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Fatal("GetPreClassPerformanceTillDate panicked - %v", r)
		}
	}()
	preClassPerformanceTillDate := RequisiteStatus{Total: len(journeyIds), Completed: 0, AttemptedPercentage: 0, Missed: len(journeyIds), InProgress: 0}
	if len(journeyIds) == 0 {
		logger.Log.Sugar().Info("GetPreClassPerformanceTillDate:: No journey ID")
	} else {
		preClassTillDateDB := sWPS.QueryPreClassPerformanceTillDate(journeyIds, mprReq.UserId)
		for _, dbEntry := range preClassTillDateDB {
			if dbEntry.Completed.Valid {
				if strings.ToUpper(dbEntry.Completed.String) == "FALSE" {
					preClassPerformanceTillDate.InProgress += dbEntry.Count
				} else if strings.ToUpper(dbEntry.Completed.String) == "TRUE" {
					preClassPerformanceTillDate.Completed += dbEntry.Count
				}
			}
		}
		preClassPerformanceTillDate.Missed = preClassPerformanceTillDate.Total - (preClassPerformanceTillDate.Completed + preClassPerformanceTillDate.InProgress)
		preClassPerformanceTillDate.AttemptedPercentage = ((preClassPerformanceTillDate.Completed + preClassPerformanceTillDate.InProgress) * 100) / preClassPerformanceTillDate.Total
	}
	return preClassPerformanceTillDate
}

func (sWPS *SubjectWisePerformanceService) QueryPreClassPerformanceTillDate(journeyIds []int, userId int64) []contract.PreAssignmentTillDate {
	return sWPS.TllmsManager.PreClassPerformanceTillDate(journeyIds, userId)
}

func (sWPS *SubjectWisePerformanceService) QueryPostClassAssessmentStatus(assessmentIds []int, mprReq *MPRReq) ([]int, []int) {
	var attemptedAssessments = make(utility.Set)
	var totalAssessments = make(utility.Set)
	totalAssessments.AddMulti(assessmentIds...)
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	assessmentStatus := sWPS.TllmsManager.PostAssessmentAttempted(mprReq.UserId, assessmentIds, fromDate, toDate)
	for _, assessment := range assessmentStatus {
		if assessment.AssessmentID.Valid {
			attemptedAssessments.AddOrUpdate(int(assessment.AssessmentID.Int32))
		}
	}
	unattemptedAssessments := totalAssessments.Difference(attemptedAssessments).List()
	return attemptedAssessments.List(), unattemptedAssessments
}

func (sWPS *SubjectWisePerformanceService) QueryUnattemptedAssessmentQuestion(assessmentIds []int) int {
	return int(sWPS.TllmsManager.UnattemptedAssessmentQuestions(assessmentIds))
}
