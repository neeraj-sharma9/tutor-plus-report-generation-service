package mpr

import (
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"golang.org/x/tools/godoc/util"
	"strings"
	"sync"
	"time"
)

const ChapterPerformanceCalloutTemplateAhead = "Ahead of %d%% students"
const ChapterPerformanceCalloutTemplateBehind = "Behind %d%% students"

var sameCityUsers []int64
var sameStateUsers []int64
var initCityUserList sync.Once
var initStateUserList sync.Once

func getMonthlyTestAssessments(neoTestClassmates map[int][]int64) []int {
	var assessmentIDs = make(util.Set)
	for key := range neoTestClassmates {
		assessmentIDs.AddOrUpdate(key)
	}
	return assessmentIDs.List()
}

func contains(list []string, key string) bool {
	for _, item := range list {
		if item == key {
			return true
		}
	}
	return false
}

func getSubjectName(grade int, subject string) string {
	s := strings.ToLower(subject)
	if grade >= 4 && grade < 9 {
		switch s {
		case "mathematics":
			return "Mathematics"
		case "science":
			return "Science"
		case "biology":
			return "Science"
		case "physics":
			return "Science"
		case "chemistry":
			return "Science"
		}
	} else {
		switch s {
		case "mathematics":
			return "Mathematics"
		case "biology":
			return "Biology"
		case "physics":
			return "Physics"
		case "chemistry":
			return "Chemistry"
		}
	}
	return subject
}

