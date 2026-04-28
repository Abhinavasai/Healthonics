package handlers

import "testing"

func TestEvaluateSummaryAgainstFixture(t *testing.T) {
	fixture := aiEvalFixture{
		ID:               "f1",
		Input:            "input text",
		ExpectedKeywords: []string{"ldl", "diet", "follow-up"},
		MinCoverage:      0.5,
	}
	result := evaluateSummaryAgainstFixture("LDL elevated. Diet change advised.", fixture)
	if result.FixtureID != "f1" {
		t.Fatalf("unexpected fixture id: %s", result.FixtureID)
	}
	if result.KeywordCoverage <= 0 {
		t.Fatalf("expected positive coverage, got %f", result.KeywordCoverage)
	}
	if !result.Passed {
		t.Fatalf("expected fixture to pass")
	}
}
