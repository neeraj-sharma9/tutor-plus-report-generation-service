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

type MonthlyTestPerformanceService struct {
	TllmsManager *manager.TllmsManager
	MTP          *MonthlyTestPerformance
}

func getMonthlyTestAssessments(neoTestClassmates map[int][]int64) []int {
	var assessmentIDs = make(utility.Set)
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

func (mTPS *MonthlyTestPerformanceService) setAssessmentQuestions(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentQuestions paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	var chapterTestedDetails = make(ChapterTestedDetails)
	chapterAssessmentQuestions := mTPS.QueryAssessmentSubjectWiseChapters(assessmentID)
	for _, row := range chapterAssessmentQuestions {
		row.Subject = getSubjectName(mprReq.UserDetailsResponse.UserInfo.Grade, row.Subject)
		if _, found := chapterTestedDetails[row.Subject]; found {
			if !contains(chapterTestedDetails[row.Subject], row.Chapter) {
				chapterTestedDetails[row.Subject] = append(chapterTestedDetails[row.Subject], row.Chapter)
			}
		} else {
			chapterTestedDetails[row.Subject] = []string{row.Chapter}
		}
		mTPS.MTP.TotalQuestions += row.Count
	}
	mTPS.MTP.ChapterTestedDetails = chapterTestedDetails
	mTPS.MTP.ChapterTested = len(chapterAssessmentQuestions)
}

func (mTPS *MonthlyTestPerformanceService) QueryAssessmentSubjectWiseChapters(assessmentID int) []contract.ChapterAssessmentQuestions {
	return mTPS.TllmsManager.AssessmentQuestionsQuery(assessmentID)
}

func (mTPS *MonthlyTestPerformanceService) setAssessmentTime(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentTime paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	mTPS.MTP.TotalExamTime = 60
	for _, assessment := range mTPS.QueryAssessmentTime(assessmentID) {
		if assessment.TotalAllowedTime.Valid {
			mTPS.MTP.TotalExamTime = int(assessment.TotalAllowedTime.Int32) / 60
		}
	}
	mTPS.MTP.NextExamDate = mprReq.UserDetailsResponse.NextMonthlyExamDate
}

func (mTPS *MonthlyTestPerformanceService) QueryAssessmentTime(assessmentID int) []contract.AssessmentTime {
	return mTPS.TllmsManager.AssessmentTimeQuery(assessmentID)
}

func (mTPS *MonthlyTestPerformanceService) setAssessmentTimeTaken(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentTimeTaken paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if mTPS.MTP.Attended {
		for _, assessment := range mTPS.QueryAssessmentTimeTaken(assessmentID, mprReq) {
			if assessment.ExamDate.Valid {
				mTPS.MTP.ExamDate = assessment.ExamDate.String
			}
			if assessment.TimeTaken.Valid {
				mTPS.MTP.TimeTaken = int(assessment.TimeTaken.Int32)
			}
			if assessment.PercentageScore.Valid {
				mTPS.MTP.PercentageScores.User = assessment.PercentageScore.Float64
			}
		}
	} else {
		for _, subject := range mprReq.UserDetailsResponse.Subjects {
			for _, class := range subject.Classes {
				emptyTestRequisite := helper.TestRequisite{}
				if class.TestRequisites != emptyTestRequisite && class.TestRequisites.Assessment == assessmentID {
					mTPS.MTP.ExamDate = class.TestRequisites.SessionDate
					break
				}
			}
		}
	}
}

func (mTPS *MonthlyTestPerformanceService) QueryAssessmentTimeTaken(assessmentID int, mprReq *MPRReq) []contract.AssessmentAttemptInfo {
	return mTPS.TllmsManager.AssessmentAttemptDetails(assessmentID, mprReq.UserId)
}

func (mTPS *MonthlyTestPerformanceService) setAssessmentQuestionAttempts(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setAssessmentQuestionAttempts paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if mTPS.MTP.Attended {
		for _, row := range mTPS.QueryAssessmentQuestionAttempts(assessmentID, mprReq) {
			if row.IsCorrect.Valid {
				if strings.ToLower(row.IsCorrect.String) == "true" {
					mTPS.MTP.CorrectAnswer += row.Count
					mTPS.MTP.QuestionAttempted += row.Count
				} else if strings.ToLower(row.IsCorrect.String) == "false" {
					mTPS.MTP.QuestionAttempted += row.Count
				}
			}
		}
	}
}

func (mTPS *MonthlyTestPerformanceService) QueryAssessmentQuestionAttempts(assessmentID int, mprReq *MPRReq) []contract.QuestionsAttempted {
	return mTPS.TllmsManager.AssessmentQuestionsAttempts(assessmentID, mprReq.UserId)
}

func (mTPS *MonthlyTestPerformanceService) GetChapterPerformanceCallout(assessmentId int, chapterName string, mprReq *MPRReq) float64 {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("GetChapterPerformanceCallout paniced - %s", r)
		}
	}()
	var rankMap map[int64]float64
	dbData := mTPS.QueryChapterWisePerformanceCallout(assessmentId, chapterName)
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
	return percentile
}

