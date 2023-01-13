package manager

import (
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/config"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/contract"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/utility"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ITllmsManager interface {
}

type TllmsManager struct {
	db *gorm.DB
}

func TllmsManagerInitializer(config *config.Config, zap *zap.Logger) *TllmsManager {
	dsn := utility.GetPostgresDSN(config.ReplicaDbHost, config.ReplicaDbPort, config.ReplicaDbName,
		config.ReplicaDbUser, config.ReplicaDbPassword)
	pg, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.Sugar().Errorf("error connecting tllms db: %v", err)
	}
	return &TllmsManager{db: pg}
}

func (tM *TllmsManager) SessionWiseAssignmentsCount(assessmentIds []int, assigneeId int64) int64 {
	var count int64
	if err := tM.db.Table("asssignments").Where("assessment_id in (?)", assessmentIds).
		Where("assignee_id = (?)", assigneeId).Count(&count).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SessionWiseAssignmentsCount query: %v, for asessmentIds: %v and "+
			"assigneeId: %v", err, assessmentIds, assigneeId)
	}
	return count
}

func (tM *TllmsManager) SessionWiseLearnJourneyCount(journeyIds []int, userId int64) []contract.AssignmentsCompletionCursor {
	var result []contract.AssignmentsCompletionCursor
	if err := tM.db.Table("learn_journey_visits").Where("journey_id in (?)", journeyIds).
		Where("user_id = (?)", userId).Select("SUM(CASE WHEN is_completed = true THEN 1 ELSE 0 END) AS completed," +
		" SUM(CASE WHEN is_completed = true OR is_completed = false THEN 1 ELSE 0 END) AS attempted").Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SessionWiseLearnJourneyCount query: %v, for journeyIds: %v and "+
			"userId: %v", err, journeyIds, userId)
	}
	return result
}

func (tM *TllmsManager) SubjectWiseAssignmentsScore(assessmentIds []int, userIds []int64, fromDate, toDate string) []contract.AssignmentScoreCursor {
	var result []contract.AssignmentScoreCursor
	if err := tM.db.Table("assignments").Where("assessment_id in (?)", assessmentIds).
		Where("assignee_id in (?)", userIds).Where("started_at BETWEEN (?) AND (?)", fromDate, toDate).
		Select("AVG(percentage_score) AS percentage_score, COUNT(completed_at) as completed").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SubjectWiseAssignmentsScore query: %v, for asessmentIds: %v and "+
			"userIds: %v, fromDate: %v, toDate: %v", err, assessmentIds, userIds, fromDate, toDate)
	}
	return result
}

func (tM *TllmsManager) SubjectWiseAssignmentsQuestions(assessmentIds []int, userId int64) []contract.AssignmentQuestionsCursor {
	var result []contract.AssignmentQuestionsCursor
	if err := tM.db.Table("question_attempts").Where("assessment_id in (?)", assessmentIds).
		Where("user_id = (?)", userId).Group("is_correct").
		Select("is_correct, count(id)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SubjectWiseAssignmentsQuestions query: %v, for asessmentIds: %v and "+
			"userId: %v", err, assessmentIds, userId)
	}
	return result
}

func (tM *TllmsManager) SubjectWiseAssignmentsAverage(assessmentIds []int, userIds []int64, fromDate,
	toDate string) []contract.AssignmentScoreCursor {
	var result []contract.AssignmentScoreCursor
	if err := tM.db.Table("assignments").Where("assessment_id in (?)", assessmentIds).
		Where("assignee_id in (?)", userIds).Where("started_at BETWEEN (?) AND (?)", fromDate, toDate).
		Select("AVG(percentage_score) AS percentage_score").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SubjectWiseAssignmentsAverage query: %v, for asessmentIds: %v and "+
			"userIds: %v, fromDate: %v, toDate: %v", err, assessmentIds, userIds, fromDate, toDate)
	}
	return result
}

func (tM *TllmsManager) PostClassPerformanceTillDate(assessmentIds []int, userIds []int64) []contract.PostAssignmentTillDate {
	var result []contract.PostAssignmentTillDate
	if err := tM.db.Table("assignments").Where("assessment_id in (?)", assessmentIds).
		Where("assignee_id in (?)", userIds).Group("status").
		Select("status, COUNT(id)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in PostClassPerformanceTillDate query: %v, for asessmentIds: %v and "+
			"userIds: %v", err, assessmentIds, userIds)
	}
	return result
}

