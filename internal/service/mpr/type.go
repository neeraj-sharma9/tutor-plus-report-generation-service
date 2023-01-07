package mpr

import (
	"github.com/google/uuid"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/contract"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/manager"
)

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

type ProgressData struct {
	Cover DateRange `json:"cover"`
	//Leap                 	Leap                    	`json:"leap"`
	MonthlyTestPerformance []MonthlyTestPerformance `json:"monthly_test_performance"`
	SummaryOfLearning      SummaryOfLearning        `json:"summary_of_learning"`
	SubjectWisePerformance []SubjectWisePerformance `json:"subject_wise_performance"`
	//FocusArea             FocusArea                	`json:"focus_area"`
	LookingAhead          helper.LookingAhead `json:"looking_ahead"`
	UserInfo              helper.UserInfo     `json:"user_info"`
	*manager.TllmsManager `json:"-"`
}

type MonthlyTestPerformance struct {
	Attended              bool                  `json:"attended"`
	AssessmentId          int                   `json:"-"`
	ChapterTestedDetails  ChapterTestedDetails  `json:"chapter_tested_details"`
	SubjectWiseScore      []SubjectWiseScore    `json:"subject_wise_score"`
	ExamDate              string                `json:"exam_date"`
	ChapterTested         int                   `json:"chapter_tested"`
	TotalExamTime         int                   `json:"total_exam_time"`
	TimeTaken             int                   `json:"time_taken"`
	NextExamDate          string                `json:"next_exam_date"`
	MonthlyTestCallout    string                `json:"monthly_test_callout"`
	TotalQuestions        int                   `json:"total_questions"`
	QuestionAttempted     int                   `json:"question_attempted"`
	CorrectAnswer         int                   `json:"correct_answer"`
	PercentageScores      PercentageScores      `json:"percentage_scores"`
	Ranks                 Ranks                 `json:"ranks"`
	SubjectChapterCallout string                `json:"subject_chapter_callout"`
	ChapterWiseAnalysis   []ChapterWiseAnalysis `json:"chapter_wise_analysis"`
	DifficultyAnalysis    DifficultyAnalysis    `json:"difficulty_analysis"`
	SkillAnalysis         SkillAnalysis         `json:"skill_analysis"`
	City                  string                `json:"-"`
}

type SummaryOfLearning struct {
	ClassAttendance ClassAttendanceForSummary `json:"class_attendance"`
	ChaptersCovered ChaptersCovered           `json:"chapters_covered"`
	Assignments     Assignments               `json:"assignments"`
	LearnerTags     LearnerTags               `json:"learner_tag"`
}

type SubjectWisePerformance struct {
	Subject              string                 `json:"subject"`
	ClassAttendance      helper.ClassAttendance `json:"class_attendance"`
	ClassQuiz            ClassQuiz              `json:"class_quiz"`
	InClassCallout       string                 `json:"in_class_callout"`
	PostClassCallout     string                 `json:"post_class_callout"`
	Assignments          PostAssignments        `json:"assignments"`
	SessionWiseBreakdown []SessionWiseBreakdown `json:"session_wise_breakdown"`
	PerformanceTillDate  PerformanceTillDate    `json:"performance_till_date"`
	CumulativeCallout    string                 `json:"cumulative_callout"`
}

type ChapterTestedDetails map[string][]string
type SubjectWiseScore map[string]string

type PercentageScores struct {
	User     float64 `json:"user"`
	National float64 `json:"national"`
	Class    float64 `json:"class"`
	City     float64 `json:"city"`
	State    float64 `json:"state"`
}

type Ranks struct {
	National int `json:"national"`
	State    int `json:"state"`
	City     int `json:"city"`
	Class    int `json:"class"`
}

type ChapterWiseAnalysis struct {
	Subject  string     `json:"subject"`
	Chapters []Chapters `json:"chapters"`
}

type DifficultyAnalysis struct {
	CallOut         string                   `json:"call_out"`
	DifficultyGraph []helper.DifficultyGraph `json:"difficulty_graph"`
}