func (mTPS *MonthlyTestPerformanceService) setChapterWiseAnalysis(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setChapterWiseAnalysis paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if mTPS.MTP.Attended {
		var chapterWiseAnalysis []ChapterWiseAnalysis
		subjectDataMap := mTPS.QueryChapterWiseData(assessmentID, mprReq)
		for subject, data := range subjectDataMap {
			var chapters []Chapters
			for _, chapter := range data {
				percentile := mTPS.GetChapterPerformanceCallout(assessmentID, chapter.Chapter, mprReq)
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
		mTPS.MTP.ChapterWiseAnalysis = chapterWiseAnalysis
	}
}

func (mTPS *MonthlyTestPerformanceService) QueryChapterWiseData(assessmentId int, mprReq *MPRReq) map[string][]Chapters {
	testInfos := mTPS.TllmsManager.ChapterWisePerformanceQuery(assessmentId, mprReq.UserId)
	dgSubjectMap := map[string][]Chapters{}

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

func (mTPS *MonthlyTestPerformanceService) QueryChapterWisePerformanceCallout(assessmentId int, chapterName string) []contract.ChapterWiseRankCursor {
	return mTPS.TllmsManager.ChapterWiseRankQuery(assessmentId, chapterName)
}

func (mTPS *MonthlyTestPerformanceService) setDifficultyAnalysis(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setDifficultyAnalysis paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if mTPS.MTP.Attended {
		difficultyList := mTPS.QueryDifficultyData(assessmentID, mprReq)
		difficultyAnalysis := DifficultyAnalysis{}
		difficultyAnalysis.DifficultyGraph = difficultyList
		difficultyAnalysis.CallOut = strings.Replace(callout.GetDifficultyAnalysisCallout(difficultyList), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		mTPS.MTP.DifficultyAnalysis = difficultyAnalysis
	}
}

func (mTPS *MonthlyTestPerformanceService) QueryDifficultyData(assessmentId int, mprReq *MPRReq) []helper.DifficultyGraph {
	testInfos := mTPS.TllmsManager.DifficultyGraphQuery(assessmentId, mprReq.UserId)
	var dgList []helper.DifficultyGraph
	for _, row := range testInfos {
		if row.Label.Valid {
			var dg *helper.DifficultyGraph
			for idx, difficulty := range dgList {
				if difficulty.Label == row.Label.String {
					dg = &dgList[idx]
				}
			}
			if dg == nil {
				dgList = append(dgList, helper.DifficultyGraph{Label: row.Label.String})
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

func (mTPS *MonthlyTestPerformanceService) setSkillAnalysis(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setSkillAnalysis paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if mTPS.MTP.Attended {
		skillList := mTPS.QuerySkillData(assessmentID, mprReq)
		skillAnalysis := SkillAnalysis{}
		skillAnalysis.SkillAnalysisGraph = skillList
		skillAnalysis.CallOut = strings.Replace(callout.GetSkillAnalysisCallout(skillList), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		mTPS.MTP.SkillAnalysis = skillAnalysis
	}
}

func (mTPS *MonthlyTestPerformanceService) QuerySkillData(assessmentId int, mprReq *MPRReq) []helper.SkillAnalysisGraph {
	testInfos := mTPS.TllmsManager.SkillAnalysisQuery(assessmentId, mprReq.UserId)
	var skillList []helper.SkillAnalysisGraph
	for _, row := range testInfos {
		if row.Label.Valid {
			var skill *helper.SkillAnalysisGraph
			for idx, _skill := range skillList {
				if _skill.Label == row.Label.String {
					skill = &skillList[idx]
				}
			}
			if skill == nil {
				skillList = append(skillList, helper.SkillAnalysisGraph{Label: row.Label.String})
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

func (mTPS *MonthlyTestPerformanceService) setSubjectWiseScore(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
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
	if mTPS.MTP.Attended {
		subjectQuestionsMap := map[string]SubjectWiseQuestions{}
		for _, subjectAttemptedQuestion := range mTPS.QuerySubjectWiseScore(assessmentID, mprReq) {
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
		mTPS.MTP.SubjectChapterCallout = strings.Replace(callout.GetSubjectChapterCallout(countGt80, countLt40, len(subjectWiseScorePercentage)), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
		mTPS.MTP.SubjectWiseScore = subjectWiseScorePercentage
	}
}

func (mTPS *MonthlyTestPerformanceService) QuerySubjectWiseScore(assessmentId int, mprReq *MPRReq) []contract.SubjectAttemptsCount {
	return mTPS.TllmsManager.SubjectWiseScoreQuery(assessmentId, mprReq.UserId)
}

func (mTPS *MonthlyTestPerformanceService) setPeersPercentageScores(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("setPeersPercentageScores paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	var percentageScore PercentageScores
	if mTPS.MTP.Attended {
		percentageScore.Class = mTPS.QueryClassAverage(assessmentID, mprReq)
		percentageScore.City = mTPS.QueryCityAverage(assessmentID, mprReq)
		percentageScore.State = mTPS.QueryStateAverage(assessmentID, mprReq)
		percentageScore.National = mTPS.QueryNationalAverage(assessmentID, mprReq)
	}
	mTPS.MTP.PercentageScores = percentageScore
}

func (mTPS *MonthlyTestPerformanceService) QueryClassAverage(assessmentID int, mprReq *MPRReq) float64 {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryClassAverage paniced %d - %s", assessmentID, r)
		}
	}()
	userList := mprReq.UserDetailsResponse.MonthlyExamClassmates[assessmentID]
	class := 0.0
	if len(userList) > 0 {
		testInfos := mTPS.TllmsManager.MultiUserAverageScoreQuery(assessmentID, userList)
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				class = testInfos[0].Average.Float64
			}
		}
	}
	return class
}

func (mTPS *MonthlyTestPerformanceService) QueryCityAverage(assessmentID int, mprReq *MPRReq) float64 {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryCityAverage paniced %d - %s", assessmentID, r)
		}
	}()
	userList := mTPS.getSameCityUsers(mprReq)
	city := 0.0
	if len(userList) > 0 {
		testInfos := mTPS.TllmsManager.MultiUserAverageScoreQuery(assessmentID, userList)
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				city = testInfos[0].Average.Float64
			}
		}
	}
	return city
}

func (mTPS *MonthlyTestPerformanceService) QueryStateAverage(assessmentID int, mprReq *MPRReq) float64 {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryStateAverage paniced %d - %s", assessmentID, r)
		}
	}()
	userList := mTPS.getSameStateUsers(mprReq)
	state := 0.0
	if len(userList) > 0 {
		testInfos := mTPS.TllmsManager.MultiUserAverageScoreQuery(assessmentID, userList)
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				state = testInfos[0].Average.Float64
			}
		}
	}
	return state
}

func (mTPS *MonthlyTestPerformanceService) QueryNationalAverage(assessmentID int, mprReq *MPRReq) float64 {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("QueryNationalAverage paniced %d - %s", assessmentID, r)
		}
	}()
	userList := mprReq.UserDetailsResponse.SameBatchUserList
	national := 0.0
	if len(userList) > 0 {
		testInfos := mTPS.TllmsManager.MultiUserAverageScoreQuery(assessmentID, userList)
		if len(testInfos) > 0 {
			if testInfos[0].Average.Valid {
				national = testInfos[0].Average.Float64
			}
		}
	}
	return national
}

func (mTPS *MonthlyTestPerformanceService) getSameCityUsers(mprReq *MPRReq) []int64 {
	initCityUserList.Do(func() {
		var err error
		sameCityUsers, err = mTPS.QuerySameCityUsers(mprReq.UserId, mprReq.UserDetailsResponse.SameBatchUserList)
		if err != nil {
			logger.Log.Sugar().Errorln(err)
		}
	})
	return sameCityUsers
}

func (mTPS *MonthlyTestPerformanceService) QuerySameCityUsers(userId int64, maxUserList []int64) ([]int64, error) {
	return mTPS.TllmsManager.CityUserListQuery(userId, maxUserList), nil
}

func (mTPS *MonthlyTestPerformanceService) getSameStateUsers(mprReq *MPRReq) []int64 {
	initStateUserList.Do(func() {
		var err error
		sameStateUsers, err = mTPS.QuerySameStateUsers(mprReq.UserId, mprReq.UserDetailsResponse.SameBatchUserList)
		if err != nil {
			logger.Log.Sugar().Fatal(err)
		}
	})
	return sameStateUsers
}

func (mTPS *MonthlyTestPerformanceService) QuerySameStateUsers(userId int64, maxUserList []int64) ([]int64, error) {
	return mTPS.TllmsManager.StateUserListQuery(userId, maxUserList), nil
}

func (mTPS *MonthlyTestPerformanceService) QueryUserRank(assessmentID int, userId int64, userList []int64) int {
	userRank := mTPS.TllmsManager.RankQuery(assessmentID, userId, userList)
	rank := 0
	if len(userRank) > 0 {
		if userRank[0].Rank.Valid {
			rank = int(userRank[0].Rank.Int64)
		}
	}
	return rank
}

func (mTPS *MonthlyTestPerformanceService) QueryUserCity(userId int64) string {
	userCity := mTPS.TllmsManager.UserCityQuery(userId)
	city := ""
	if len(userCity) > 0 {
		if userCity[0].City.Valid {
			city = userCity[0].City.String
		}
	}
	return city
}

func (mTPS *MonthlyTestPerformanceService) SetRanks(mprReq *MPRReq, assessmentID int, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetRanks paniced %d - %s", assessmentID, r)
		}
		wg.Done()
	}()
	if mTPS.MTP.Attended {
		mTPS.MTP.Ranks.National = mTPS.QueryUserRank(assessmentID, mprReq.UserId, mprReq.UserDetailsResponse.SameBatchUserList)
		mTPS.MTP.Ranks.State = mTPS.QueryUserRank(assessmentID, mprReq.UserId, mTPS.getSameStateUsers(mprReq))
		mTPS.MTP.Ranks.City = mTPS.QueryUserRank(assessmentID, mprReq.UserId, mTPS.getSameCityUsers(mprReq))
		mTPS.MTP.Ranks.Class = mTPS.QueryUserRank(assessmentID, mprReq.UserId, mprReq.UserDetailsResponse.MonthlyExamClassmates[assessmentID])
		mTPS.MTP.City = mTPS.QueryUserCity(mprReq.UserId)
	}
}