func (tM *TllmsManager) PreClassPerformanceTillDate(journeyIds []int, userId int64) []contract.PreAssignmentTillDate {
	var result []contract.PreAssignmentTillDate
	if err := tM.db.Table("learn_journey_visits").Where("journey_id in (?)", journeyIds).
		Where("user_id = (?)", userId).Group("is_completed").
		Select("is_completed, count(id)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in PreClassPerformanceTillDate query: %v, for journeyIds: %v and "+
			"userId: %v", err, journeyIds, userId)
	}
	return result
}

func (tM *TllmsManager) AssessmentQuestionsQuery(assessmentId int) []contract.ChapterAssessmentQuestions {
	var result []contract.ChapterAssessmentQuestions
	if err := tM.db.Table("assessment_questions").
		Joins("INNER JOIN questions ON questions.id = assessment_questions.question_id ").
		Joins("INNER JOIN categories c ON c.id = questions.category_id").
		Joins("INNER JOIN categories ch ON split_part(c.ancestry,'/', 2)::int = ch.id AND ch.type='ChapterCategory'").
		Where("assessment_id = (?)", assessmentId).Group("ch.name, ch.subject").
		Select("ch.name, ch.subject, count(questions.id)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in AssessmentQuestionsQuery query: %v, for assessmentId: %v", err, assessmentId)
	}
	return result
}

func (tM *TllmsManager) AssessmentTimeQuery(assessmentId int) []contract.AssessmentTime {
	var result []contract.AssessmentTime
	if err := tM.db.Table("assessments").
		Where("id = (?)", assessmentId).
		Select("time_allowed").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in AssessmentTimeQuery query: %v, for assessmentId: %v", err, assessmentId)
	}
	return result
}

func (tM *TllmsManager) AssessmentAttemptDetails(assessmentId int, userId int64) []contract.AssessmentAttemptInfo {
	var result []contract.AssessmentAttemptInfo
	if err := tM.db.Table("assignments").Where("assessment_id = (?)", assessmentId).
		Where("assignee_id = (?)", userId).
		Select("percentage_score, TO_CHAR(started_at,'DD Mon') AS exam_date," +
			"DATE_PART('hour',to_char(completed_at,'YYYY-MM-DD HH24:MI')::timestamp - to_char(started_at,'YYYY-MM-DD HH24:MI')::timestamp) *60 " +
			"+ DATE_PART('minute',to_char(completed_at,'YYYY-MM-DD HH24:MI')::timestamp - to_char(started_at,'YYYY-MM-DD HH24:MI')::timestamp) " +
			"AS time_taken").Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in AssessmentAttemptDetails query: %v, for assessmentId: %v, userId: %v",
			err, assessmentId, userId)
	}
	return result
}

func (tM *TllmsManager) AssessmentQuestionsAttempts(assessmentId int, userId int64) []contract.QuestionsAttempted {
	var result []contract.QuestionsAttempted
	if err := tM.db.Table("question_attempts").Where("assessment_id = (?)", assessmentId).
		Where("user_id = (?)", userId).Group("is_correct").
		Select("is_correct, count(id)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in AssessmentQuestionsAttempts query: %v, for asessmentId: %v and "+
			"userId: %v", err, assessmentId, userId)
	}
	return result
}

func (tM *TllmsManager) ChapterWisePerformanceQuery(assessmentId int, userId int64) []contract.ChapterWisePerformanceCursor {
	var result []contract.ChapterWisePerformanceCursor
	if err := tM.db.Table("question_attempts").
		Joins("INNER JOIN categories c ON c.id = question_attempts.category_id").
		Joins("INNER JOIN categories ch ON split_part(c.ancestry,'/', 2)::int = ch.id AND ch.type='ChapterCategory'").
		Where("assessment_id = (?)", assessmentId).Where("user_id = (?)", userId).
		Group("ch.name, is_correct, ch.subject").
		Select("DISTINCT ch.name, ch.subject, is_correct, count(*)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in ChapterWisePerformanceQuery query: %v, for assessmentId: %v, userId: %v",
			err, assessmentId, userId)
	}
	return result
}