func (report *ProgressData) SetMonthlyTestPerformance(mprReq *MPRReq, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello SetMonthlyTestPerformance ----%+v", report))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetMonthlyTestPerformance paniced - %s", r)
		}
		wg.Done()
	}()
	totalAssessmentIds := getMonthlyTestAssessments(mprReq.UserDetailsResponse.MonthlyExamClassmates)
	totalAssessmentIds = assessmentIdsWithoutSubjectiveAssessment(totalAssessmentIds)
	var attendedAssessments = make(util.Set)
	attendedAssessments.AddMulti(GetAttendedAssessments(totalAssessmentIds, mprReq)...)
	var monthlyTests []MonthlyTestPerformance
	for _, assessmentId := range totalAssessmentIds {
		var monthlyTestPerformance = &MonthlyTestPerformance{Attended: false, AssessmentId: assessmentId}
		if attendedAssessments.Has(assessmentId) {
			monthlyTestPerformance.Attended = true
		}
		var waitGroup sync.WaitGroup
		waitGroup.Add(10)
		go monthlyTestPerformance.setAssessmentQuestions(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setAssessmentTime(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setAssessmentTimeTaken(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setAssessmentQuestionAttempts(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setChapterWiseAnalysis(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setDifficultyAnalysis(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setSkillAnalysis(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setSubjectWiseScore(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.setPeersPercentageScores(mprReq, assessmentId, &waitGroup)
		go monthlyTestPerformance.SetRanks(mprReq, assessmentId, &waitGroup)
		waitGroup.Wait()
		monthlyTestPerformance.SetMonthlyTestPerformanceCallout(mprReq, assessmentId)
		monthlyTests = append(monthlyTests, *monthlyTestPerformance)
	}
	report.MonthlyTestPerformance = monthlyTests
}

func (monthlyTest *MonthlyTestPerformance) setAssessmentQuestions(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello setAssessmentQuestions ----%+v", monthlyTest))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentQuestions paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	var chapterTestedDetails = make(ChapterTestedDetails)
	chapterAssessmentQuestions := QueryAssessmentSubjectWiseChapters(assessmentID)
	for _, row := range chapterAssessmentQuestions {
		row.Subject = getSubjectName(mprReq.UserDetailsResponse.UserInfo.Grade, row.Subject)
		if _, found := chapterTestedDetails[row.Subject]; found {
			if !contains(chapterTestedDetails[row.Subject], row.Chapter) {
				chapterTestedDetails[row.Subject] = append(chapterTestedDetails[row.Subject], row.Chapter)
			}
		} else {
			chapterTestedDetails[row.Subject] = []string{row.Chapter}
		}
		monthlyTest.TotalQuestions += row.Count
	}
	monthlyTest.ChapterTestedDetails = chapterTestedDetails
	monthlyTest.ChapterTested = len(chapterAssessmentQuestions)
	log.Debugf(fmt.Sprintf("---Bye setAssessmentQuestions ----%+v", monthlyTest))
}

func QueryAssessmentSubjectWiseChapters(assessmentID int) []ChapterAssessmentQuestions {
	var assessmentQuestions []ChapterAssessmentQuestions
	query, args, err := sqlx.In(AssessmentQuestionsQuery,
		assessmentID)
	db.CheckError(err, "sqlx-QueryAssessmentSubjectWiseChapters")
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&assessmentQuestions, query, args...)
	db.CheckError(err, "QueryAssessmentSubjectWiseChapters", query)
	return assessmentQuestions
}

func (monthlyTest *MonthlyTestPerformance) setAssessmentTime(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello setAssessmentTime ----%+v", monthlyTest))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentTime paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	monthlyTest.TotalExamTime = 60
	for _, assessment := range QueryAssessmentTime(assessmentID) {
		if assessment.TotalAllowedTime.Valid {
			monthlyTest.TotalExamTime = int(assessment.TotalAllowedTime.Int32) / 60
		}
	}
	monthlyTest.NextExamDate = mprReq.UserDetailsResponse.NextMonthlyExamDate
}

func QueryAssessmentTime(assessmentID int) []AssessmentTime {
	var assessmentTime []AssessmentTime
	query, args, err := sqlx.In(AssessmentTimeQuery,
		assessmentID)
	db.CheckError(err, "sqlx-QueryAssessmentTime")
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&assessmentTime, query, args...)
	db.CheckError(err, "QueryAssessmentTime", query)
	return assessmentTime
}

func (monthlyTest *MonthlyTestPerformance) setAssessmentTimeTaken(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello setAssessmentTimeTaken ----%+v", monthlyTest))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentTimeTaken paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		for _, assessment := range QueryAssessmentTimeTaken(assessmentID, mprReq) {
			if assessment.ExamDate.Valid {
				monthlyTest.ExamDate = assessment.ExamDate.String
			}
			if assessment.TimeTaken.Valid {
				monthlyTest.TimeTaken = int(assessment.TimeTaken.Int32)
			}
			if assessment.PercentageScore.Valid {
				monthlyTest.PercentageScores.User = assessment.PercentageScore.Float64
			}
		}
	} else {
		for _, subject := range mprReq.UserDetailsResponse.Subjects {
			for _, class := range subject.Classes {
				emptyTestRequisite := TestRequisite{}
				if class.TestRequisites != emptyTestRequisite && class.TestRequisites.Assessment == assessmentID {
					monthlyTest.ExamDate = class.TestRequisites.SessionDate
					break
				}
			}
		}
	}
	log.Debugf(fmt.Sprintf("---Bye setAssessmentTimeTaken ----%+v", monthlyTest))
}

func QueryAssessmentTimeTaken(assessmentID int, mprReq *MPRReq) []AssessmentAttemptInfo {
	var assessmentInfo []AssessmentAttemptInfo
	query, args, err := sqlx.In(AssessmentAttemptDetails,
		mprReq.UserId,
		assessmentID)
	db.CheckError(err, "sqlx-QueryAssessmentTimeTaken")
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&assessmentInfo, query, args...)
	db.CheckError(err, "QueryAssessmentTimeTaken", query)
	return assessmentInfo
}

func (monthlyTest *MonthlyTestPerformance) setAssessmentQuestionAttempts(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello setAssessmentQuestionAttempts ----%+v", monthlyTest))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentQuestionAttempts paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		for _, row := range QueryAssessmentQuestionAttempts(assessmentID, mprReq) {
			if row.IsCorrect.Valid {
				if strings.ToLower(row.IsCorrect.String) == "true" {
					monthlyTest.CorrectAnswer += row.Count
					monthlyTest.QuestionAttempted += row.Count
				} else if strings.ToLower(row.IsCorrect.String) == "false" {
					monthlyTest.QuestionAttempted += row.Count
				}
			}
		}
	}
	log.Debugf(fmt.Sprintf("---Bye setAssessmentQuestionAttempts ----%+v", monthlyTest))
}

