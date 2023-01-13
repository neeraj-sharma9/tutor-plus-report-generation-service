package mpr

import (
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/contract"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/mpr/callout"
	"golang.org/x/tools/godoc/util"
	"math"
	"strings"
	"sync"
	"time"
)

func (report *ProgressData) SetClassAttendance(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetClassAttendance SubjectWise paniced - %s", r)
		}
		wg.Done()
	}()

	allSubjectWisePerformance := report.SubjectWisePerformance
	for subjectOfPerformance := range report.SubjectWisePerformance {
		for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
			if allSubjectWisePerformance[subjectOfPerformance].Subject == neoSubject.Subject {
				allSubjectWisePerformance[subjectOfPerformance].ClassAttendance.TotalClasses = neoSubject.TotalClasses
				allSubjectWisePerformance[subjectOfPerformance].ClassAttendance.OnTime = neoSubject.OnTime
				allSubjectWisePerformance[subjectOfPerformance].ClassAttendance.LateDays = neoSubject.LateDays
				allSubjectWisePerformance[subjectOfPerformance].ClassAttendance.Missed = neoSubject.TotalClasses - neoSubject.TotalAttended
				break
			}
		}
	}
}

func (report *ProgressData) SetSessionWiseBreakdown(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetSessionWiseBreakdown paniced - %s", r)
		}
		wg.Done()
	}()

	allSubjectWisePerformance := report.SubjectWisePerformance
	for index := range report.SubjectWisePerformance {
		for _, subject := range mprReq.UserDetailsResponse.Subjects {
			if allSubjectWisePerformance[index].Subject == subject.Subject {
				var sessionWiseBreakdownArr []SessionWiseBreakdown
				for _, class := range subject.Classes {
					var waitGroup sync.WaitGroup
					sessionWiseBreakdown := &SessionWiseBreakdown{Session: class.TopicName, Attended: class.Attended,
						TllmsManager: report.TllmsManager}
					waitGroup.Add(3)
					go sessionWiseBreakdown.setSessionWisePostClassInfo(&class, mprReq, &waitGroup)
					go sessionWiseBreakdown.setSessionWisePreClassInfo(&class, mprReq, &waitGroup)
					go sessionWiseBreakdown.setSessionWiseInClassInfo(&class, mprReq, &waitGroup)
					waitGroup.Wait()
					sessionWiseBreakdownArr = append(sessionWiseBreakdownArr, *sessionWiseBreakdown)
				}
				allSubjectWisePerformance[index].SessionWiseBreakdown = sessionWiseBreakdownArr
			}
		}
	}
}

func (session *SessionWiseBreakdown) setSessionWisePostClassInfo(class *helper.ClassesModel, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSessionWisePostClassInfo paniced - %s", r)
		}
		wg.Done()
	}()
	postClassAssignments := class.PostRequisites.Assessments
	assessmentAttempted := session.QuerySessionWisePostClass(postClassAssignments, mprReq.UserId)
	session.PostClass.Attempted = assessmentAttempted
	session.PostClass.Total = len(postClassAssignments)
}

func (session *SessionWiseBreakdown) setSessionWisePreClassInfo(class *helper.ClassesModel, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSessionWisePreClassInfo paniced - %s", r)
		}
		wg.Done()
	}()
	preClassJourney := class.PreRequisites.Journeys
	//preClassVideos := class.PreRequisites.K12Videos
	sessionWisePreClassJourneys := session.QuerySessionWisePreClassJourney(preClassJourney, mprReq.UserId)
	//videos := QuerySessionWisePreClassVideo(preClassVideos, mprReq.UserId)
	if len(sessionWisePreClassJourneys) > 0 {
		if sessionWisePreClassJourneys[0].Attempted.Valid {
			session.PreClass.Attempted = int(sessionWisePreClassJourneys[0].Attempted.Int32)
		}
		if sessionWisePreClassJourneys[0].Completed.Valid {
			session.PreClass.Completed = int(sessionWisePreClassJourneys[0].Completed.Int32)
		}
		//session.PreClass.Total = len(preClassJourney) + len(preClassVideos)
		session.PreClass.Total = len(preClassJourney)
	}
}

