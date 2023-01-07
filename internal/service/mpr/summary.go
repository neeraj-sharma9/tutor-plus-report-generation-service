package mpr

import (
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/service/mpr/callout"
	"math"
	"strings"
)

type SummaryService struct {
	TllmsManager *manager.TllmsManager
}

func (sS *SummaryService) GetChaptersCovered(mprReq *MPRReq) ChaptersCovered {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("GetChaptersCovered paniced - %s", r)
		}
	}()
	chapterCovered := ChaptersCovered{TotalChapters: 0, Completed: 0, Missed: 0}
	for _, subject := range mprReq.UserDetailsResponse.Subjects {
		for _, chapter := range subject.ChapterCovered {
			if chapter.ClassesAttended.TotalClasses > 0 {
				chapterCovered.TotalChapters += 1
				if chapter.ClassesAttended.TotalAttended > 0 {
					chapterCovered.Completed += 1
				} else {
					chapterCovered.Missed += 1
				}
			}
		}
	}
	chapterCovered.Callout = strings.Replace(callout.GetChapterCoveredCallout(chapterCovered.Completed*100/chapterCovered.TotalChapters), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
	return chapterCovered
}

func (sS *SummaryService) GetClassAttendance(mprReq *MPRReq) ClassAttendanceForSummary {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("GetClassAttendance paniced - %s", r)
		}
	}()
	classAttendance := ClassAttendanceForSummary{OnTime: 0, TotalClasses: 0, Missed: 0, LateDays: 0}
	for _, subject := range mprReq.UserDetailsResponse.Subjects {
		classAttendance.TotalClasses += subject.TotalClasses
		classAttendance.LateDays += subject.LateDays
		classAttendance.OnTime += subject.OnTime
		classAttendance.Missed += subject.TotalClasses - subject.TotalAttended
	}
	if classAttendance.TotalClasses > 0 {
		classAttendant := classAttendance.TotalClasses - classAttendance.Missed
		percentage := float64(classAttendant*100) / float64(classAttendance.TotalClasses)
		classAttendance.Callout = strings.Replace(callout.GetClassAttendanceCallout(percentage), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		return classAttendance
	}
	return classAttendance
}

func (sS *SummaryService) GetSummaryPageAssignments(mprReq *MPRReq,
	subjectWisePerformance []SubjectWisePerformance) Assignments {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("GetSummaryPageAssignments paniced - %s", r)
		}
	}()
	assignments := Assignments{
		TotalAssignments: 0,
		Completed:        0,
		Missed:           0,
		Score:            0,
		Callout:          "",
	}
	for _, subject := range subjectWisePerformance {
		assignments.TotalAssignments += subject.Assignments.TotalAssignments
		assignments.Completed += subject.Assignments.CompletedAssignments
		assignments.Score += float64(subject.Assignments.PercentageScore)
	}
	coverage := 0.0
	assignments.Missed = assignments.TotalAssignments - assignments.Completed
	if assignments.TotalAssignments > 0 {
		assignments.Score = assignments.Score / float64(len(subjectWisePerformance))
		assignments.Score = math.Round(assignments.Score*100) / 100
		assignments.CompletedPerc = float64(assignments.Completed) * 100 / float64(assignments.TotalAssignments)
		assignments.CompletedPerc = math.Round(assignments.CompletedPerc*100) / 100
		coverage = assignments.CompletedPerc
	}
	postClassCallout := strings.Replace(callout.GetSummaryPostClassCallout(assignments.Score, coverage), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
	assignments.Callout = postClassCallout
	return assignments
}

func (sS *SummaryService) GetSummaryLearnerTags(mprReq *MPRReq, report ProgressData) LearnerTags {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("GetSummaryLearnerTags paniced - %s", r)
		}
	}()
	learnerTags := LearnerTags{Regularity: "needs_to_improve",
		Punctuality:          "needs_to_improve",
		HomeWorkCompletion:   "needs_to_improve",
		ActiveInClass:        "needs_to_improve",
		SelfDirectedLearning: "needs_to_improve"}
	classes := mprReq.UserDetailsResponse.OverAllPerformance.ByjusClassInfo.TotalClasses
	if classes > 0 {
		attended := mprReq.UserDetailsResponse.OverAllPerformance.ByjusClassInfo.TotalAttended
		regularity := float64(attended*100) / float64(classes)
		learnerTags.Regularity = GetPerformanceByValue(regularity)
		onTime := mprReq.UserDetailsResponse.OverAllPerformance.ByjusClassInfo.TotalOnTime
		punctuality := float64(onTime*100) / float64(classes)
		learnerTags.Punctuality = GetPerformanceByValue(punctuality)
	}
	homework := report.SummaryOfLearning.Assignments.CompletedPerc
	learnerTags.HomeWorkCompletion = GetPerformanceByValue(homework)
	subjectWisePerformance := report.SubjectWisePerformance
	var totalClassQuiz, totalAttemptedQuiz, preClassCompleted, preClassTotal = 0, 0, 0, 0
	for _, subject := range subjectWisePerformance {
		totalClassQuiz += subject.ClassQuiz.TotalClassQuiz
		totalAttemptedQuiz += subject.ClassQuiz.Attempted
		sessionWiseBreakdown := subject.SessionWiseBreakdown
		for _, sessions := range sessionWiseBreakdown {
			preClassCompleted += sessions.PreClass.Completed
			preClassTotal += sessions.PreClass.Total
		}
	}
	if totalClassQuiz > 0 {
		activeInClass := float64(totalAttemptedQuiz*100) / float64(totalClassQuiz)
		learnerTags.ActiveInClass = GetPerformanceByValue(activeInClass)
	}
	if preClassTotal > 0 {
		selfDirectedLearning := float64(preClassCompleted*100) / float64(preClassTotal)
		learnerTags.SelfDirectedLearning = GetPerformanceByValue(selfDirectedLearning)
	}
	return learnerTags

}

func GetPerformanceByValue(value float64) string {
	if value >= 80 {
		return "outstanding"
	} else if value >= 70 {
		return "meets_expectation"
	} else if value >= 60 {
		return "beginning_to_improve"
	} else {
		return "needs_to_improve"
	}
}