func (tM *TllmsManager) ChapterWiseRankQuery(assessmentId int, chapterName string) []contract.ChapterWiseRankCursor {
	var result []contract.ChapterWiseRankCursor
	if err := tM.db.Exec("SELECT f.user_id, PERCENT_RANK() over (order by total)*100 as percent_rank FROM ( "+
		"SELECT DISTINCT ch.name, qa.user_id, sum(score) AS total FROM question_attempts qa "+
		"INNER JOIN categories c ON c.id = qa.category_id "+
		"INNER JOIN categories ch ON split_part(c.ancestry,'/', 2)::int = ch.id and ch.type='ChapterCategory' "+
		"WHERE assessment_id=(?) and ch.name = (?) and split_part(c.ancestry,'/', 2) != '' "+
		"GROUP BY qa.user_id, ch.name ORDER BY total desc ) as f", assessmentId, chapterName).Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in ChapterWisePerformanceQuery query: %v, for assessmentId: %v, chapterName: %v",
			err, assessmentId, chapterName)
	}
	return result
}

func (tM *TllmsManager) DifficultyGraphQuery(assessmentId int, userId int64) []contract.DifficultySkillCursor {
	var result []contract.DifficultySkillCursor
	if err := tM.db.Table("question_attempts").
		Joins("INNER JOIN questions q ON q.id=question_attempts.question_id").
		Where("assessment_id = (?)", assessmentId).Where("user_id = (?)", userId).
		Group("q.difficulty, is_correct").
		Select("SELECT CASE WHEN q.difficulty <= 1.0 THEN 'Easy' " +
			"WHEN q.difficulty > 1.0 AND q.difficulty <= 3.0 THEN 'Medium' " +
			"WHEN q.difficulty > 3.0 AND q.difficulty <= 5.0 THEN 'Hard' " +
			"END AS label, is_correct, count(*)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in DifficultyGraphQuery query: %v, for assessmentId: %v, userId: %v",
			err, assessmentId, userId)
	}
	return result
}

func (tM *TllmsManager) SkillAnalysisQuery(assessmentId int, userId int64) []contract.SkillAnalysisGraph {
	var result []contract.SkillAnalysisGraph
	if err := tM.db.Table("question_attempts").
		Joins("INNER JOIN questions q ON q.id=question_attempts.question_id").
		Joins("INNER JOIN raw_questions rq ON rq.id=q.raw_question_id").
		Where("assessment_id = (?)", assessmentId).Where("user_id = (?)", userId).
		Group("rq.skill, is_correct").
		Select("distinct rq.skill as label, is_correct, count(*)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SkillAnalysisQuery query: %v, for assessmentId: %v, userId: %v",
			err, assessmentId, userId)
	}
	return result
}

func (tM *TllmsManager) SubjectWiseScoreQuery(assessmentId int, userId int64) []contract.SubjectAttemptsCount {
	var result []contract.SubjectAttemptsCount
	if err := tM.db.Table("question_attempts").
		Joins("INNER JOIN categories c ON c.id = question_attempts.category_id").
		Where("assessment_id = (?)", assessmentId).Where("user_id = (?)", userId).
		Group("c.subject, is_correct").
		Select("c.subject, qa.is_correct, count(qa.id)").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in SubjectWiseScoreQuery query: %v, for assessmentId: %v, userId: %v",
			err, assessmentId, userId)
	}
	return result
}

func (tM *TllmsManager) MultiUserAverageScoreQuery(assessmentId int, userIds map[int][]int64) []contract.MultiUserAverageScore {
	var result []contract.MultiUserAverageScore
	if err := tM.db.Table("assignments").
		Where("assessment_id = (?)", assessmentId).Where("assignee_id in (?)", userIds).
		Select("avg(percentage_score) as average").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in MultiUserAverageScoreQuery query: %v, for assessmentId: %v, userIds: %+v",
			err, assessmentId, userIds)
	}
	return result
}