func (session *SessionWiseBreakdown) setSessionWiseInClassInfo(class *helper.ClassesModel, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSessionWiseInClassInfo paniced - %s", r)
		}
		wg.Done()
	}()
	session.InClass.Total = class.PollQuiz.TotalQuestion
	session.InClass.Attempted = class.PollQuiz.Correct + class.PollQuiz.Incorrect
}

func (report *ProgressData) SetSubjectWiseAssignments(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetSubjectAssignments paniced - %s", r)
		}
		wg.Done()
	}()

	allSubjectWisePerformance := report.SubjectWisePerformance
	for index := range report.SubjectWisePerformance {
		for _, subject := range mprReq.UserDetailsResponse.Subjects {
			if allSubjectWisePerformance[index].Subject == subject.Subject {
				subjectWisePostClassAssessments := make(util.Set)
				for _, chapter := range subject.Classes {
					subjectWisePostClassAssessments.AddMulti(chapter.PostRequisites.Assessments...)
				}
				assignments := &PostAssignments{TotalAssignments: len(subjectWisePostClassAssessments.List())}
				var waitGroup sync.WaitGroup
				waitGroup.Add(4)
				go assignments.SetPostAssessmentScore(subjectWisePostClassAssessments.List(), mprReq, &waitGroup)
				go assignments.SetPostAssessmentQuestions(subjectWisePostClassAssessments.List(), mprReq, &waitGroup)
				go assignments.SetPostAssessmentClassAvg(subjectWisePostClassAssessments.List(), mprReq, &waitGroup)
				go assignments.SetPostAssessmentNationalAvg(subjectWisePostClassAssessments.List(), mprReq, &waitGroup)
				waitGroup.Wait()
				allSubjectWisePerformance[index].Assignments = *assignments
				proficiency := allSubjectWisePerformance[index].Assignments.PercentageScore
				coverage := 0.0
				if allSubjectWisePerformance[index].Assignments.TotalAssignments > 0 {
					completedAssignment := allSubjectWisePerformance[index].Assignments.CompletedAssignments
					coverage = float64(completedAssignment*100) / float64(allSubjectWisePerformance[index].Assignments.TotalAssignments)
				}

				postClassCallout := strings.Replace(callout.GetSubjectWisePostClassCallout(proficiency, float64(coverage)), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
				postClassCallout = strings.Replace(postClassCallout, "<Subject>", subject.Subject, -1)
				allSubjectWisePerformance[index].PostClassCallout = postClassCallout
				cumulativeCallout := strings.Replace(callout.GetSubjectWiseCumulativeCallout(), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
				cumulativeCallout = strings.Replace(cumulativeCallout, "<Subject>", subject.Subject, -1)
				allSubjectWisePerformance[index].CumulativeCallout = cumulativeCallout
			}
		}
	}
}

func (subjectWiseAssignment *PostAssignments) SetPostAssessmentScore(assessmentIds []int, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetPostAssessmentScore ----%+v", subjectWiseAssignment))
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("SetPostAssessmentScore panicked - %s", r)
		}
		wg.Done()
	}()
	subjectWiseAssignment.PercentageScore = 0
	subjectWiseAssignment.CompletedAssignments = 0
	if len(assessmentIds) == 0 {
		log.Info("SetPostAssessmentScore:: No assessment ID")
		return
	}
	subjectWisePostClassInfo := QuerySubjectWiseAssignmentScore(assessmentIds, mprReq)
	if len(subjectWisePostClassInfo) > 0 {
		if subjectWisePostClassInfo[0].PercentageScore.Valid {
			subjectWiseAssignment.PercentageScore = int(math.Round(subjectWisePostClassInfo[0].PercentageScore.Float64))
		}
		if subjectWisePostClassInfo[0].Completed.Valid {
			subjectWiseAssignment.CompletedAssignments = int(subjectWisePostClassInfo[0].Completed.Int32)
		}
	}
	log.Debugf(fmt.Sprintf("---Bye SetPostAssessmentScore ----%+v", subjectWiseAssignment))
}