type SkillAnalysis struct {
	CallOut            string                      `json:"call_out"`
	SkillAnalysisGraph []helper.SkillAnalysisGraph `json:"skill_analysis_graph"`
}

type ClassAttendanceForSummary struct {
	TotalClasses int    `json:"total_classes"`
	OnTime       int    `json:"on_time"`
	LateDays     int    `json:"late_days"`
	Missed       int    `json:"missed"`
	Callout      string `json:"callout"`
}

type ChaptersCovered struct {
	TotalChapters int    `json:"total_chapters"`
	Completed     int    `json:"completed"`
	Missed        int    `json:"missed"`
	Callout       string `json:"callout"`
}

type Assignments struct {
	TotalAssignments int     `json:"total_assignments"`
	Completed        int     `json:"completed"`
	CompletedPerc    float64 `json:"completed_percentage"`
	Missed           int     `json:"missed"`
	Score            float64 `json:"score"`
	Callout          string  `json:"callout"`
}

type LearnerTags struct {
	Regularity           string `json:"regularity"`
	Punctuality          string `json:"punctuality"`
	HomeWorkCompletion   string `json:"homework_completion"`
	SelfDirectedLearning string `json:"self_directed_learning"`
	ActiveInClass        string `json:"active_in_class"`
}

type ClassQuiz struct {
	Correct         int     `json:"correct"`
	Incorrect       int     `json:"incorrect"`
	NotAttempted    int     `json:"not_attempted"`
	ClassAvg        int     `json:"class_avg"`
	NationalAvg     int     `json:"national_avg"`
	TotalClassQuiz  int     `json:"total_class_quiz"`
	TotalQuestions  int     `json:"total_questions"`
	PercentageScore float64 `json:"percentage_score"`
	Attempted       int     `json:"attempted"`
}

type PostAssignments struct {
	Correct              int `json:"correct"`
	Incorrect            int `json:"incorrect"`
	NotAttempted         int `json:"not_attempted"`
	ClassAvg             int `json:"class_avg"`
	NationalAvg          int `json:"national_avg"`
	TotalAssignments     int `json:"total_assignments"`
	CompletedAssignments int `json:"-"`
	PercentageScore      int `json:"percentage_score"`
}

type SessionWiseBreakdown struct {
	Session   string    `json:"session"`
	Attended  bool      `json:"attended"`
	PreClass  PreClass  `json:"pre_class"`
	InClass   InClass   `json:"in_class"`
	PostClass PostClass `json:"post_class"`
}

type PerformanceTillDate struct {
	PreRequisite    RequisiteStatus        `json:"pre_requisite,omitempty"`
	PostRequisite   RequisiteStatus        `json:"post_requisite,omitempty"`
	ClassAttendance helper.ClassAttendance `json:"class_attendance,omitempty"`
	PollQuestion    helper.PollQuestion    `json:"poll_question,omitempty"`
}

type Chapters struct {
	Chapter            string `json:"chapter"`
	Correct            int    `json:"correct"`
	Incorrect          int    `json:"incorrect"`
	NotAttempted       int    `json:"not_attempted"`
	PerformanceCallout string `json:"performance_callout"`
	Direction          string `json:"direction"`
}

type PreClass struct {
	Total     int `json:"total"`
	Attempted int `json:"attempted"`
	Completed int `json:"-"`
}

type InClass struct {
	Total     int `json:"total"`
	Attempted int `json:"attempted"`
}

type PostClass struct {
	Total     int `json:"total"`
	Attempted int `json:"attempted"`
}

type RequisiteStatus struct {
	Total               int `json:"total"`
	Completed           int `json:"completed"`
	AttemptedPercentage int `json:"attempted_percentage"`
	Missed              int `json:"missed"`
	InProgress          int `json:"in_progress"`
}

type DateRange struct {
	FromDate      string `json:"from_date"`
	ToDate        string `json:"to_date"`
	FromDateEpoch int64  `json:"from_date_epoch"`
	ToDateEpoch   int64  `json:"to_date_epoch"`
}
