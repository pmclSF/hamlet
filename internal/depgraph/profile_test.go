package depgraph

import (
	"testing"
)

func TestAnalyzeProfile_SmallGraph(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	cov := AnalyzeCoverage(g)
	dupes := DetectDuplicates(g)
	fanout := AnalyzeFanout(g, DefaultFanoutThreshold)
	insights := ProfileInsights{
		Coverage:   &cov,
		Duplicates: &dupes,
		Fanout:     &fanout,
	}

	profile := AnalyzeProfile(g, insights)

	// 4 tests → tiny.
	if profile.TestVolume != "tiny" {
		t.Errorf("expected tiny volume, got %s", profile.TestVolume)
	}
	if profile.CIPressure != "low" {
		t.Errorf("expected low CI pressure, got %s", profile.CIPressure)
	}
}

func TestAnalyzeProfile_NilInsights(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := AnalyzeProfile(g, ProfileInsights{})

	if profile.CoverageConfidence != "low" {
		t.Errorf("expected low coverage confidence with nil insights, got %s", profile.CoverageConfidence)
	}
	if profile.RedundancyLevel != "low" {
		t.Errorf("expected low redundancy with nil insights, got %s", profile.RedundancyLevel)
	}
}

func TestDetectEdgeCases_FewTests(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	profile := RepoProfile{TestVolume: "tiny", CIPressure: "low"}
	cases := DetectEdgeCases(profile, g, ProfileInsights{})

	types := map[EdgeCaseType]bool{}
	for _, c := range cases {
		types[c.Type] = true
	}
	if !types[EdgeCaseFewTests] {
		t.Error("expected FEW_TESTS edge case")
	}
	if !types[EdgeCaseFastCIAlready] {
		t.Error("expected FAST_CI_ALREADY edge case")
	}
}

func TestApplyEdgeCasePolicy_FewTests(t *testing.T) {
	t.Parallel()
	cases := []EdgeCase{
		{Type: EdgeCaseFewTests, Severity: "critical"},
	}
	profile := RepoProfile{TestVolume: "tiny"}

	policy := ApplyEdgeCasePolicy(cases, profile)

	if !policy.OptimizationDisabled {
		t.Error("expected optimization disabled for FEW_TESTS")
	}
	if policy.FallbackLevel != FallbackFullSuite {
		t.Errorf("expected FullSuite fallback, got %s", policy.FallbackLevel)
	}
	if policy.ConfidenceAdjustment >= 1.0 {
		t.Errorf("expected confidence adjustment < 1.0, got %f", policy.ConfidenceAdjustment)
	}
	if len(policy.Recommendations) == 0 {
		t.Error("expected at least one recommendation")
	}
}

func TestApplyEdgeCasePolicy_NoEdgeCases(t *testing.T) {
	t.Parallel()
	profile := RepoProfile{TestVolume: "large"}

	policy := ApplyEdgeCasePolicy(nil, profile)

	if policy.OptimizationDisabled {
		t.Error("optimization should not be disabled with no edge cases")
	}
	if policy.ConfidenceAdjustment != 1.0 {
		t.Errorf("expected confidence 1.0 with no edge cases, got %f", policy.ConfidenceAdjustment)
	}
	if policy.FallbackLevel != FallbackDirectDeps {
		t.Errorf("expected DirectDeps fallback, got %s", policy.FallbackLevel)
	}
}