func (subjectWiseAssignment *PostAssignments) SetPostAssessmentQuestions(assessmentIds []int, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetPostAssessmentQuestions ----%+v", subjectWiseAssignment))
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("QueryPostAssessmentQuestions panicked - %s", r)
		}
		wg.Done()
	}()
	subjectWiseAssignment.Correct = 0
	subjectWiseAssignment.Incorrect = 0
	subjectWiseAssignment.NotAttempted = 0
	if len(assessmentIds) == 0 {
		log.Info("QueryPostAssessmentQuestions:: No assessment ID")
		return
	}
	attemptedAssessmentIds, unattemptedAssessmentIds := QueryPostClassAssessmentStatus(assessmentIds, mprReq)
	log.Infof("SetPostAssessmentQuestions:: ", attemptedAssessmentIds, unattemptedAssessmentIds)
	if len(attemptedAssessmentIds) > 0 {
		subjectWiseQuestions := QuerySubjectWiseAssignmentQuestions(attemptedAssessmentIds, mprReq.UserId)
		for _, item := range subjectWiseQuestions {
			if item.Correct.Valid {
				if strings.ToLower(item.Correct.String) == "true" {
					subjectWiseAssignment.Correct += item.Count
				} else if strings.ToLower(item.Correct.String) == "false" {
					subjectWiseAssignment.Incorrect += item.Count
				} else {
					subjectWiseAssignment.NotAttempted += item.Count
				}
			} else {
				subjectWiseAssignment.NotAttempted += item.Count
			}
		}
	}
	if len(unattemptedAssessmentIds) > 0 {
		subjectWiseAssignment.NotAttempted += QueryUnattemptedAssessmentQuestion(unattemptedAssessmentIds)
	}
	log.Debugf(fmt.Sprintf("---Bye SetPostAssessmentQuestions ----%+v", subjectWiseAssignment))
}

func (subjectWiseAssignment *PostAssignments) SetPostAssessmentClassAvg(assessmentIds []int, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetPostAssessmentClassAvg ----%+v, mprReq: %+v", subjectWiseAssignment, mprReq))
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("SetPostAssessmentClassAvg panicked - %s", r)
		}
		wg.Done()
	}()
	subjectWiseAssignment.PercentageScore = 0
	var userIds []int64
	for _, list := range mprReq.UserDetailsResponse.MonthlyExamClassmates {
		userIds = append(userIds, list...)
		break
	}
	if len(assessmentIds) == 0 || len(userIds) == 0 {
		log.Info("SetPostAssessmentClassAvg:: Assessment list or User list empty")
		return
	}
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	subjectWisePostClassInfo := QuerySubjectWiseAssignmentClassAverage(assessmentIds, userIds, fromDate, toDate)
	if subjectWisePostClassInfo[0].PercentageScore.Valid {
		subjectWiseAssignment.ClassAvg = int(math.Round(subjectWisePostClassInfo[0].PercentageScore.Float64))
	}
	log.Debugf(fmt.Sprintf("---Bye SetPostAssessmentClassAvg ----%+v", subjectWiseAssignment))
}

func (report *ProgressData) SetPerformanceTillDate(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetPerformanceTillDate paniced - %s", r)
		}
		wg.Done()
	}()
	allSubjectWisePerformance := report.SubjectWisePerformance
	for subjectOfPerformance := range report.SubjectWisePerformance {
		for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
			if allSubjectWisePerformance[subjectOfPerformance].Subject == neoSubject.Subject {
				var waitGroup sync.WaitGroup
				waitGroup.Add(2)
				allSubjectWisePerformance[subjectOfPerformance].PerformanceTillDate.SetPostClassPerformanceTillDate(neoSubject.PerformanceTillDate.PostRequisite, mprReq, &waitGroup)
				allSubjectWisePerformance[subjectOfPerformance].PerformanceTillDate.SetPreClassPerformanceTillDate(neoSubject.PerformanceTillDate.PreRequisite, mprReq, &waitGroup)
				waitGroup.Wait()
				allSubjectWisePerformance[subjectOfPerformance].PerformanceTillDate.ClassAttendance = neoSubject.PerformanceTillDate.ClassAttendance
				allSubjectWisePerformance[subjectOfPerformance].PerformanceTillDate.PollQuestion = neoSubject.PerformanceTillDate.PollQuestion
				break
			}
		}
	}
}

