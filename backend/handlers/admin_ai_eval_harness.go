package handlers

import (
	"strings"
)

type aiEvalFixture struct {
	ID               string
	Input            string
	ExpectedKeywords []string
	MinCoverage      float64
}

type aiEvalCaseResult struct {
	FixtureID        string
	InputExcerpt     string
	ExpectedKeywords []string
	SummaryExcerpt   string
	KeywordCoverage  float64
	Passed           bool
}

var aiEvalFixtures = []aiEvalFixture{
	{
		ID:               "cbc-anemia-risk",
		Input:            "CBC report shows hemoglobin 9.1 g/dL with low MCV, ferritin 8 ng/mL and mild fatigue symptoms.",
		ExpectedKeywords: []string{"hemoglobin", "low", "ferritin", "fatigue"},
		MinCoverage:      0.5,
	},
	{
		ID:               "lipid-followup",
		Input:            "Lipid profile indicates LDL 170 mg/dL, triglycerides 220 mg/dL. Recommend diet change and follow-up in 6 weeks.",
		ExpectedKeywords: []string{"ldl", "triglycerides", "diet", "follow-up"},
		MinCoverage:      0.5,
	},
	{
		ID:               "bp-monitoring",
		Input:            "Home blood pressure logs average 152/96 across 10 days. Consider medication adjustment and sodium reduction.",
		ExpectedKeywords: []string{"blood pressure", "average", "medication", "sodium"},
		MinCoverage:      0.5,
	},
}

func evaluateSummaryAgainstFixture(summary string, fixture aiEvalFixture) aiEvalCaseResult {
	lowerSummary := strings.ToLower(summary)
	matched := 0
	for _, kw := range fixture.ExpectedKeywords {
		if strings.Contains(lowerSummary, strings.ToLower(kw)) {
			matched++
		}
	}
	coverage := 0.0
	if len(fixture.ExpectedKeywords) > 0 {
		coverage = float64(matched) / float64(len(fixture.ExpectedKeywords))
	}
	inputExcerpt := fixture.Input
	if len(inputExcerpt) > 180 {
		inputExcerpt = inputExcerpt[:180] + "..."
	}
	summaryExcerpt := summary
	if len(summaryExcerpt) > 220 {
		summaryExcerpt = summaryExcerpt[:220] + "..."
	}
	return aiEvalCaseResult{
		FixtureID:        fixture.ID,
		InputExcerpt:     inputExcerpt,
		ExpectedKeywords: fixture.ExpectedKeywords,
		SummaryExcerpt:   summaryExcerpt,
		KeywordCoverage:  coverage,
		Passed:           coverage >= fixture.MinCoverage,
	}
}