func QueryAssessmentQuestionAttempts(assessmentID int, mprReq *MPRReq) []QuestionsAttempted {
	var questionsAttempted []QuestionsAttempted
	query, args, err := sqlx.In(AssessmentQuestionsAttempts,
		assessmentID, mprReq.UserId)
	db.CheckError(err, "sqlx-QueryAssessmentQuestionAttempts")
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&questionsAttempted, query, args...)
	db.CheckError(err, "QueryAssessmentQuestionAttempts", query)
	return questionsAttempted
}

func GetChapterPerformanceCallout(assessmentId int, chapterName string, mprReq *MPRReq) float64 {
	log.Debugf("----hello GetChapterPerformanceCallout---")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("GetChapterPerformanceCallout paniced - %s", r)
		}
	}()
	var rankMap map[int64]float64
	dbData := QueryChapterWisePerformanceCallout(assessmentId, chapterName)
	log.Debugf(fmt.Sprintf("DB response %+v, %+v, %+v", assessmentId, chapterName, dbData))
	rankMap = make(map[int64]float64, len(dbData))
	for _, item := range dbData {
		if item.UserId.Valid {
			rankMap[item.UserId.Int64] = item.AheadOf
		}
	}
	percentile := 0.0
	if per, found := rankMap[mprReq.UserId]; found {
		percentile = per
	}
	log.Debugf("----bye bye GetChapterPerformanceCallout---")
	return percentile
}

func (monthlyTest *MonthlyTestPerformance) setChapterWiseAnalysis(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf(fmt.Sprintf("---Hello setChapterWiseAnalysis ----%+v", monthlyTest))
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setChapterWiseAnalysis paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		var chapterWiseAnalysis []ChapterWiseAnalysis
		subjectDataMap := QueryChapterWiseData(assessmentID, mprReq)
		log.Debugf(fmt.Sprintf("setChapterWiseAnalysis map of subject wise chapters: %+v", subjectDataMap))
		for subject, data := range subjectDataMap {
			var chapters []Chapters
			for _, chapter := range data {
				log.Debugf(fmt.Sprintf("Getting data for chapter : %+v", chapter))
				percentile := GetChapterPerformanceCallout(assessmentID, chapter.Chapter, mprReq)
				if percentile >= 50.0 {
					chapter.Direction = "up"
					chapter.PerformanceCallout = fmt.Sprintf(ChapterPerformanceCalloutTemplateAhead, int(percentile+0.5))
				} else {
					chapter.Direction = "down"
					per := 100.00 - percentile
					chapter.PerformanceCallout = fmt.Sprintf(ChapterPerformanceCalloutTemplateBehind, int(per+0.5))
				}
				chapters = append(chapters, chapter)
			}
			chapterWiseAnalysis = append(chapterWiseAnalysis, ChapterWiseAnalysis{Subject: subject, Chapters: chapters})
		}
		monthlyTest.ChapterWiseAnalysis = chapterWiseAnalysis
	}
	log.Debugf(fmt.Sprintf("---Bye setChapterWiseAnalysis ----%+v", monthlyTest))
}