func (report *ProgressData) SetClassQuiz(mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetClassQuiz ----%+v", report))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetClassQuiz paniced - %s", r)
		}
		wg.Done()
	}()
	allSubjectWisePerformance := report.SubjectWisePerformance
	for subjectOfPerformance := range report.SubjectWisePerformance {
		for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
			if allSubjectWisePerformance[subjectOfPerformance].Subject == neoSubject.Subject {
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
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.Attempted = totalAttempted
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.ClassAvg = classAvgPercentage
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.PercentageScore = percentage
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.TotalClassQuiz = classCount
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.Correct = correctAnswer
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.TotalQuestions = totalQues
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.Incorrect = incorrectAnswer
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.NotAttempted = notAttempted
				allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.NationalAvg = neoSubject.PollQuestionNationalPer

				break
			}
		}
	}
	log.Debugf(fmt.Sprintf("---Bye SetClassQuiz ----%+v", report))
}

func (report *ProgressData) SetInClassCallout(mprReq *helper.MPRReq) {
	log.Debugf(fmt.Sprintf("---Hello SetInClassCallout ----%+v", report))
	allSubjectWisePerformance := report.SubjectWisePerformance
	for subjectOfPerformance := range report.SubjectWisePerformance {
		for _, neoSubject := range mprReq.UserDetailsResponse.Subjects {
			if allSubjectWisePerformance[subjectOfPerformance].Subject == neoSubject.Subject {
				attended := neoSubject.TotalAttended
				var attendancePer = 0.0
				totalClasses := neoSubject.TotalClasses
				if totalClasses > 0 {
					per := (float64(attended) / float64(totalClasses)) * 100
					attendancePer = math.Round(per)
				}
				attempted := allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.Attempted
				totalQuizClass := allSubjectWisePerformance[subjectOfPerformance].ClassQuiz.TotalQuestions
				var inClassAttemptedPercentage = 0.0
				if totalQuizClass > 0 {
					score := (float64(attempted) / float64(totalQuizClass)) * 100
					inClassAttemptedPercentage = math.Round(score)
					// inClassAttemptedPercentage = (attempted / totalQuizClass) * 100
				}
				inClassCallOut := strings.Replace(callout.GetInClassCallout(attendancePer, inClassAttemptedPercentage), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
				inClassCallOut = strings.Replace(inClassCallOut, "<Subject>", neoSubject.Subject, -1)
				allSubjectWisePerformance[subjectOfPerformance].InClassCallout = inClassCallOut

			}
		}
	}
	log.Debugf(fmt.Sprintf("---Bye SetInClassCallout ----%+v", report))
}

func (session *SessionWiseBreakdown) QuerySessionWisePostClass(assessmentIds []int, userId int64) int {
	return int(session.TllmsManager.SessionWiseAssignmentsCount(assessmentIds, userId))
}

func (session *SessionWiseBreakdown) QuerySessionWisePreClassJourney(journeys []int, userId int64) []contract.AssignmentsCompletionCursor {
	var sessionWisePreClassJourney []contract.AssignmentsCompletionCursor
	if len(journeys) == 0 {
		return sessionWisePreClassJourney
	}
	sessionWisePreClassJourney = session.TllmsManager.SessionWiseLearnJourneyCount(journeys, userId)
	return sessionWisePreClassJourney
}

func QuerySessionWisePreClassVideo(videos []int, userId int64) int {
	if len(videos) == 0 {
		return 0
	}
	query, args, err := sqlx.In(SessionWiseVideoCount,
		videos,
		userId)
	db.CheckError(err, "sqlx-QuerySessionWisePreClassVideo")

	var sessionWisePreClassVideo []AssignmentAttemptedCursor
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&sessionWisePreClassVideo, query, args...)

	db.CheckError(err, "QuerySessionWisePreClassVideo")

	return sessionWisePreClassVideo[0].Count
}

