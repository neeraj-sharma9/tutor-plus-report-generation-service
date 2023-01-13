package callout

import (
	"encoding/json"
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/utility"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

var callout map[string]interface{}
var initCallOut sync.Once

func InitiateMPRConfigs() {
	GetCallOutMap()
}

func LoadCallOutMap() (data map[string]interface{}, err error) {
	calloutBytes, err := utility.GetFileContent("/internal/service/mpr/callout/callout.json")
	if err != nil {
		err = fmt.Errorf("failed to fetch callout configuration : %s", err)
		return
	}
	err = json.Unmarshal(calloutBytes, &data)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal callout configuration : %s", err)
		return
	}
	return
}

func GetCallOutMap() map[string]interface{} {
	initCallOut.Do(func() {
		var err error
		callout, err = LoadCallOutMap()
		if err != nil {
			logger.Log.Sugar().Errorf("Error in GetCalloutMap(): %v", err)
		}
	})
	return callout
}

func GetChapterCoveredCallout(chaptersCompleted int) string {
	key := GetChapterCoverageKey(chaptersCompleted)
	coverageMap := GetCallOutMap()[SummaryOfLearning].(map[string]interface{})[ChaptersCovered].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetClassAttendanceCallout(percentage float64) string {
	key := GetClassAttendanceKey(percentage)
	coverageMap := GetCallOutMap()[SummaryOfLearning].(map[string]interface{})[SummaryOfClassAttendance].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetInClassCallout(attendance float64, classQuiz float64) string {
	key := GetAttendanceAndInClassKey(attendance, classQuiz)
	if key == "" {
		return ""
	}
	coverageMap := GetCallOutMap()[SubjectWisePerformance].(map[string]interface{})[AttendanceAndInClass].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetSubjectWisePostClassCallout(proficiency int, coverage float64) string {
	key := GetPostClassCalloutKey(float64(proficiency), coverage)
	if key == "" {
		return ""
	}
	coverageMap := GetCallOutMap()[SubjectWisePerformance].(map[string]interface{})[PostClassCallout].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetSummaryPostClassCallout(proficiency float64, coverage float64) string {
	key := GetPostClassCalloutKey(proficiency, coverage)
	if key == "" {
		return ""
	}
	coverageMap := GetCallOutMap()[SummaryOfLearning].(map[string]interface{})[PostClassCallout].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetSubjectWiseCumulativeCallout() string {
	coverageMap := GetCallOutMap()[SubjectWisePerformance].(map[string]interface{})[CumulativeCallout].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetDifficultyAnalysisCallout(dg []helper.DifficultyGraph) string {
	var easy, medium, hard = 0.0, 0.0, 0.0
	for _, difficulty := range dg {
		totalQues := difficulty.Incorrect + difficulty.Correct + difficulty.NotAttempted
		if totalQues == 0 {
			break
		}
		answerPercent := (float64(difficulty.Correct) * 100.0) / float64(totalQues)
		if strings.ToUpper(difficulty.Label) == "EASY" {
			easy = answerPercent
		} else if strings.ToUpper(difficulty.Label) == "MEDIUM" {
			medium = answerPercent
		} else if strings.ToUpper(difficulty.Label) == "HARD" {
			hard = answerPercent
		}
	}
	key := GetDifficultyAnalysisCalloutKey(easy, medium, hard)
	coverageMap := GetCallOutMap()[MonthlyTest].(map[string]interface{})[DifficultyAnalysis].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetSkillAnalysisCallout(skillList []helper.SkillAnalysisGraph) string {
	var conceptual, memory, application, analyze = 0.0, 0.0, 0.0, 0.0
	for _, skill := range skillList {
		totalQues := skill.Incorrect + skill.Correct + skill.NotAttempted
		if totalQues == 0 {
			break
		}
		answerPercent := (float64(skill.Correct) * 100.0) / float64(totalQues)
		if strings.ToUpper(skill.Label) == "CONCEPTUAL" {
			conceptual = answerPercent
		} else if strings.ToUpper(skill.Label) == "MEMORY" {
			memory = answerPercent
		} else if strings.ToUpper(skill.Label) == "ANALYSIS" {
			analyze = answerPercent
		} else if strings.ToUpper(skill.Label) == "APPLICATION BASED" {
			application = answerPercent
		}
	}
	key := GetSkillAnalysisCalloutKey(conceptual, memory, application, analyze)
	coverageMap := GetCallOutMap()[MonthlyTest].(map[string]interface{})[SkillAnalysis].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetSubjectChapterCallout(countGt80 int, countLt40 int, totalSubs int) string {
	key := GetSubjectChapterCalloutKey(countGt80, countLt40, totalSubs)
	coverageMap := GetCallOutMap()[MonthlyTest].(map[string]interface{})[SubjectChapter].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	return coverageMap[random].(string)
}

func GetMonthlyTestPerformanceCallout(national int, state int, city int, class int, userScore float64, classAvg float64, attended bool) string {
	key := GetMonthlyTestPerformanceCalloutKey(national, state, city, class, userScore, classAvg, attended)
	coverageMap := GetCallOutMap()[MonthlyTest].(map[string]interface{})[MonthlyTestPerformance].(map[string]interface{})[key].([]interface{})
	random := rand.Intn(len(coverageMap))
	performanceCallout := coverageMap[random].(string)
	rank := ""
	if key == "0" { // <Rank> placeholder is only in callouts case-1(AIR) and case-2(State)
		rank = strconv.Itoa(national)
	} else if key == "1" {
		rank = strconv.Itoa(state)
	}
	performanceCallout = strings.Replace(performanceCallout, "<Rank>", rank, -1)
	return performanceCallout
}