func QueryChapterWiseData(assessmentId int, mprReq *MPRReq) map[string][]Chapters {
	query, args, err := sqlx.In(ChapterWisePerformanceQuery,
		assessmentId,
		mprReq.UserId)
	db.CheckError(err, "sqlx-ChapterWisePerformanceQuery")

	var testInfos []ChapterWisePerformanceCursor
	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&testInfos, query, args...)
	dgSubjectMap := map[string][]Chapters{}
	db.CheckError(err, "QueryChapterWiseData")

	for _, row := range testInfos {
		row.Subject = getSubjectName(mprReq.UserDetailsResponse.UserInfo.Grade, row.Subject)
		var dg *Chapters
		if _, found := dgSubjectMap[row.Subject]; !found {
			dgSubjectMap[row.Subject] = []Chapters{}
		}
		for idx, chapter := range dgSubjectMap[row.Subject] {
			if chapter.Chapter == row.Chapter {
				dg = &dgSubjectMap[row.Subject][idx]
			}
		}
		if dg == nil {
			dgSubjectMap[row.Subject] = append(dgSubjectMap[row.Subject], Chapters{Chapter: row.Chapter})
			dg = &dgSubjectMap[row.Subject][len(dgSubjectMap[row.Subject])-1]
		}
		if row.IsCorrect.Valid {
			if strings.ToLower(row.IsCorrect.String) == "true" {
				dg.Correct += row.Count
			} else if strings.ToLower(row.IsCorrect.String) == "false" {
				dg.Incorrect += row.Count
			} else {
				dg.NotAttempted += row.Count
			}
		} else {
			dg.NotAttempted += row.Count
		}
	}
	return dgSubjectMap
}

func QueryChapterWisePerformanceCallout(assessmentId int, chapterName string) []ChapterWiseRankCursor {

	query, args, err := sqlx.In(ChapterWiseRankQuery,
		assessmentId,
		chapterName)
	db.CheckError(err, "sqlx-ChapterWiseRankQuery")

	var chapterWiseRankInfo []ChapterWiseRankCursor
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&chapterWiseRankInfo, query, args...)

	db.CheckError(err, "QueryChapterWisePerformanceCallout")

	return chapterWiseRankInfo
}

func (monthlyTest *MonthlyTestPerformance) setDifficultyAnalysis(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf("----hello setDifficultyAnalysis ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setDifficultyAnalysis paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		difficultyList := QueryDifficultyData(assessmentID, mprReq)
		log.Debugf(fmt.Sprintf("setDifficultyAnalysis ::::: difficultyList %+v ", difficultyList))
		difficultyAnalysis := DifficultyAnalysis{}
		difficultyAnalysis.DifficultyGraph = difficultyList
		difficultyAnalysis.CallOut = strings.Replace(callout.GetDifficultyAnalysisCallout(difficultyList), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		monthlyTest.DifficultyAnalysis = difficultyAnalysis
	}
	log.Debugf("----bye bye setDifficultyAnalysis ----")
}

func QueryDifficultyData(assessmentId int, mprReq *MPRReq) []util.DifficultyGraph {
	query, args, err := sqlx.In(DifficultyGraphQuery,
		assessmentId,
		mprReq.UserId)
	db.CheckError(err, "sqlx-DifficultyGraphQuery")

	var testInfos []DifficultySkillCursor
	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&testInfos, query, args...)
	db.CheckError(err, "QueryDifficultyData")
	var dgList []util.DifficultyGraph
	log.Debugf("QueryDifficultyData: DB result: %+v", testInfos)
	for _, row := range testInfos {
		if row.Label.Valid {
			var dg *util.DifficultyGraph
			for idx, difficulty := range dgList {
				if difficulty.Label == row.Label.String {
					dg = &dgList[idx]
				}
			}
			if dg == nil {
				dgList = append(dgList, util.DifficultyGraph{Label: row.Label.String})
				dg = &dgList[len(dgList)-1]
			}
			if row.IsCorrect.Valid {
				if strings.ToLower(row.IsCorrect.String) == "true" {
					dg.Correct += row.Count
				} else if strings.ToLower(row.IsCorrect.String) == "false" {
					dg.Incorrect += row.Count
				} else {
					dg.NotAttempted += row.Count
				}
			} else {
				dg.NotAttempted += row.Count
			}
		}
	}
	return dgList
}