func (tM *TllmsManager) CityUserListQuery(userId int64, maxUserList []int64) []int64 {
	var result []int64
	if err := tM.db.Table("logistics_user_locations").
		Joins("INNER JOIN logistics_user_locations l2 ON logistics_user_locations.city = l2.city AND l2.user_id = (?)", userId).
		Where("user_id in (?)", maxUserList).
		Select("user_id").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in CityUserListQuery query: %v, for userId: %v, maxUserList: %+v",
			err, userId, maxUserList)
	}
	return result
}

func (tM *TllmsManager) StateUserListQuery(userId int64, maxUserList []int64) []int64 {
	var result []int64
	if err := tM.db.Table("logistics_user_locations").
		Joins("INNER JOIN logistics_user_locations l2 ON logistics_user_locations.state = l2.state AND l2.user_id = (?)", userId).
		Where("user_id in (?)", maxUserList).
		Select("user_id").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in StateUserListQuery query: %v, for userId: %v, maxUserList: %+v",
			err, userId, maxUserList)
	}
	return result
}

func (tM *TllmsManager) UserCityQuery(userId int64) []contract.UserCity {
	var result []contract.UserCity
	if err := tM.db.Table("logistics_user_locations").
		Where("user_id = (?)", userId).
		Select("city").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in UserCityQuery query: %v, for userId: %v", err, userId)
	}
	return result
}

func (tM *TllmsManager) RankQuery(assessmentId int, userId int64, userList []int64) []contract.UserRank {
	var result []contract.UserRank
	if err := tM.db.Exec("SELECT rank_number FROM ( "+
		"SELECT assignee_id, Rank() over (ORDER BY percentage_score DESC) AS rank_number "+
		"FROM assignments WHERE assessment_id = (?) AND assignee_id IN (?)) AS rank_view "+
		"WHERE assignee_id = (?)", assessmentId, userList, userId).Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in RankQuery query: %v, for assessmentId: %v, userList: %v, userId: %v",
			err, assessmentId, userList, userId)
	}
	return result
}

func (tM *TllmsManager) AssessmentsAttended(userId int64, assessmentList []int, fromDate, toDate string) []contract.AttendedAssessment {
	var result []contract.AttendedAssessment
	if err := tM.db.Table("assignments").
		Where("assignee_id = (?) AND assessment_id in (?) AND status = 'Graded' AND completed_at BETWEEN (?) AND (?) ",
			userId, assessmentList, fromDate, toDate).
		Select("assessment_id").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in AssessmentsAttended query: %v, for userId: %v, assessmentList: %v, "+
			"fromDate: %v, toDate: %v", err, userId, assessmentList, fromDate, toDate)
	}
	return result
}

func (tM *TllmsManager) PostAssessmentAttempted(userId int64, assessmentList []int, fromDate, toDate string) []contract.AttemptedPostAssessment {
	var result []contract.AttemptedPostAssessment
	if err := tM.db.Table("assignments").
		Where("assignee_id = (?) AND assessment_id in (?) AND status = 'Graded' AND completed_at BETWEEN (?) AND (?) ",
			userId, assessmentList, fromDate, toDate).
		Select("assessment_id").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in PostAssessmentAttempted query: %v, for userId: %v, assessmentList: %v, "+
			"fromDate: %v, toDate: %v", err, userId, assessmentList, fromDate, toDate)
	}
	return result
}

func (tM *TllmsManager) UnattemptedAssessmentQuestions(assessmentIds []int) int64 {
	var result int64
	if err := tM.db.Table("assessment_questions").
		Where("assessment_id in (?)", assessmentIds).Count(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in UnattemptedAssessmentQuestions query: %v, for assessment_ids: %v",
			err, assessmentIds)
	}
	return result
}

func (tM *TllmsManager) GetAssessmentIdsExcludingSubjectiveAssessment(totalAssessmentIds []int) []int {
	var result []int
	if err := tM.db.Table("assessments").
		Where("id in (?) AND type != 'SubjectiveAssessment'", totalAssessmentIds).
		Select("id").
		Scan(&result).Error; err != nil {
		logger.Log.Sugar().Errorf("Error in GetAssessmentIdsExcludingSubjectiveAssessment query: %v, "+
			"for assessmentIds: %v", err, totalAssessmentIds)
	}
	return result
}