func QuerySubjectWiseAssignmentScore(assessmentIds []int, mprReq *helper.MPRReq) []AssignmentScoreCursor {
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	query, args, err := sqlx.In(SubjectWiseAssignmentsScore, assessmentIds, []int64{mprReq.UserId}, fromDate, toDate)
	db.CheckError(err, "sqlx-QueryPostAssessmentScore")
	var subjectWisePostClassInfo []AssignmentScoreCursor
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&subjectWisePostClassInfo, query, args...)
	db.CheckError(err, "QueryPostAssessmentScore", query)
	return subjectWisePostClassInfo
}

func QuerySubjectWiseAssignmentQuestions(assessmentIds []int, userId int64) []AssignmentQuestionsCursor {
	query, args, err := sqlx.In(SubjectWiseAssignmentsQuestions,
		assessmentIds,
		userId)
	db.CheckError(err, "sqlx-QueryPostAssessmentQuestions")
	var subjectWiseQuestions []AssignmentQuestionsCursor
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&subjectWiseQuestions, query, args...)
	db.CheckError(err, "QueryPostAssessmentQuestions", query)
	return subjectWiseQuestions
}

func QuerySubjectWiseAssignmentClassAverage(assessmentIds []int, userIds []int64, fromDate string, toDate string) []AssignmentScoreCursor {
	var subjectWisePostClassInfo []AssignmentScoreCursor
	query, args, err := sqlx.In(SubjectWiseAssignmentsAverage,
		assessmentIds,
		userIds, fromDate, toDate)
	db.CheckError(err, "sqlx-QuerySubjectWiseAssignmentClassAverage")
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&subjectWisePostClassInfo, query, args...)
	db.CheckError(err, "QuerySubjectWiseAssignmentClassAverage", query)
	return subjectWisePostClassInfo
}

func (subjectWiseAssignment *PostAssignments) SetPostAssessmentNationalAvg(assessmentIds []int, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetPostAssessmentNationalAvg ----%+v, mprReq: %+v", subjectWiseAssignment, mprReq))
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("SetPostAssessmentNationalAvg panicked - %s", r)
		}
		wg.Done()
	}()
	subjectWiseAssignment.PercentageScore = 0
	var userIds []int64
	userIds = append(userIds, mprReq.UserDetailsResponse.SameBatchUserList...)
	if len(assessmentIds) == 0 || len(userIds) == 0 {
		log.Info("SetPostAssessmentNationalAvg:: Assessment list or User list empty")
		return
	}
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.DATE_LAYOUT)
	subjectWisePostClassInfo := QuerySubjectWiseAssignmentNationalAverage(assessmentIds, userIds, fromDate, toDate)
	if subjectWisePostClassInfo[0].PercentageScore.Valid {
		subjectWiseAssignment.NationalAvg = int(math.Round(subjectWisePostClassInfo[0].PercentageScore.Float64))
	}
	log.Debugf(fmt.Sprintf("---Bye SetPostAssessmentNationalAvg ----%+v", subjectWiseAssignment))
}

func QuerySubjectWiseAssignmentNationalAverage(assessmentIds []int, userIds []int64, fromDate string, toDate string) []AssignmentScoreCursor {
	var subjectWisePostClassInfo []AssignmentScoreCursor
	query, args, err := sqlx.In(SubjectWiseAssignmentsAverage,
		assessmentIds,
		userIds, fromDate, toDate)
	db.CheckError(err, "sqlx-QuerySubjectWiseAssignmentNationalAverage")
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&subjectWisePostClassInfo, query, args...)
	db.CheckError(err, "QuerySubjectWiseAssignmentNationalAverage", query)
	return subjectWisePostClassInfo
}