func (monthlyTest *MonthlyTestPerformance) setSkillAnalysis(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf("----hello setSkillAnalysis ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSkillAnalysis paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		skillList := QuerySkillData(assessmentID, mprReq)
		log.Debugf(fmt.Sprintf("setSkillAnalysis ::::: skillList %+v ", skillList))
		skillAnalysis := SkillAnalysis{}
		skillAnalysis.SkillAnalysisGraph = skillList
		skillAnalysis.CallOut = strings.Replace(callout.GetSkillAnalysisCallout(skillList), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		monthlyTest.SkillAnalysis = skillAnalysis
	}
	log.Debugf("----bye bye setSkillAnalysis ----")
}

func QuerySkillData(assessmentId int, mprReq *MPRReq) []util.SkillAnalysisGraph {
	query, args, err := sqlx.In(SkillAnalysisQuery,
		assessmentId,
		mprReq.UserId)
	db.CheckError(err, "sqlx-DifficultyGraphQuery")

	var testInfos []DifficultySkillCursor
	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&testInfos, query, args...)
	db.CheckError(err, "QuerySkillData")
	var skillList []util.SkillAnalysisGraph
	log.Debugf("QuerySkillData: DB result: %+v", testInfos)
	for _, row := range testInfos {
		if row.Label.Valid {
			var skill *util.SkillAnalysisGraph
			for idx, _skill := range skillList {
				if _skill.Label == row.Label.String {
					skill = &skillList[idx]
				}
			}
			if skill == nil {
				skillList = append(skillList, util.SkillAnalysisGraph{Label: row.Label.String})
				skill = &skillList[len(skillList)-1]
			}
			if row.IsCorrect.Valid {
				if strings.ToLower(row.IsCorrect.String) == "true" {
					skill.Correct += row.Count
				} else if strings.ToLower(row.IsCorrect.String) == "false" {
					skill.Incorrect += row.Count
				} else {
					skill.NotAttempted += row.Count
				}
			} else {
				skill.NotAttempted += row.Count
			}
		}
	}
	return skillList
}

func (monthlyTest *MonthlyTestPerformance) setSubjectWiseScore(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf("----hello setSubjectWiseScore ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSubjectWiseScore paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	type SubjectWiseQuestions struct {
		Correct int
		Total   int
	}
	if monthlyTest.Attended {
		subjectQuestionsMap := map[string]SubjectWiseQuestions{}
		for _, subjectAttemptedQuestion := range QuerySubjectWiseScore(assessmentID, mprReq) {
			if subjectAttemptedQuestion.Subject.Valid {
				subjectName := getSubjectName(mprReq.UserDetailsResponse.UserInfo.Grade, subjectAttemptedQuestion.Subject.String)
				if _, found := subjectQuestionsMap[subjectName]; !found {
					subjectQuestionsMap[subjectName] = SubjectWiseQuestions{Correct: 0, Total: 0}
				}
				var subject = subjectQuestionsMap[subjectName]
				if subjectAttemptedQuestion.IsCorrect.Valid {
					if subjectAttemptedQuestion.IsCorrect.String == "true" {
						subject.Correct += subjectAttemptedQuestion.Count
						subject.Total += subjectAttemptedQuestion.Count
					} else {
						subject.Total += subjectAttemptedQuestion.Count
					}
				} else {
					subject.Total += subjectAttemptedQuestion.Count
				}
				subjectQuestionsMap[subjectName] = subject
			}
		}

		var subjectWiseScorePercentage []SubjectWiseScore
		countGt80 := 0
		countLt40 := 0
		for subject, questions := range subjectQuestionsMap {
			score := (float64(questions.Correct) * 100.0) / float64(questions.Total)
			if score < 40 {
				countLt40 += 1
			} else if score > 80 {
				countGt80 += 1
			}
			subjectWiseScorePercentage = append(subjectWiseScorePercentage, SubjectWiseScore{subject: fmt.Sprintf("%v", score)})
		}
		monthlyTest.SubjectChapterCallout = strings.Replace(callout.GetSubjectChapterCallout(countGt80, countLt40, len(subjectWiseScorePercentage)), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		monthlyTest.SubjectWiseScore = subjectWiseScorePercentage
	}
}

func QuerySubjectWiseScore(assessmentId int, mprReq *MPRReq) []SubjectAttemptsCount {
	query, args, err := sqlx.In(SubjectWiseScoreQuery,
		assessmentId,
		mprReq.UserId)
	db.CheckError(err, "sqlx-QuerySubjectWiseScore")

	var testInfos []SubjectAttemptsCount
	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&testInfos, query, args...)
	db.CheckError(err, "QuerySubjectWiseScore")
	return testInfos
}

