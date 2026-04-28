package handlers

// KnowledgeHealthScore returns 100 when freshly reviewed and decreases linearly toward 0 as
// days_since_review approaches review_interval_days (phase 6 staleness model).
func KnowledgeHealthScore(daysSinceReview int, reviewIntervalDays int) int {
	if reviewIntervalDays <= 0 {
		reviewIntervalDays = 180
	}
	if daysSinceReview <= 0 {
		return 100
	}
	score := 100 - (daysSinceReview*100)/reviewIntervalDays
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// KnowledgeIsStale is true when days_since_review exceeds the configured interval.
func KnowledgeIsStale(daysSinceReview int, reviewIntervalDays int) bool {
	if reviewIntervalDays <= 0 {
		reviewIntervalDays = 180
	}
	return daysSinceReview > reviewIntervalDays
}