func (performanceTillDate *PerformanceTillDate) SetPostClassPerformanceTillDate(assessmentIds []int, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Fatal("SetPostClassPerformanceTillDate panicked - %v", r)
		}
		wg.Done()
	}()
	postClassPerformanceTillDate := RequisiteStatus{Total: len(assessmentIds), Completed: 0, AttemptedPercentage: 0, Missed: len(assessmentIds), InProgress: 0}
	if len(assessmentIds) == 0 {
		logger.Log.Sugar().Info("SetPostClassPerformanceTillDate:: No assessment ID")
	} else {
		postClassTillDateDB := QueryPostClassPerformanceTillDate(assessmentIds, mprReq.UserId)
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
	performanceTillDate.PostRequisite = postClassPerformanceTillDate
	log.Debugf(fmt.Sprintf("---Bye SetPostClassPerformanceTillDate ----%+v", performanceTillDate))
}

func QueryPostClassPerformanceTillDate(assessmentIds []int, userId int64) []PostAssignmentTillDate {
	query, args, err := sqlx.In(PostClassPerformanceTillDate,
		assessmentIds,
		userId)
	db.CheckError(err, "sqlx-QueryPostClassPerformanceTillDate")
	var PostClassInfo []PostAssignmentTillDate
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&PostClassInfo, query, args...)
	db.CheckError(err, "QueryPostClassPerformanceTillDate", query)
	return PostClassInfo
}

func (performanceTillDate *PerformanceTillDate) SetPreClassPerformanceTillDate(journeyIds []int, mprReq *helper.MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetPreClassPerformanceTillDate ----%+v", performanceTillDate))
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("SetPreClassPerformanceTillDate panicked - %s", r)
		}
		wg.Done()
	}()
	preClassPerformanceTillDate := RequisiteStatus{Total: len(journeyIds), Completed: 0, AttemptedPercentage: 0, Missed: len(journeyIds), InProgress: 0}
	if len(journeyIds) == 0 {
		log.Info("SetPreClassPerformanceTillDate:: No journey ID")
	} else {
		preClassTillDateDB := QueryPreClassPerformanceTillDate(journeyIds, mprReq.UserId)
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
	performanceTillDate.PreRequisite = preClassPerformanceTillDate
	log.Debugf(fmt.Sprintf("---Bye SetPreClassPerformanceTillDate ----%+v", performanceTillDate))
}

func QueryPreClassPerformanceTillDate(journeyIds []int, userId int64) []PreAssignmentTillDate {
	query, args, err := sqlx.In(PreClassPerformanceTillDate,
		userId,
		journeyIds)
	db.CheckError(err, "sqlx-QueryPreClassPerformanceTillDate")
	var PreClassInfo []PreAssignmentTillDate
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&PreClassInfo, query, args...)
	db.CheckError(err, "QueryPreClassPerformanceTillDate", query)
	return PreClassInfo
}

func QueryPostClassAssessmentStatus(assessmentIds []int, mprReq *helper.MPRReq) ([]int, []int) {
	var attemptedAssessments = make(util.Set)
	var totalAssessments = make(util.Set)
	totalAssessments.AddMulti(assessmentIds...)
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	query, args, err := sqlx.In(PostAssessmentAttempted, assessmentIds, mprReq.UserId, fromDate, toDate)
	db.CheckError(err, "sqlx-QueryPostClassAssessmentStatus")
	var assessmentStatus []AttemptedPostAssessment
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&assessmentStatus, query, args...)
	db.CheckError(err, "QueryPostClassAssessmentStatus", query)
	for _, assessment := range assessmentStatus {
		if assessment.AssessmentID.Valid {
			attemptedAssessments.AddOrUpdate(int(assessment.AssessmentID.Int32))
		}
	}
	unattemptedAssessments := totalAssessments.Difference(attemptedAssessments).List()
	return attemptedAssessments.List(), unattemptedAssessments
}

func QueryUnattemptedAssessmentQuestion(assessmentIds []int) int {
	if len(assessmentIds) == 0 {
		return 0
	}
	query, args, err := sqlx.In(UnattemptedAssessmentQuestions,
		assessmentIds)
	db.CheckError(err, "sqlx-QueryUnattemptedAssessmentQuestion")

	var unattemptedQuestions []AssignmentAttemptedCursor
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&unattemptedQuestions, query, args...)
	db.CheckError(err, "QueryUnattemptedAssessmentQuestion", query)

	return unattemptedQuestions[0].Count
}