func (monthlyTest *MonthlyTestPerformance) setPeersPercentageScores(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf("----hello setPeersPercentageScores ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setPeersPercentageScores paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		var waitGroup sync.WaitGroup
		waitGroup.Add(4)
		go monthlyTest.PercentageScores.QueryClassAverage(assessmentID, mprReq, &waitGroup)
		go monthlyTest.PercentageScores.QueryCityAverage(assessmentID, mprReq, &waitGroup)
		go monthlyTest.PercentageScores.QueryStateAverage(assessmentID, mprReq, &waitGroup)
		go monthlyTest.PercentageScores.QueryNationalAverage(assessmentID, mprReq, &waitGroup)
		waitGroup.Wait()
	}
	log.Debugf("----Bye setPeersPercentageScores ---- %+v", monthlyTest.PercentageScores)
}

func (percentageScores *PercentageScores) QueryClassAverage(assessmentID int, mprReq *MPRReq, wg *sync.WaitGroup) {
	log.Debugf("----hello QueryClassAverage ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryClassAverage paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	userList := mprReq.UserDetailsResponse.MonthlyExamClassmates[assessmentID]
	if len(userList) > 0 {
		query, args, err := sqlx.In(MultiUserAverageScoreQuery, userList, assessmentID)
		db.CheckError(err, "sqlx-QueryClassAverage")

		var testInfos []MultiUserAverageScore
		query = db.GetPGDbReader().Rebind(query)

		err = db.GetPGDbReader().
			Select(&testInfos, query, args...)
		db.CheckError(err, "QueryClassAverage")
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				percentageScores.Class = testInfos[0].Average.Float64
			}
		}
	}
}

func (percentageScores *PercentageScores) QueryCityAverage(assessmentID int, mprReq *MPRReq, wg *sync.WaitGroup) {
	log.Debugf("----hello QueryCityAverage ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryCityAverage paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	userList := getSameCityUsers(mprReq)
	if len(userList) > 0 {
		query, args, err := sqlx.In(MultiUserAverageScoreQuery, userList, assessmentID)
		db.CheckError(err, "sqlx-QueryCityAverage")

		var testInfos []MultiUserAverageScore
		query = db.GetPGDbReader().Rebind(query)

		err = db.GetPGDbReader().
			Select(&testInfos, query, args...)
		db.CheckError(err, "QueryCityAverage")
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				percentageScores.City = testInfos[0].Average.Float64
			}
		}
	}
	log.Debugf(fmt.Sprintf("----Bye QueryCityAverage ---- %+v", percentageScores))
}

