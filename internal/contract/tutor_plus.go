package contract

import "github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"

type UserDetailsResponse struct {
	Data `json:"data"`
}

type Data struct {
	UserInfo              helper.UserInfo     `json:"user_info"`
	ReportMonth           DateRange           `json:"report_month"`
	Subjects              []SubjectModel      `json:"subjects"`
	NextMonthlyExamDate   string              `json:"next_monthly_exam_date,omitempty"`
	MonthlyExamClassmates map[int][]int64     `json:"monthly_exam_classmates,omitempty"`
	LookingAhead          helper.LookingAhead `json:"looking_ahead"`
	SameBatchUserList     []int64             `json:"users_of_same_batch"`
	OverAllPerformance    OverAllPerformance  `json:"overall_performance"`
}

type DateRange struct {
	FromDate      string `json:"from_date"`
	ToDate        string `json:"to_date"`
	FromDateEpoch int64  `json:"from_date_epoch"`
	ToDateEpoch   int64  `json:"to_date_epoch"`
}

type SubjectModel struct {
	Subject                  string                 `json:"subject"`
	Type                     string                 `json:"type"`
	CallOut                  string                 `json:"call_out"`
	TotalClasses             int                    `json:"total_classes"`
	OnTime                   int                    `json:"on_time"`
	LateDays                 int                    `json:"late_days"`
	TotalChapter             int                    `json:"total_chapter"`
	TotalAttended            int                    `json:"total_attended"`
	MonthlyTestAssignmentIds []int                  `json:"monthly_test_assignment_ids,omitempty"`
	ChapterCovered           []ChapterCovered       `json:"chapter_covered"`
	Classes                  []helper.ClassesModel  `json:"classes,omitempty"`
	PerformanceTillDate      NeoPerformanceTillDate `json:"performance_till_date,omitempty"`
	PollQuestionNationalPer  int                    `json:"poll_question_national_percentage"`
}

type OverAllPerformance struct {
	ByjusClassInfo ByjusClassInfo `json:"byjus_class_info"`
}

type ChapterCovered struct {
	ChapterName     string          `json:"chapter_name"`
	ConceptTaught   []string        `json:"concept_taught"`
	ClassesAttended ClassesAttended `json:"classes_attended"`
}

type NeoPerformanceTillDate struct {
	PreRequisite    []int                  `json:"all_pre_requisites_ids,omitempty"`
	PostRequisite   []int                  `json:"all_post_requisite_ids,omitempty"`
	ClassAttendance helper.ClassAttendance `json:"class_attendance,omitempty"`
	PollQuestion    helper.PollQuestion    `json:"poll_question,omitempty"`
}

type ByjusClassInfo struct {
	TotalClasses  int `json:"total_classes"`
	TotalAttended int `json:"total_attended"`
	TotalChapter  int `json:"total_chapter"`
	TotalOnTime   int `json:"total_on_time"`
}

type ClassesAttended struct {
	TotalClasses  int   `json:"total_classes"`
	TotalAttended int   `json:"total_attended"`
	Sessions      []int `json:"sessions"`
}
