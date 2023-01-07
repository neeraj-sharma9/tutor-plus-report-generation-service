package helper

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

type ClassAttendance struct {
	TotalClasses int `json:"total_classes,omitempty"`
	OnTime       int `json:"on_time"`
	LateDays     int `json:"late_days"`
	Missed       int `json:"missed"`
}

type PollQuestion struct {
	Correct     int `json:"correct"`
	Incorrect   int `json:"incorrect"`
	Unattempted int `json:"unattempted"`
	Attempted   int `json:"total_attempted"`
	Questions   int `json:"total_questions"`
}