func (percentageScores *PercentageScores) QueryStateAverage(assessmentID int, mprReq *MPRReq, wg *sync.WaitGroup) {
	log.Debugf("----hello QueryStateAverage ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryStateAverage paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	userList := getSameStateUsers(mprReq)
	if len(userList) > 0 {
		query, args, err := sqlx.In(MultiUserAverageScoreQuery, userList, assessmentID)
		db.CheckError(err, "sqlx-QueryStateAverage")

		var testInfos []MultiUserAverageScore
		query = db.GetPGDbReader().Rebind(query)

		err = db.GetPGDbReader().
			Select(&testInfos, query, args...)
		db.CheckError(err, "QueryStateAverage")
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				percentageScores.State = testInfos[0].Average.Float64
			}
		}
	}
	log.Debugf(fmt.Sprintf("----Bye QueryStateAverage ---- %+v", percentageScores))
}

func (percentageScores *PercentageScores) QueryNationalAverage(assessmentID int, mprReq *MPRReq, wg *sync.WaitGroup) {
	log.Debugf("----hello QueryNationalAverage ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryNationalAverage paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	userList := mprReq.UserDetailsResponse.SameBatchUserList
	if len(userList) > 0 {
		query, args, err := sqlx.In(MultiUserAverageScoreQuery, userList, assessmentID)
		db.CheckError(err, "sqlx-QueryNationalAverage")

		var testInfos []MultiUserAverageScore
		query = db.GetPGDbReader().Rebind(query)

		err = db.GetPGDbReader().
			Select(&testInfos, query, args...)
		db.CheckError(err, "QueryNationalAverage")
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				percentageScores.National = testInfos[0].Average.Float64
			}
		}
	}
}

func getSameCityUsers(mprReq *MPRReq) []int64 {
	initCityUserList.Do(func() {
		var err error
		sameCityUsers, err = QuerySameCityUsers(mprReq.UserId, mprReq.UserDetailsResponse.SameBatchUserList)
		if err != nil {
			log.GetLogger().Errorln(err)
		}
	})
	return sameCityUsers
}

func QuerySameCityUsers(userId int64, maxUserList []int64) ([]int64, error) {
	var userList []UserList
	query, args, err := sqlx.In(CityUserListQuery,
		userId, maxUserList)
	db.CheckError(err, "sqlx-CityUserListQuery")

	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&userList, query, args...)
	db.CheckError(err, "QuerySameCityUsers")
	var result []int64
	if len(userList) > 0 {
		for _, item := range userList {
			result = append(result, item.UserId)
		}
	}
	return result, err
}

func getSameStateUsers(mprReq *MPRReq) []int64 {
	initStateUserList.Do(func() {
		var err error
		sameStateUsers, err = QuerySameStateUsers(mprReq.UserId, mprReq.UserDetailsResponse.SameBatchUserList)
		if err != nil {
			log.GetLogger().Fatal(err)
		}
	})
	return sameStateUsers
}

func QuerySameStateUsers(userId int64, maxUserList []int64) ([]int64, error) {
	var userList []UserList
	query, args, err := sqlx.In(StateUserListQuery,
		userId, maxUserList)
	db.CheckError(err, "sqlx-StateUserListQuery")

	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&userList, query, args...)
	db.CheckError(err, "QuerySameStateUsers")
	var result []int64
	if len(userList) > 0 {
		for _, item := range userList {
			result = append(result, item.UserId)
		}
	}
	return result, err
}

func QueryUserRank(assessmentID int, userId int64, userList []int64) int {
	var userRank []UserRank
	rank := 0
	if len(userList) == 0 {
		return rank
	}
	query, args, err := sqlx.In(RankQuery,
		assessmentID, userList, userId)
	db.CheckError(err, "sqlx-RankQuery")

	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&userRank, query, args...)
	db.CheckError(err, "QueryUserRank")
	if len(userRank) > 0 {
		if userRank[0].Rank.Valid {
			rank = int(userRank[0].Rank.Int64)
		}
	}
	return rank
}

