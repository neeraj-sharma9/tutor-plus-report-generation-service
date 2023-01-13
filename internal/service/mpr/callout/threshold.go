package callout

func GetChapterCoverageKey(value int) string {
	if value > 75 {
		return "C1"
	} else if value > 50 {
		return "C2"
	} else if value > 25 {
		return "C3"
	} else if value > 0 {
		return "C4"
	} else {
		return "C5"
	}
}

func GetClassAttendanceKey(value float64) string {
	if value >= 70 {
		return "A1"
	} else if value >= 40 {
		return "A2"
	} else if value > 0 {
		return "A3"
	} else {
		return "A4"
	}
}

func GetAttendanceAndInClassKey(attendance float64, classQuiz float64) string {
	if attendance >= 70 && classQuiz > 75 {
		return "B1"
	} else if attendance >= 70 && classQuiz > 50 {
		return "B2"
	} else if attendance >= 70 && classQuiz > 25 {
		return "B3"
	} else if attendance >= 70 && classQuiz > 0 {
		return "B4"
	} else if attendance >= 70 && classQuiz == 0 {
		return "B5"
	} else if attendance >= 40 && classQuiz > 75 {
		return "B6"
	} else if attendance >= 40 && classQuiz > 50 {
		return "B7"
	} else if attendance >= 40 && classQuiz > 0 {
		return "B8"
	} else if attendance >= 40 && classQuiz == 0 {
		return "B9"
	} else if attendance >= 0 && classQuiz > 50 {
		return "B10"
	} else if attendance >= 0 && classQuiz > 0 {
		return "B11"
	} else if attendance > 0 && classQuiz == 0 {
		return "B12"
	} else if attendance == 0 && classQuiz == 0 {
		return "B13"
	} else {
		return ""
	}
}

func GetPostClassCalloutKey(proficiency float64, coverage float64) string {
	if proficiency == 0 || coverage == 0 {
		return "0" // P5=0 or C5=0
	} else if proficiency < 70 && coverage <= 50 {
		return "1" // P3 - P4 - C3 -C4
	} else if proficiency < 70 && coverage > 50 {
		return "2" // P3 - P4 - C1- C2
	} else if proficiency > 70 && proficiency < 81 && coverage <= 50 {
		return "3" // P2 - C3- C4
	} else if proficiency > 70 && proficiency < 81 && coverage > 50 && coverage <= 75 {
		return "4" // P2- C2
	} else if proficiency >= 70 && proficiency < 81 && coverage > 75 && coverage <= 100 {
		return "5" // P2- C1
	} else if proficiency >= 81 && coverage <= 25 {
		return "6" // P1- C4
	} else if proficiency >= 81 && coverage > 25 && coverage <= 50 {
		return "7" // P1- C3
	} else if proficiency >= 81 && coverage > 50 && coverage <= 75 {
		return "8" // P1- C2
	} else if proficiency >= 81 && coverage > 75 {
		return "9" // P1- C1
	} else {
		return ""
	}
}

func GetDifficultyAnalysisCalloutKey(easy float64, medium float64, hard float64) (key string) {
	key = "0"
	if easy >= 75 && medium >= 75 && hard >= 75 {
		key = "1"
	} else if easy >= 50 {
		if medium >= 50 {
			if easy < 75 && medium < 75 && hard >= 50 && hard < 75 {
				key = "2"
			} else if hard < 50 {
				key = "3"
			}
		} else if hard < 50 {
			key = "4"
		}
	} else if medium < 50 && hard < 50 {
		key = "5"
	}
	return
}

func GetSkillAnalysisCalloutKey(conceptual float64, memory float64, application float64, analyze float64) (key string) {
	key = "0"
	if conceptual >= 75 && memory >= 75 && application >= 75 && analyze >= 75 {
		key = "1"
	} else if conceptual >= 50 && memory >= 50 && conceptual < 75 && memory < 75 {
		if application >= 50 && analyze >= 50 && application < 75 && analyze < 75 {
			key = "2"
		} else if application < 50 && analyze < 50 {
			key = "3"
		}
	} else if conceptual < 50 && memory < 50 && application < 50 && analyze < 50 {
		key = "4"
	}
	return
}

func GetSubjectChapterCalloutKey(countGt80 int, countLt40 int, totalSubs int) (key string) {
	key = "0"
	if countLt40 == totalSubs && countGt80 == 0 {
		key = "1"
	} else if countGt80 == totalSubs && countLt40 == 0 {
		key = "2"
	}
	return
}

func GetMonthlyTestPerformanceCalloutKey(national int, state int, city int, class int, userScore float64, classAvg float64, attended bool) (key string) {
	key = "7"
	if attended {
		if national > 0 && national <= 100 {
			key = "0"
		} else if state > 0 && state <= 50 {
			key = "1"
		} else if city > 0 && city <= 25 {
			key = "2"
		} else if class > 0 && class <= 5 {
			key = "4"
		} else {
			if userScore > classAvg {
				key = "4"
			} else if userScore == classAvg {
				key = "5"
			} else {
				key = "6"
			}
		}
	}
	return
}