func (mTPS *MonthlyTestPerformanceService) SetMonthlyTestPerformanceCallout(mprReq *MPRReq, assessmentID int) {
	defer func() {
		if r := recover(); r != nil {
			mprReq.ReqStatus = false
			mprReq.ErrorMsg = fmt.Sprintf("SetMonthlyTestPerformanceCallout paniced %d - %s", assessmentID, r)
		}
	}()
	monthlyTestPerformanceCallout := strings.Replace(callout.GetMonthlyTestPerformanceCallout(mTPS.MTP.Ranks.National, mTPS.MTP.Ranks.State,
		mTPS.MTP.Ranks.City, mTPS.MTP.Ranks.Class, mTPS.MTP.PercentageScores.User, mTPS.MTP.PercentageScores.Class,
		mTPS.MTP.Attended), "<User>", mprReq.UserDetailsResponse.UserInfo.Name, -1)
	monthlyTestPerformanceCallout = strings.Replace(monthlyTestPerformanceCallout, "<City>", mTPS.MTP.City, -1)
	mTPS.MTP.MonthlyTestCallout = monthlyTestPerformanceCallout
}

func (mTPS *MonthlyTestPerformanceService) GetAttendedAssessments(assessmentList []int, mprReq *MPRReq) []int {
	var result []int
	fromDate := time.Unix(mprReq.EpchFrmDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	toDate := time.Unix(mprReq.EpchToDate-constant.IST_OFFSET, 0).Format(constant.TIME_LAYOUT)
	logger.Log.Sugar().Infoln("GetAttendedAssessments:: ", fromDate, toDate)
	assessmentAttended := mTPS.TllmsManager.AssessmentsAttended(mprReq.UserId, assessmentList, fromDate, toDate)
	for _, assessment := range assessmentAttended {
		if assessment.AssessmentID.Valid {
			result = append(result, int(assessment.AssessmentID.Int32))
		}
	}
	return result
}

func (mTPS *MonthlyTestPerformanceService) assessmentIdsWithoutSubjectiveAssessment(totalAssessmentIds []int) []int {
	return mTPS.TllmsManager.GetAssessmentIdsExcludingSubjectiveAssessment(totalAssessmentIds)
}
