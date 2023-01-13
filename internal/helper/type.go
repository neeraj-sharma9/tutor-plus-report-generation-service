package helper

import (
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/contract"
)

type LookingAhead struct {
	FromDate       string           `json:"from_date"`
	ToDate         string           `json:"to_date"`
	SessionsCount  []SessionsCount  `json:"sessions_count"`
	SessionsDetail []SessionsDetail `json:"sessions_detail"`
}

type SessionsCount struct {
	Subject string `json:"subject"`
	Count   int    `json:"count"`
}

type SessionsDetail struct {
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Day     string `json:"day"`
	Topic   string `json:"topic"`
}

type UserInfo struct {
	Name             string `json:"name"`
	PremiumAccountID string `json:"premium_account_id"`
	UserID           string `json:"user_id"`
	Grade            int    `json:"grade"`
}

type MPRReq struct {
	JobId               uuid.UUID
	UserId              int64
	EpchFrmDate         int64
	EpchToDate          int64
	ReqStatus           bool
	ErrorMsg            string
	UserDetailsResponse contract.UserDetailsResponse
	State               string
}

type ClassesModel struct {
	Attended       bool          `json:"attended"`
	ClassID        string        `json:"class_id"`
	RawTopicID     int           `json:"raw_topic_id,omitempty"`
	PreRequisites  Requisites    `json:"pre_requisites"`
	PostRequisites Requisites    `json:"post_requisites"`
	TestRequisites TestRequisite `json:"test_requisite"`
	PollQuiz       PollQuiz      `json:"poll_question"`
	TopicName      string        `json:"topic_name"`
}

type Requisites struct {
	K12Videos   []int         `json:"k12_videos"`
	Assessments []int         `json:"assessments"`
	Journeys    []int         `json:"journeys"`
	Practice    []interface{} `json:"practice"`
}

type TestRequisite struct {
	Assessment  int    `json:"assessment"`
	TestType    string `json:"test_type"`
	SessionDate string `json:"session_date"`
}

type PollQuiz struct {
	Correct        int `json:"correct_answers"`
	Incorrect      int `json:"incorrect_answers"`
	TotalAttempted int `json:"total_attempted"`
	TotalQuestion  int `json:"total_questions"`
	ClassAverage   int `json:"class_average"`
}

type SkillAnalysisGraph struct {
	Label        string `json:"label"`
	Correct      int    `json:"correct"`
	Incorrect    int    `json:"incorrect"`
	NotAttempted int    `json:"not_attempted"`
}

type DifficultyGraph struct {
	Label        string `json:"label"`
	Correct      int    `json:"correct"`
	Incorrect    int    `json:"incorrect"`
	NotAttempted int    `json:"not_attempted"`
}