func QueryUserCity(userId int64) string {
	var userRank []UserCity
	rank := ""
	query, args, err := sqlx.In(UserCityQuery, userId)
	db.CheckError(err, "sqlx-UserCityQuery")

	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&userRank, query, args...)
	db.CheckError(err, "QueryUserCity")
	if len(userRank) > 0 {
		if userRank[0].City.Valid {
			rank = userRank[0].City.String
		}
	}
	return rank
}

func (monthlyTest *MonthlyTestPerformance) SetRanks(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	log.Debugf("----hello SetRanks ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetRanks paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if monthlyTest.Attended {
		monthlyTest.Ranks.National = QueryUserRank(assessmentID, mprReq.UserId, mprReq.UserDetailsResponse.SameBatchUserList)
		monthlyTest.Ranks.State = QueryUserRank(assessmentID, mprReq.UserId, getSameStateUsers(mprReq))
		monthlyTest.Ranks.City = QueryUserRank(assessmentID, mprReq.UserId, getSameCityUsers(mprReq))
		monthlyTest.Ranks.Class = QueryUserRank(assessmentID, mprReq.UserId, mprReq.UserDetailsResponse.MonthlyExamClassmates[assessmentID])
		monthlyTest.City = QueryUserCity(mprReq.UserId)
	}
	log.Debugf(fmt.Sprintf("----Bye SetRanks ---- %+v", monthlyTest))
}

func (monthlyTest *MonthlyTestPerformance) SetMonthlyTestPerformanceCallout(mprReq *MPRReq, assessmentID int) {
	log.Debugf("----hello SetMonthlyTestPerformanceCallout ----")
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetMonthlyTestPerformanceCallout paniced %d - %s", assessmentID, r)
		}
	}()
	monthlyTestPerformanceCallout := strings.Replace(callout.GetMonthlyTestPerformanceCallout(monthlyTest.Ranks.National, monthlyTest.Ranks.State,
		monthlyTest.Ranks.City, monthlyTest.Ranks.Class, monthlyTest.PercentageScores.User, monthlyTest.PercentageScores.Class,
		monthlyTest.Attended), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
	monthlyTestPerformanceCallout = strings.Replace(monthlyTestPerformanceCallout, "<City>", monthlyTest.City, -1)
	monthlyTest.MonthlyTestCallout = monthlyTestPerformanceCallout
	log.Debugf(fmt.Sprintf("----Bye SetMonthlyTestPerformanceCallout ---- %+v", monthlyTest))
}

func GetAttendedAssessments(assessmentList []int, mprReq *MPRReq) []int {
	var result []int
	var assessmentAttended []AttendedAssessment
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	log.GetLogger().Infoln("GetAttendedAssessments:: ", fromDate, toDate)
	query, args, err := sqlx.In(AssessmentsAttended, mprReq.UserId, assessmentList, fromDate, toDate)
	db.CheckError(err, "sqlx-GetAttendedAssessments")

	query = db.GetPGDbReader().Rebind(query)

	err = db.GetPGDbReader().
		Select(&assessmentAttended, query, args...)
	db.CheckError(err, "GetAttendedAssessments")
	for _, assessment := range assessmentAttended {
		if assessment.AssessmentID.Valid {
			result = append(result, int(assessment.AssessmentID.Int32))
		}
	}
	return result
}

func assessmentIdsWithoutSubjectiveAssessment(totalAssessmentIds []int) []int {
	query, args, err := sqlx.In(GetAssessmentIdsExcludingSubjectiveAssessment,
		totalAssessmentIds)
	db.CheckError(err, "sqlx-GetAssessmentIdsExcludingSubjectiveAssessmentQuery")

	var assessmentIds []int
	query = db.GetPGDbReader().Rebind(query)
	err = db.GetPGDbReader().
		Select(&assessmentIds, query, args...)

	db.CheckError(err, "GetAssessmentIdsExcludingSubjectiveAssessmentQuery")

	return assessmentIds
}
