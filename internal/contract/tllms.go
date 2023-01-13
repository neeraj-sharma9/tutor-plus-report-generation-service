package contract

import "database/sql"

type AssignmentsCompletionCursor struct {
	Completed sql.NullInt32 `db:"completed"`
	Attempted sql.NullInt32 `db:"attempted"`
}

type AssignmentScoreCursor struct {
	PercentageScore sql.NullFloat64 `db:"percentage_score"`
	Completed       sql.NullInt32   `db:"completed"`
}

type AssignmentQuestionsCursor struct {
	Correct sql.NullString `db:"is_correct"`
	Count   int            `db:"count"`
}

type PostAssignmentTillDate struct {
	Status sql.NullString `db:"status"`
	Count  int            `db:"count"`
}

type PreAssignmentTillDate struct {
	Completed sql.NullString `db:"is_completed"`
	Count     int            `db:"count"`
}

type ChapterAssessmentQuestions struct {
	Chapter string `db:"name"`
	Subject string `db:"subject"`
	Count   int    `db:"count"`
}

type AssessmentTime struct {
	TotalAllowedTime sql.NullInt32 `db:"time_allowed"`
}

type AssessmentAttemptInfo struct {
	PercentageScore sql.NullFloat64 `db:"percentage_score"`
	ExamDate        sql.NullString  `db:"exam_date"`
	TimeTaken       sql.NullInt32   `db:"time_taken"`
}

type QuestionsAttempted struct {
	IsCorrect sql.NullString `db:"is_correct"`
	Count     int            `db:"count"`
}

type ChapterWisePerformanceCursor struct {
	Chapter   string         `db:"name"`
	Subject   string         `db:"subject"`
	IsCorrect sql.NullString `db:"is_correct"`
	Count     int            `db:"count"`
}

type ChapterWiseRankCursor struct {
	UserId  sql.NullInt64 `db:"user_id"`
	AheadOf float64       `db:"percent_rank"`
}

type DifficultySkillCursor struct {
	Label     sql.NullString `db:"label"`
	IsCorrect sql.NullString `db:"is_correct"`
	Count     int            `db:"count"`
}

type SkillAnalysisGraph struct {
	Label        string `json:"label"`
	Correct      int    `json:"correct"`
	Incorrect    int    `json:"incorrect"`
	NotAttempted int    `json:"not_attempted"`
}

type SubjectAttemptsCount struct {
	Subject   sql.NullString `db:"subject"`
	IsCorrect sql.NullString `db:"is_correct"`
	Count     int            `db:"count"`
}

type MultiUserAverageScore struct {
	Average sql.NullFloat64 `db:"average"`
}

type UserRank struct {
	Rank sql.NullInt64 `db:"rank_number"`
}

type UserCity struct {
	City sql.NullString `db:"city"`
}

type AttendedAssessment struct {
	AssessmentID sql.NullInt32 `db:"assessment_id"`
}

type AttemptedPostAssessment struct {
	AssessmentID sql.NullInt32 `db:"assessment_id"`
	//Attempted	 	sql.NullBool 	`db:"attempted"`
}

type AssignmentAttemptedCursor struct {
	Count int `db:"count"`
}
